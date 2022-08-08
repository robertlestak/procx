package nats

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type NATS struct {
	Client        *nats.Conn
	URL           string
	Subject       *string
	QueueGroup    *string
	CredsFile     *string
	JWTFile       *string
	NKeyFile      *string
	Username      *string
	Password      *string
	Token         *string
	EnableTLS     *bool
	TLSInsecure   *bool
	TLSCA         *string
	TLSCert       *string
	TLSKey        *string
	ClearResponse *string
	FailResponse  *string
	Key           *string
}

func (d *NATS) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "nats",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	if os.Getenv(prefix+"NATS_URL") != "" {
		d.URL = os.Getenv(prefix + "NATS_URL")
	}
	if os.Getenv(prefix+"NATS_SUBJECT") != "" {
		v := os.Getenv(prefix + "NATS_SUBJECT")
		d.Subject = &v
	}
	if os.Getenv(prefix+"NATS_CREDS_FILE") != "" {
		v := os.Getenv(prefix + "NATS_CREDS_FILE")
		d.CredsFile = &v
	}
	if os.Getenv(prefix+"NATS_JWT_FILE") != "" {
		v := os.Getenv(prefix + "NATS_JWT_FILE")
		d.JWTFile = &v
	}
	if os.Getenv(prefix+"NATS_NKEY_FILE") != "" {
		v := os.Getenv(prefix + "NATS_NKEY_FILE")
		d.NKeyFile = &v
	}
	if os.Getenv(prefix+"NATS_USERNAME") != "" {
		v := os.Getenv(prefix + "NATS_USERNAME")
		d.Username = &v
	}
	if os.Getenv(prefix+"NATS_PASSWORD") != "" {
		v := os.Getenv(prefix + "NATS_PASSWORD")
		d.Password = &v
	}
	if os.Getenv(prefix+"NATS_TOKEN") != "" {
		v := os.Getenv(prefix + "NATS_TOKEN")
		d.Token = &v
	}
	if os.Getenv(prefix+"NATS_ENABLE_TLS") != "" {
		v := os.Getenv(prefix+"NATS_ENABLE_TLS") == "true"
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"NATS_TLS_INSECURE") != "" {
		v := os.Getenv(prefix+"NATS_TLS_INSECURE") == "true"
		d.TLSInsecure = &v
	}
	if os.Getenv(prefix+"NATS_TLS_CA_FILE") != "" {
		v := os.Getenv(prefix + "NATS_TLS_CA_FILE")
		d.TLSCA = &v
	}
	if os.Getenv(prefix+"NATS_TLS_CERT_FILE") != "" {
		v := os.Getenv(prefix + "NATS_TLS_CERT_FILE")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"NATS_TLS_KEY_FILE") != "" {
		v := os.Getenv(prefix + "NATS_TLS_KEY_FILE")
		d.TLSKey = &v
	}
	if os.Getenv(prefix+"NATS_CLEAR_RESPONSE") != "" {
		v := os.Getenv(prefix + "NATS_CLEAR_RESPONSE")
		d.ClearResponse = &v
	}
	if os.Getenv(prefix+"NATS_FAIL_RESPONSE") != "" {
		v := os.Getenv(prefix + "NATS_FAIL_RESPONSE")
		d.FailResponse = &v
	}
	if os.Getenv(prefix+"NATS_QUEUE_GROUP") != "" {
		v := os.Getenv(prefix + "NATS_QUEUE_GROUP")
		d.QueueGroup = &v
	}
	return nil
}

func (d *NATS) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "nats",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.URL = *flags.NATSURL
	d.Subject = flags.NATSSubject
	d.CredsFile = flags.NATSCredsFile
	d.JWTFile = flags.NATSJWTFile
	d.NKeyFile = flags.NATSNKeyFile
	d.Username = flags.NATSUsername
	d.Password = flags.NATSPassword
	d.QueueGroup = flags.NATSQueueGroup
	d.Token = flags.NATSToken
	d.EnableTLS = flags.NATSEnableTLS
	d.TLSInsecure = flags.NATSTLSInsecure
	d.TLSCA = flags.NATSTLSCAFile
	d.TLSCert = flags.NATSTLSCertFile
	d.TLSKey = flags.NATSTLSKeyFile
	d.ClearResponse = flags.NATSClearResponse
	d.FailResponse = flags.NATSFailResponse
	l.Debug("Loaded flags")
	return nil
}

func (d *NATS) tlsConfig() (*tls.Config, error) {
	l := log.WithFields(log.Fields{
		"pkg": "nats",
		"fn":  "tlsConfig",
	})
	l.Debug("Creating TLS config")
	tc := &tls.Config{}
	if d.EnableTLS != nil && *d.EnableTLS {
		l.Debug("Enabling TLS")
		if d.TLSInsecure != nil && *d.TLSInsecure {
			l.Debug("Enabling TLS insecure")
			tc.InsecureSkipVerify = true
		}
		if d.TLSCA != nil && *d.TLSCA != "" {
			l.Debug("Enabling TLS CA")
			caCert, err := ioutil.ReadFile(*d.TLSCA)
			if err != nil {
				l.Errorf("%+v", err)
				return tc, err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tc.RootCAs = caCertPool
		}
		if d.TLSCert != nil && *d.TLSCert != "" {
			l.Debug("Enabling TLS cert")
			cert, err := tls.LoadX509KeyPair(*d.TLSCert, *d.TLSKey)
			if err != nil {
				l.Errorf("%+v", err)
				return tc, err
			}
			tc.Certificates = []tls.Certificate{cert}
		}
	}
	l.Debug("Created TLS config")
	return tc, nil
}

func (d *NATS) authOpts() []nats.Option {
	l := log.WithFields(log.Fields{
		"pkg": "nats",
		"fn":  "authOpts",
	})
	l.Debug("Creating auth options")
	opts := []nats.Option{}
	if d.CredsFile != nil && *d.CredsFile != "" {
		l.Debug("Enabling creds file")
		opts = append(opts, nats.UserCredentials(*d.CredsFile))
	}
	if d.Username != nil && *d.Username != "" {
		l.Debug("Enabling username")
		opts = append(opts, nats.UserInfo(*d.Username, *d.Password))
	}
	if d.Token != nil && *d.Token != "" {
		l.Debug("Enabling token")
		opts = append(opts, nats.Token(*d.Token))
	}
	if d.JWTFile != nil && *d.JWTFile != "" && d.NKeyFile != nil && *d.NKeyFile != "" {
		l.Debug("Enabling JWT file")
		opts = append(opts, nats.UserCredentials(*d.JWTFile, *d.NKeyFile))
	}
	l.Debug("Created auth options")
	return opts
}

func (d *NATS) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "nats",
		"fn":  "Init",
	})
	l.Debug("Initializing nats driver")
	if d.URL == "" {
		l.Error("url is empty")
		return errors.New("url is empty")
	}
	opts := []nats.Option{}
	if d.EnableTLS != nil && *d.EnableTLS {
		l.Debug("Enabling TLS")
		tc, err := d.tlsConfig()
		if err != nil {
			l.Errorf("%+v", err)
			return err
		}
		opts = append(opts, nats.Secure(tc))
	}
	opts = append(opts, d.authOpts()...)
	nc, err := nats.Connect(d.URL, opts...)
	if err != nil {
		l.Errorf("error connecting to nats: %v", err)
		return err
	}
	d.Client = nc
	return nil
}

func (d *NATS) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "nats",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from nats")
	ch := make(chan *nats.Msg, 64)
	var sub *nats.Subscription
	var err error
	if d.QueueGroup != nil && *d.QueueGroup != "" {
		l.Debug("Enabling queue group")
		sub, err = d.Client.ChanQueueSubscribe(*d.Subject, *d.QueueGroup, ch)
	} else {
		sub, err = d.Client.ChanSubscribe(*d.Subject, ch)
	}
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	defer sub.Unsubscribe()
	msg := <-ch
	l.Debug("Got work from nats")
	if msg == nil {
		l.Debug("No work found")
		return nil, nil
	}
	if msg.Reply != "" {
		v := msg.Reply
		d.Key = &v
	}
	return bytes.NewReader(msg.Data), nil
}

func (d *NATS) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "nats",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from nats")
	if d.Key == nil || *d.Key == "" {
		return nil
	}
	var resp []byte
	if d.ClearResponse != nil && *d.ClearResponse != "" {
		resp = []byte(*d.ClearResponse)
	} else {
		resp = []byte("0")
	}
	if err := d.Client.Publish(*d.Key, resp); err != nil {
		l.Errorf("%+v", err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func (d *NATS) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "nats",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	if d.Key == nil || *d.Key == "" {
		return nil
	}
	var resp []byte
	if d.FailResponse != nil && *d.FailResponse != "" {
		resp = []byte(*d.FailResponse)
	} else {
		resp = []byte("1")
	}
	if err := d.Client.Publish(*d.Key, resp); err != nil {
		l.Errorf("%+v", err)
		return err
	}
	l.Debug("Handled failure")
	return nil
}

func (d *NATS) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "nats",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up")
	if d.Client != nil {
		d.Client.Close()
	}
	l.Debug("Cleaned up")
	return nil
}
