package activemq

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"

	stomp "github.com/go-stomp/stomp/v3"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type ActiveMQ struct {
	Client  *stomp.Conn
	Address string
	Type    *string
	Name    *string
	// TLS
	EnableTLS   *bool
	TLSInsecure *bool
	TLSCert     *string
	TLSKey      *string
	TLSCA       *string
	message     *stomp.Message
}

func (d *ActiveMQ) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "activemq",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	if os.Getenv(prefix+"ACTIVEMQ_ADDRESS") != "" {
		d.Address = os.Getenv(prefix + "ACTIVEMQ_ADDRESS")
	}
	if os.Getenv(prefix+"ACTIVEMQ_TYPE") != "" {
		v := os.Getenv(prefix + "ACTIVEMQ_TYPE")
		d.Type = &v
	}
	if os.Getenv(prefix+"ACTIVEMQ_NAME") != "" {
		v := os.Getenv(prefix + "ACTIVEMQ_NAME")
		d.Name = &v
	}
	if os.Getenv(prefix+"ACTIVEMQ_ENABLE_TLS") != "" {
		v := os.Getenv(prefix+"ACTIVEMQ_ENABLE_TLS") == "true"
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"ACTIVEMQ_TLS_INSECURE") != "" {
		v := os.Getenv(prefix+"ACTIVEMQ_TLS_INSECURE") == "true"
		d.TLSInsecure = &v
	}
	if os.Getenv(prefix+"ACTIVEMQ_TLS_CERT_FILE") != "" {
		v := os.Getenv(prefix + "ACTIVEMQ_TLS_CERT_FILE")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"ACTIVEMQ_TLS_KEY_FILE") != "" {
		v := os.Getenv(prefix + "ACTIVEMQ_TLS_KEY_FILE")
		d.TLSKey = &v
	}
	if os.Getenv(prefix+"ACTIVEMQ_TLS_CA_FILE") != "" {
		v := os.Getenv(prefix + "ACTIVEMQ_TLS_CA_FILE")
		d.TLSCA = &v
	}
	l.Debug("Loaded environment")
	return nil
}

func (d *ActiveMQ) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "activemq",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.Address = *flags.ActiveMQAddress
	d.Type = flags.ActiveMQType
	d.Name = flags.ActiveMQName
	d.EnableTLS = flags.ActiveMQEnableTLS
	d.TLSInsecure = flags.ActiveMQTLSInsecure
	d.TLSCert = flags.ActiveMQTLSCert
	d.TLSKey = flags.ActiveMQTLSKey
	d.TLSCA = flags.ActiveMQTLSCA
	l.Debug("Loaded flags")
	return nil
}

func (d *ActiveMQ) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "activemq",
		"fn":  "Init",
	})
	l.Debug("Initializing activemq driver")
	// create connection
	var err error
	var conn net.Conn
	if *d.EnableTLS {
		l.Debug("Creating TLS connection")
		tc, err := utils.TlsConfig(d.EnableTLS, d.TLSInsecure, d.TLSCA, d.TLSCert, d.TLSKey)
		if err != nil {
			return err
		}

		conn, err = tls.Dial("tcp", d.Address, tc)
		if err != nil {
			return err
		}
	} else {
		l.Debug("Creating non-TLS connection")
		conn, err = net.Dial("tcp", d.Address)
		if err != nil {
			return err
		}
	}
	d.Client, err = stomp.Connect(conn)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	l.Debug("Initialized activemq driver")
	return nil
}

func (d *ActiveMQ) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "activemq",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from activemq")
	if d.Type == nil || *d.Type == "" {
		return nil, fmt.Errorf("ActiveMQ type is not set")
	}
	if d.Name == nil || *d.Name == "" {
		return nil, fmt.Errorf("ActiveMQ name is not set")
	}
	p := fmt.Sprintf("/%s/%s", *d.Type, *d.Name)
	sub, err := d.Client.Subscribe(p, stomp.AckClient)
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	msg := <-sub.C
	if msg.Err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	l.Debug("Got work")
	d.message = msg
	return bytes.NewReader(msg.Body), nil
}

func (d *ActiveMQ) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "activemq",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from activemq")
	if d.message == nil {
		return nil
	}
	if err := d.Client.Ack(d.message); err != nil {
		l.Errorf("%+v", err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func (d *ActiveMQ) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "activemq",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	if d.message == nil {
		return nil
	}
	if err := d.Client.Nack(d.message); err != nil {
		l.Errorf("%+v", err)
		return err
	}
	l.Debug("Handled failure")
	return nil
}

func (d *ActiveMQ) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "activemq",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up")
	if err := d.Client.Disconnect(); err != nil {
		l.Errorf("%+v", err)
		return err
	}
	l.Debug("Cleaned up")
	return nil
}
