package pulsar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	pulsarlog "github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type Pulsar struct {
	Client          pulsar.Client
	Address         string
	Subscription    *string
	Topic           *string
	TopicsPattern   *string
	Topics          []string
	AuthToken       *string
	AuthTokenFile   *string
	AuthCertPath    *string
	AuthKeyPath     *string
	AuthOAuthParams *map[string]string
	// TLS
	TLSTrustCertsFilePath      *string
	TLSAllowInsecureConnection *bool
	TLSValidateHostname        *bool
	message                    pulsar.Message
	consumer                   pulsar.Consumer
}

func (d *Pulsar) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "pulsar",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	if os.Getenv(prefix+"PULSAR_ADDRESS") != "" {
		d.Address = os.Getenv(prefix + "PULSAR_ADDRESS")
	}
	if os.Getenv(prefix+"PULSAR_SUBSCRIPTION") != "" {
		v := os.Getenv(prefix + "PULSAR_SUBSCRIPTION")
		d.Subscription = &v
	}
	if os.Getenv(prefix+"PULSAR_TOPIC") != "" {
		v := os.Getenv(prefix + "PULSAR_TOPIC")
		d.Topic = &v
	}
	if os.Getenv(prefix+"PULSAR_TOPICS") != "" {
		v := os.Getenv(prefix + "PULSAR_TOPICS")
		d.Topics = strings.Split(v, ",")
	}
	if os.Getenv(prefix+"PULSAR_TOPICS_PATTERN") != "" {
		v := os.Getenv(prefix + "PULSAR_TOPICS_PATTERN")
		d.TopicsPattern = &v
	}
	if os.Getenv(prefix+"PULSAR_TLS_TRUST_CERTS_FILE") != "" {
		v := os.Getenv(prefix + "PULSAR_TLS_TRUST_CERTS_FILE")
		d.TLSTrustCertsFilePath = &v
	}
	if os.Getenv(prefix+"PULSAR_TLS_ALLOW_INSECURE_CONNECTION") != "" {
		v := os.Getenv(prefix+"PULSAR_TLS_ALLOW_INSECURE_CONNECTION") == "true"
		d.TLSAllowInsecureConnection = &v
	}
	if os.Getenv(prefix+"PULSAR_TLS_VALIDATE_HOSTNAME") != "" {
		v := os.Getenv(prefix+"PULSAR_TLS_VALIDATE_HOSTNAME") == "true"
		d.TLSValidateHostname = &v
	}
	if os.Getenv(prefix+"PULSAR_AUTH_TOKEN") != "" {
		v := os.Getenv(prefix + "PULSAR_AUTH_TOKEN")
		d.AuthToken = &v
	}
	if os.Getenv(prefix+"PULSAR_AUTH_TOKEN_FILE") != "" {
		v := os.Getenv(prefix + "PULSAR_AUTH_TOKEN_FILE")
		d.AuthTokenFile = &v
	}
	if os.Getenv(prefix+"PULSAR_AUTH_CERT_FILE") != "" {
		v := os.Getenv(prefix + "PULSAR_AUTH_CERT_FILE")
		d.AuthCertPath = &v
	}
	if os.Getenv(prefix+"PULSAR_AUTH_KEY_FILE") != "" {
		v := os.Getenv(prefix + "PULSAR_AUTH_KEY_FILE")
		d.AuthKeyPath = &v
	}
	if os.Getenv(prefix+"PULSAR_AUTH_OAUTH_PARAMS") != "" {
		v := os.Getenv(prefix + "PULSAR_AUTH_OAUTH_PARAMS")
		var m map[string]string
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return err
		}
		d.AuthOAuthParams = &m
	}
	l.Debug("Loaded environment")
	return nil
}

func (d *Pulsar) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "pulsar",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.Address = *flags.PulsarAddress
	d.Subscription = flags.PulsarSubscription
	d.Topic = flags.PulsarTopic
	d.TopicsPattern = flags.PulsarTopicsPattern
	ts := strings.Split(*flags.PulsarTopics, ",")
	for _, t := range ts {
		if t != "" {
			nt := strings.TrimSpace(t)
			d.Topics = append(d.Topics, nt)
		}
	}
	d.TLSTrustCertsFilePath = flags.PulsarTLSTrustCertsFilePath
	d.TLSAllowInsecureConnection = flags.PulsarTLSAllowInsecureConnection
	d.TLSValidateHostname = flags.PulsarTLSValidateHostname
	d.AuthToken = flags.PulsarAuthToken
	d.AuthTokenFile = flags.PulsarAuthTokenFile
	d.AuthCertPath = flags.PulsarAuthCertFile
	d.AuthKeyPath = flags.PulsarAuthKeyFile
	oauthParams := make(map[string]string)
	if flags.PulsarAuthOAuthParams != nil && *flags.PulsarAuthOAuthParams != "" {
		if err := json.Unmarshal([]byte(*flags.PulsarAuthOAuthParams), &oauthParams); err != nil {
			l.WithError(err).Error("Failed to parse PulsarAuthOAuthParams")
			return err
		}
		d.AuthOAuthParams = &oauthParams
	}
	l.Debug("Loaded flags")
	return nil
}

func (d *Pulsar) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "pulsar",
		"fn":  "Init",
	})

	l.Debug("Initializing pulsar driver")
	ll := log.StandardLogger()
	ll.Level = log.FatalLevel
	lr := pulsarlog.NewLoggerWithLogrus(ll)
	up := "pulsar://"
	opts := pulsar.ClientOptions{
		OperationTimeout:           30 * time.Second,
		ConnectionTimeout:          30 * time.Second,
		TLSTrustCertsFilePath:      *d.TLSTrustCertsFilePath,
		TLSAllowInsecureConnection: *d.TLSAllowInsecureConnection,
		TLSValidateHostname:        *d.TLSValidateHostname,
		Logger:                     lr,
	}
	if *d.AuthCertPath != "" && *d.AuthKeyPath != "" {
		opts.Authentication = pulsar.NewAuthenticationTLS(*d.AuthCertPath, *d.AuthKeyPath)
		up = "pulsar+ssl://"
	}
	if *d.AuthToken != "" {
		opts.Authentication = pulsar.NewAuthenticationToken(*d.AuthToken)
	}
	if *d.AuthTokenFile != "" {
		opts.Authentication = pulsar.NewAuthenticationTokenFromFile(*d.AuthTokenFile)
	}
	if *d.AuthOAuthParams != nil && len(*d.AuthOAuthParams) > 0 {
		opts.Authentication = pulsar.NewAuthenticationOAuth2(*d.AuthOAuthParams)
	}
	opts.URL = up + d.Address
	client, err := pulsar.NewClient(opts)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	d.Client = client
	l.Debug("Initialized pulsar driver")
	return nil
}

func (d *Pulsar) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "pulsar",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from pulsar")
	opts := pulsar.ConsumerOptions{
		SubscriptionName: *d.Subscription,
	}
	if len(d.Topics) > 0 {
		opts.Topics = d.Topics
	} else if *d.TopicsPattern != "" {
		opts.TopicsPattern = *d.TopicsPattern
	} else if *d.Topic != "" {
		opts.Topic = *d.Topic
	} else {
		l.Error("No topic specified")
		return nil, errors.New("no topic specified")
	}
	consumer, err := d.Client.Subscribe(opts)
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	d.consumer = consumer
	msg, err := consumer.Receive(context.Background())
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	l.Debug("Got work")
	d.message = msg
	return bytes.NewReader(msg.Payload()), nil
}

func (d *Pulsar) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "pulsar",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from pulsar")
	d.consumer.Ack(d.message)
	l.Debug("Cleared work")
	return nil
}

func (d *Pulsar) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "pulsar",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	d.consumer.Nack(d.message)
	l.Debug("Handled failure")
	return nil
}

func (d *Pulsar) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "pulsar",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up")
	d.Client.Close()
	l.Debug("Cleaned up")
	return nil
}
