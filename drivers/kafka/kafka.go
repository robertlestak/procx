package kafka

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/utils"
	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	log "github.com/sirupsen/logrus"
)

type SaslType string

var (
	SaslTypePlain = SaslType("plain")
	SaslTypeScram = SaslType("scram")
)

type Kafka struct {
	Client  *kafka.Reader
	Brokers []string
	Group   *string
	Topic   *string
	// TLS
	EnableTLS   *bool
	TLSInsecure *bool
	TLSCert     *string
	TLSKey      *string
	TLSCA       *string
	// SASL
	EnableSASL *bool
	SaslType   *SaslType
	Username   *string
	Password   *string
}

func (d *Kafka) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "kafka",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	if os.Getenv(prefix+"KAFKA_BROKERS") != "" {
		d.Brokers = strings.Split(os.Getenv(prefix+"KAFKA_BROKERS"), ",")
	}
	if os.Getenv(prefix+"KAFKA_GROUP") != "" {
		v := os.Getenv(prefix + "KAFKA_GROUP")
		d.Group = &v
	}
	if os.Getenv(prefix+"KAFKA_TOPIC") != "" {
		v := os.Getenv(prefix + "KAFKA_TOPIC")
		d.Topic = &v
	}
	if os.Getenv(prefix+"KAFKA_ENABLE_TLS") != "" {
		v := os.Getenv(prefix+"KAFKA_ENABLE_TLS") == "true"
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"KAFKA_TLS_INSECURE") != "" {
		v := os.Getenv(prefix+"KAFKA_TLS_INSECURE") == "true"
		d.TLSInsecure = &v
	}
	if os.Getenv(prefix+"KAFKA_TLS_CERT_FILE") != "" {
		v := os.Getenv(prefix + "KAFKA_TLS_CERT_FILE")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"KAFKA_TLS_KEY_FILE") != "" {
		v := os.Getenv(prefix + "KAFKA_TLS_KEY_FILE")
		d.TLSKey = &v
	}
	if os.Getenv(prefix+"KAFKA_TLS_CA_FILE") != "" {
		v := os.Getenv(prefix + "KAFKA_TLS_CA_FILE")
		d.TLSCA = &v
	}
	if os.Getenv(prefix+"KAFKA_ENABLE_SASL") != "" {
		v := os.Getenv(prefix+"KAFKA_ENABLE_SASL") == "true"
		d.EnableSASL = &v
	}
	if os.Getenv(prefix+"KAFKA_SASL_TYPE") != "" {
		v := SaslType(os.Getenv(prefix + "KAFKA_SASL_TYPE"))
		d.SaslType = &v
	}
	if os.Getenv(prefix+"KAFKA_SASL_USERNAME") != "" {
		v := os.Getenv(prefix + "KAFKA_SASL_USERNAME")
		d.Username = &v
	}
	if os.Getenv(prefix+"KAFKA_SASL_PASSWORD") != "" {
		v := os.Getenv(prefix + "KAFKA_SASL_PASSWORD")
		d.Password = &v
	}
	l.Debug("Loaded environment")
	return nil
}

func (d *Kafka) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "kafka",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	b := *flags.KafkaBrokers
	if b != "" {
		d.Brokers = strings.Split(b, ",")
	}
	d.Group = flags.KafkaGroup
	d.Topic = flags.KafkaTopic
	d.EnableTLS = flags.KafkaEnableTLS
	d.TLSInsecure = flags.KafkaTLSInsecure
	d.TLSCert = flags.KafkaCertFile
	d.TLSKey = flags.KafkaKeyFile
	d.TLSCA = flags.KafkaCAFile
	d.EnableSASL = flags.KafkaEnableSasl
	if flags.KafkaSaslType != nil {
		t := SaslType(*flags.KafkaSaslType)
		d.SaslType = &t
	}
	d.Username = flags.KafkaSaslUsername
	d.Password = flags.KafkaSaslPassword
	l.Debug("Loaded flags")
	return nil
}

func (d *Kafka) saslConfig() (sasl.Mechanism, error) {
	l := log.WithFields(log.Fields{
		"pkg": "kafka",
		"fn":  "saslConfig",
	})
	l.Debug("Loading SASL config")
	var m sasl.Mechanism
	var err error
	if d.SaslType != nil && *d.SaslType == SaslTypePlain {
		l.Debug("SASL type is PLAIN")
		m = plain.Mechanism{
			Username: *d.Username,
			Password: *d.Password,
		}
	} else if d.SaslType != nil && *d.SaslType == SaslTypeScram {
		l.Debug("SASL type is SCRAM")
		m, err = scram.Mechanism(scram.SHA512, *d.Username, *d.Password)
		if err != nil {
			l.Error(err)
			return nil, err
		}

	} else {
		l.Debug("SASL type is NONE")
		return nil, nil
	}
	l.Debug("Loaded SASL config")
	return m, nil
}

func (d *Kafka) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "kafka",
		"fn":  "Init",
	})
	l.Debug("Initializing kafka driver")
	kc := kafka.ReaderConfig{
		Brokers: d.Brokers,
	}
	if d.Topic != nil {
		kc.Topic = *d.Topic
	}
	if d.Group != nil {
		kc.GroupID = *d.Group
	}
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	if d.EnableTLS != nil && *d.EnableTLS {
		tc, err := utils.TlsConfig(d.EnableTLS, d.TLSInsecure, d.TLSCA, d.TLSCert, d.TLSKey)
		if err != nil {
			return err
		}
		dialer.TLS = tc
	}
	if d.EnableSASL != nil && *d.EnableSASL {
		m, err := d.saslConfig()
		if err != nil {
			return err
		}
		if m != nil {
			kc.Dialer.SASLMechanism = m
		}
	}
	kc.Dialer = dialer
	d.Client = kafka.NewReader(kc)
	return nil
}

func (d *Kafka) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "kafka",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from kafka")
	m, err := d.Client.ReadMessage(context.Background())
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("Got work")
	return bytes.NewReader(m.Value), nil
}

func (d *Kafka) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "kafka",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from kafka")
	l.Debug("Cleared work")
	return nil
}

func (d *Kafka) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "kafka",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	l.Debug("Handled failure")
	return nil
}

func (d *Kafka) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "kafka",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up")
	if err := d.Client.Close(); err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleaned up")
	return nil
}
