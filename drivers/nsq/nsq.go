package nsq

import (
	"bytes"
	"errors"
	"io"
	"os"

	nsq "github.com/nsqio/go-nsq"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type NSQ struct {
	Client            *nsq.Consumer
	NsqLookupdAddress *string
	NsqdAddress       *string
	Topic             *string
	Channel           *string
	data              chan []byte
	// TLS
	EnableTLS   *bool
	TLSInsecure *bool
	TLSCert     *string
	TLSKey      *string
	TLSCA       *string
}

func (d *NSQ) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "nsq",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	if os.Getenv(prefix+"NSQ_NSQLOOKUPD_ADDRESS") != "" {
		v := os.Getenv(prefix + "NSQ_NSQLOOKUPD_ADDRESS")
		d.NsqLookupdAddress = &v
	}
	if os.Getenv(prefix+"NSQ_NSQD_ADDRESS") != "" {
		v := os.Getenv(prefix + "NSQ_NSQD_ADDRESS")
		d.NsqdAddress = &v
	}
	if os.Getenv(prefix+"NSQ_TOPIC") != "" {
		v := os.Getenv(prefix + "NSQ_TOPIC")
		d.Topic = &v
	}
	if os.Getenv(prefix+"NSQ_CHANNEL") != "" {
		v := os.Getenv(prefix + "NSQ_CHANNEL")
		d.Channel = &v
	}
	if os.Getenv(prefix+"NSQ_ENABLE_TLS") != "" {
		v := os.Getenv(prefix+"NSQ_ENABLE_TLS") == "true"
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"NSQ_TLS_INSECURE") != "" {
		v := os.Getenv(prefix+"NSQ_TLS_INSECURE") == "true"
		d.TLSInsecure = &v
	}
	if os.Getenv(prefix+"NSQ_TLS_CERT_FILE") != "" {
		v := os.Getenv(prefix + "NSQ_TLS_CERT_FILE")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"NSQ_TLS_KEY_FILE") != "" {
		v := os.Getenv(prefix + "NSQ_TLS_KEY_FILE")
		d.TLSKey = &v
	}
	if os.Getenv(prefix+"NSQ_TLS_CA_FILE") != "" {
		v := os.Getenv(prefix + "NSQ_TLS_CA_FILE")
		d.TLSCA = &v
	}
	l.Debug("Loaded environment")
	return nil
}

func (d *NSQ) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "nsq",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.NsqLookupdAddress = flags.NSQNSQLookupdAddress
	d.NsqdAddress = flags.NSQNSQDAddress
	d.Topic = flags.NSQTopic
	d.Channel = flags.NSQChannel
	d.EnableTLS = flags.NSQEnableTLS
	d.TLSInsecure = flags.NSQTLSSkipVerify
	d.TLSCert = flags.NSQCertFile
	d.TLSKey = flags.NSQKeyFile
	d.TLSCA = flags.NSQCAFile
	l.Debug("Loaded flags")
	return nil
}

func (d *NSQ) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "nsq",
		"fn":  "Init",
	})
	l.Debug("Initializing nsq driver")
	cfg := nsq.NewConfig()
	cfg.MaxInFlight = 1
	if d.EnableTLS != nil && *d.EnableTLS {
		cfg.TlsV1 = true
		t, err := utils.TlsConfig(d.EnableTLS, d.TLSInsecure, d.TLSCA, d.TLSCert, d.TLSKey)
		if err != nil {
			l.Errorf("%+v", err)
			return err
		}
		cfg.TlsConfig = t
	}
	consumer, err := nsq.NewConsumer(*d.Topic, *d.Channel, cfg)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	switch os.Getenv("NSQ_LOG_LEVEL") {
	case "debug":
		consumer.SetLoggerLevel(nsq.LogLevelDebug)
	case "info":
		consumer.SetLoggerLevel(nsq.LogLevelInfo)
	case "warn":
		consumer.SetLoggerLevel(nsq.LogLevelWarning)
	case "error":
		consumer.SetLoggerLevel(nsq.LogLevelError)
	case "fatal":
		consumer.SetLoggerLevel(nsq.LogLevelError)
	default:
		consumer.SetLoggerLevel(nsq.LogLevelError)
	}
	d.Client = consumer
	return nil
}

func (d *NSQ) handleMessage(msg *nsq.Message) error {
	d.data <- msg.Body
	return nil
}

func (d *NSQ) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "nsq",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from nsq")
	d.data = make(chan []byte)
	d.Client.AddHandler(nsq.HandlerFunc(d.handleMessage))
	var err error
	if d.NsqLookupdAddress != nil && *d.NsqLookupdAddress != "" {
		err = d.Client.ConnectToNSQLookupd(*d.NsqLookupdAddress)
	} else if d.NsqdAddress != nil && *d.NsqdAddress != "" {
		err = d.Client.ConnectToNSQD(*d.NsqdAddress)
	} else {
		return nil, errors.New("no nsqd address or nsqlookupd address specified")
	}
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	msg := <-d.data
	l.Debug("Got work")
	return bytes.NewReader(msg), nil
}

func (d *NSQ) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "nsq",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from nsq")
	l.Debug("Cleared work")
	return nil
}

func (d *NSQ) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "nsq",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	l.Debug("Handled failure")
	return nil
}

func (d *NSQ) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "nsq",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up")
	d.Client.Stop()
	<-d.Client.StopChan
	l.Debug("Cleaned up")
	return nil
}
