package rabbitmq

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/robertlestak/procx/internal/flags"
	log "github.com/sirupsen/logrus"
)

type RabbitMQ struct {
	Client *amqp.Connection
	URL    string
	Queue  string
}

func (d *RabbitMQ) LoadEnv(prefix string) error {
	if os.Getenv(prefix+"RABBITMQ_URL") != "" {
		d.URL = os.Getenv(prefix + "RABBITMQ_URL")
	}
	if os.Getenv(prefix+"RABBITMQ_QUEUE") != "" {
		d.Queue = os.Getenv(prefix + "RABBITMQ_QUEUE")
	}
	return nil
}

func (d *RabbitMQ) LoadFlags() error {
	d.URL = *flags.RabbitMQURL
	d.Queue = *flags.RabbitMQQueue
	return nil
}

func (d *RabbitMQ) Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "Init",
	})
	l.Debug("Initializing rabbitmq driver")
	conn, err := amqp.Dial(d.URL)
	if err != nil {
		return err
	}
	d.Client = conn
	return nil
}

func (d *RabbitMQ) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWork",
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
	sd := string(msg.Body)
	return &sd, nil
}

func (d *RabbitMQ) ClearWork() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWork",
	})
	l.Debug("Clearing work from rabbitmq")
	return nil
}

func (d *RabbitMQ) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWork",
	})
	l.Debug("Clearing work from rabbitmq")
	return nil
}
