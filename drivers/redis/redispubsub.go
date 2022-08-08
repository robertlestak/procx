package redis

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type RedisPubSub struct {
	Client   *redis.Client
	Host     string
	Port     string
	Password string
	Key      string
	// TLS
	EnableTLS   *bool
	TLSInsecure *bool
	TLSCert     *string
	TLSKey      *string
	TLSCA       *string
}

func (d *RedisPubSub) LoadEnv(prefix string) error {
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
	return nil
}

func (d *RedisPubSub) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.Host = *flags.RedisHost
	d.Port = *flags.RedisPort
	d.Password = *flags.RedisPassword
	d.Key = *flags.RedisKey
	d.EnableTLS = flags.RedisEnableTLS
	d.TLSInsecure = flags.RedisTLSSkipVerify
	d.TLSCert = flags.RedisCertFile
	d.TLSKey = flags.RedisKeyFile
	d.TLSCA = flags.RedisCAFile
	return nil
}

func (d *RedisPubSub) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "Init",
	})
	cfg := &redis.Options{
		Addr:        fmt.Sprintf("%s:%s", d.Host, d.Port),
		Password:    d.Password,
		DB:          0,
		DialTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Second,
	}
	if d.EnableTLS != nil && *d.EnableTLS {
		cfg.TLSConfig = tlsConfig(*d.TLSInsecure, *d.TLSCert, *d.TLSKey, *d.TLSCA)
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

func (d *RedisPubSub) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from redis pub/sub")
	l.Debug("Receiving message from redis subscription")
	sub := d.Client.Subscribe(d.Key)
	defer sub.Close()
	for {
		msg, err := sub.ReceiveMessage()
		if err != nil {
			// If the queue is empty, return nil
			if err == redis.Nil {
				l.Debug("Queue is empty")
				return nil, nil
			}
			l.WithError(err).Error("Failed to receive message")
			continue
		}
		l.Debug("Received message")
		return strings.NewReader(msg.Payload), nil
	}
}

func (d *RedisPubSub) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from redis pub/sub")
	return nil
}

func (d *RedisPubSub) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	return nil
}

func (d *RedisPubSub) Cleanup() error {
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
