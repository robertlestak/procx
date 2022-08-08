package redis

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type RedisList struct {
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

func (d *RedisList) LoadEnv(prefix string) error {
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

func (d *RedisList) LoadFlags() error {
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

func tlsConfig(insecure bool, certFile string, keyFile string, caFile string) *tls.Config {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "tlsConfig",
	})
	l.Debug("Configuring TLS")
	cfg := &tls.Config{}
	if insecure {
		l.Debug("TLS is insecure")
		cfg.InsecureSkipVerify = insecure
	}
	if certFile != "" {
		l.Debug("Loading TLS certificate")
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			l.WithError(err).Error("Failed to load TLS certificate")
			return nil
		}
		cfg.Certificates = []tls.Certificate{cert}
	}
	if caFile != "" {
		l.Debug("Loading TLS CA")
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			l.WithError(err).Error("Failed to load TLS CA")
			return nil
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		cfg.RootCAs = caCertPool
	}
	l.Debug("TLS configured")
	return cfg
}

func (d *RedisList) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "Init",
	})
	l.Debug("Initializing redis list driver")
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

func (d *RedisList) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from redis list")
	msg, err := d.Client.LPop(d.Key).Result()
	if err != nil {
		// If the queue is empty, return nil
		if err == redis.Nil {
			l.Debug("Queue is empty")
			return nil, nil
		}
		l.WithError(err).Error("Failed to receive message")
		return nil, err
	}
	l.Debug("Received message")
	return strings.NewReader(msg), nil
}

func (d *RedisList) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from redis list")
	return nil
}

func (d *RedisList) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "redis",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	return nil
}

func (d *RedisList) Cleanup() error {
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
