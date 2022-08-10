package redis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type StreamOp string

var (
	StreamOpAck = StreamOp("ack")
	StreamOpDel = StreamOp("del")
)

type RedisStream struct {
	Client        *redis.Client
	Host          string
	Port          string
	Password      string
	ConsumerName  *string
	ConsumerGroup *string
	Key           string
	ValueKeys     []string
	MessageID     *string
	ClearOp       *StreamOp
	FailOp        *StreamOp
	// TLS
	EnableTLS   *bool
	TLSInsecure *bool
	TLSCert     *string
	TLSKey      *string
	TLSCA       *string
}

func (d *RedisStream) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment variables")
	if os.Getenv(prefix+"REDIS_HOST") != "" {
		d.Host = os.Getenv(prefix + "REDIS_HOST")
	}
	if os.Getenv(prefix+"REDIS_PORT") != "" {
		d.Port = os.Getenv(prefix + "REDIS_PORT")
	}
	if os.Getenv(prefix+"REDIS_PASSWORD") != "" {
		d.Password = os.Getenv(prefix + "REDIS_PASSWORD")
	}
	if os.Getenv(prefix+"REDIS_KEY") != "" {
		d.Key = os.Getenv(prefix + "REDIS_KEY")
	}
	if os.Getenv(prefix+"REDIS_STREAM_CONSUMER_GROUP") != "" {
		v := os.Getenv(prefix + "REDIS_STREAM_CONSUMER_GROUP")
		d.ConsumerGroup = &v
	}
	if os.Getenv(prefix+"REDIS_STREAM_CONSUMER_NAME") != "" {
		v := os.Getenv(prefix + "REDIS_STREAM_CONSUMER_NAME")
		d.ConsumerName = &v
	}
	if os.Getenv(prefix+"REDIS_STREAM_CLEAR_OP") != "" {
		v := StreamOp(os.Getenv(prefix + "REDIS_STREAM_CLEAR_OP"))
		d.ClearOp = &v
	}
	if os.Getenv(prefix+"REDIS_STREAM_FAIL_OP") != "" {
		v := StreamOp(os.Getenv(prefix + "REDIS_STREAM_FAIL_OP"))
		d.FailOp = &v
	}
	if os.Getenv(prefix+"REDIS_ENABLE_TLS") != "" {
		v := os.Getenv(prefix+"REDIS_ENABLE_TLS") == "true"
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"REDIS_TLS_INSECURE") != "" {
		v := os.Getenv(prefix+"REDIS_TLS_INSECURE") == "true"
		d.TLSInsecure = &v
	}
	if os.Getenv(prefix+"REDIS_TLS_CERT_FILE") != "" {
		v := os.Getenv(prefix + "REDIS_TLS_CERT_FILE")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"REDIS_TLS_KEY_FILE") != "" {
		v := os.Getenv(prefix + "REDIS_TLS_KEY_FILE")
		d.TLSKey = &v
	}
	if os.Getenv(prefix+"REDIS_TLS_CA_FILE") != "" {
		v := os.Getenv(prefix + "REDIS_TLS_CA_FILE")
		d.TLSCA = &v
	}
	if os.Getenv(prefix+"REDIS_STREAM_VALUE_KEYS") != "" {
		v := strings.Split(os.Getenv(prefix+"REDIS_STREAM_VALUE_KEYS"), ",")
		d.ValueKeys = v
	}
	return nil
}

func cleanStringSlice(s []string) []string {
	var r []string
	for _, v := range s {
		if v != "" {
			r = append(r, v)
		}
	}
	return r
}

func (d *RedisStream) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.Host = *flags.RedisHost
	d.Port = *flags.RedisPort
	d.Password = *flags.RedisPassword
	d.Key = *flags.RedisKey
	d.ValueKeys = cleanStringSlice(strings.Split(*flags.RedisValueKeys, ","))
	d.EnableTLS = flags.RedisEnableTLS
	d.TLSInsecure = flags.RedisTLSSkipVerify
	d.TLSCert = flags.RedisCertFile
	d.TLSKey = flags.RedisKeyFile
	d.TLSCA = flags.RedisCAFile
	if flags.RedisConsumerGroup != nil {
		d.ConsumerGroup = flags.RedisConsumerGroup
	}
	if flags.RedisConsumerName != nil {
		d.ConsumerName = flags.RedisConsumerName
	} else {
		v := uuid.New().String()
		d.ConsumerName = &v
	}
	if flags.RedisClearOp != nil {
		v := StreamOp(*flags.RedisClearOp)
		d.ClearOp = &v
	}
	if flags.RedisFailOp != nil {
		v := StreamOp(*flags.RedisFailOp)
		d.FailOp = &v
	}
	return nil
}

func (d *RedisStream) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "Init",
	})
	l.Debug("Initializing redis stream driver")
	cfg := &redis.Options{
		Addr:        fmt.Sprintf("%s:%s", d.Host, d.Port),
		Password:    d.Password,
		DB:          0,
		DialTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Second,
	}
	if d.EnableTLS != nil && *d.EnableTLS {
		tc, err := utils.TlsConfig(d.EnableTLS, d.TLSInsecure, d.TLSCA, d.TLSCert, d.TLSKey)
		if err != nil {
			return err
		}
		cfg.TLSConfig = tc
	}
	d.Client = redis.NewClient(cfg)
	cmd := d.Client.Ping()
	if cmd.Err() != nil {
		l.Error("Failed to connect to redis")
		return cmd.Err()
	}
	l.Debug("Connected to redis")
	return nil
}

func (d *RedisStream) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from redis stream")
	var message redis.XMessage
	if d.ConsumerGroup != nil && *d.ConsumerGroup != "" {
		l.Debug("Getting work from redis stream with consumer group")
		res := d.Client.XReadGroup(&redis.XReadGroupArgs{
			Group:    *d.ConsumerGroup,
			Consumer: *d.ConsumerName,
			Streams:  []string{d.Key, ">"},
			Count:    1,
			Block:    time.Second * 0,
		})
		if res.Err() != nil {
			l.Error("Failed to get work from redis stream")
			return nil, res.Err()
		}
		if len(res.Val()) == 0 {
			l.Debug("No work found in redis stream")
			return nil, nil
		}
		l.Debug("Got work from redis stream")
		message = res.Val()[0].Messages[0]
	} else {
		l.Debug("Getting work from redis stream without consumer group")
		res := d.Client.XRead(&redis.XReadArgs{
			Streams: []string{d.Key, "$"},
			Count:   1,
			Block:   time.Second * 0,
		})
		if res.Err() != nil {
			l.Error("Failed to get work from redis stream")
			return nil, res.Err()
		}
		if len(res.Val()) == 0 {
			l.Debug("No work found in redis stream")
			return nil, nil
		}
		l.Debug("Got work from redis stream")
		message = res.Val()[0].Messages[0]
	}
	l.Debug("Got work from redis stream", message)
	d.MessageID = &message.ID
	var data any
	if len(d.ValueKeys) > 0 {
		ldata := make(map[string]interface{})
		for _, key := range d.ValueKeys {
			var ok bool
			ldata[key], ok = message.Values[key]
			if !ok {
				l.Error("Failed to get value from redis stream")
				return nil, errors.New("failed to get value from redis stream")
			}
		}
		data = ldata
	} else {
		l.Debug("Getting message from redis stream")
		data = message.Values
	}
	jd, err := json.Marshal(data)
	if err != nil {
		l.Error("Failed to marshal message")
		return nil, err
	}
	return bytes.NewReader(jd), nil
}

func (d *RedisStream) xdel() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "xdel",
	})
	l.Debug("Deleting message from redis stream")
	if d.MessageID == nil {
		l.Error("No message id found")
		return errors.New("no message id found")
	}
	res := d.Client.XDel(d.Key, *d.MessageID)
	if res.Err() != nil {
		l.Error("Failed to delete message from redis stream")
		return res.Err()
	}
	l.Debug("Deleted message from redis stream")
	return nil
}

func (d *RedisStream) ack() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "ack",
	})
	l.Debug("Acking message")
	if d.MessageID == nil {
		l.Error("No message id found")
		return errors.New("no message id found")
	}
	res := d.Client.XAck(d.Key, *d.ConsumerGroup, *d.MessageID)
	if res.Err() != nil {
		l.Error("Failed to ack message")
		return res.Err()
	}
	l.Debug("Acked message")
	return nil
}

func (d *RedisStream) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from redis stream")
	if d.ClearOp == nil {
		return nil
	}
	switch *d.ClearOp {
	case StreamOpDel:
		return d.xdel()
	case StreamOpAck:
		return d.ack()
	default:
		return nil
	}
}

func (d *RedisStream) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	if d.FailOp == nil {
		return nil
	}
	switch *d.FailOp {
	case StreamOpDel:
		return d.xdel()
	case StreamOpAck:
		return d.ack()
	default:
		return nil
	}
}

func (d *RedisStream) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up")
	if err := d.Client.Close(); err != nil {
		l.WithError(err).Error("Failed to close redis client")
		return err
	}
	return nil
}
