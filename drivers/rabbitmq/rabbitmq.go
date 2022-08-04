package rabbitmq

import (
	"bytes"
	"io"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type RabbitMQ struct {
	Client *amqp.Connection
	URL    string
	Queue  string
}

func (d *RabbitMQ) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "rabbitmq",
		"fn":  "LoadEnv",
	})
	l.Debug("LoadEnv")
	if os.Getenv(prefix+"RABBITMQ_URL") != "" {
		d.URL = os.Getenv(prefix + "RABBITMQ_URL")
	}
	if os.Getenv(prefix+"RABBITMQ_QUEUE") != "" {
		d.Queue = os.Getenv(prefix + "RABBITMQ_QUEUE")
	}
	return nil
}

func (d *RabbitMQ) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "rabbitmq",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.URL = *flags.RabbitMQURL
	d.Queue = *flags.RabbitMQQueue
	return nil
}

func (d *RabbitMQ) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "rabbitmq",
		"fn":  "Init",
	})
	l.Debug("Initializing rabbitmq driver")
	conn, err := amqp.Dial(d.URL)
	if err != nil {
		return err
	}
	d.Client = conn
	return nil
}

func (d *RabbitMQ) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "rabbitmq",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from rabbitmq")
	ch, err := d.Client.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(d.Queue, false, false, true, false, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	msg := <-msgs
	l.Debug("Received message from rabbitmq")
	if msg.Body == nil {
		return nil, nil
	}
	return bytes.NewReader(msg.Body), nil
}

func (d *RabbitMQ) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "rabbitmq",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from rabbitmq")
	return nil
}

func (d *RabbitMQ) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "rabbitmq",
		"fn":  "HandleFailure",
	})
	l.Debug("Clearing work from rabbitmq")
	return nil
}

func (d *RabbitMQ) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "rabbitmq",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up rabbitmq driver")
	if err := d.Client.Close(); err != nil {
		return err
	}
	return nil
}
