package procx

import (
	"github.com/robertlestak/procx/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *ProcX) InitRabbitMQ() error {
	l := log.WithFields(log.Fields{
		"app": "procx",
	})
	l.Debug("starting")
	err := client.CreateRabbitMQClient(j.Driver.RabbitMQ.URL)
	if err != nil {
		return err
	}
	client.RabbitMQQueue = j.Driver.RabbitMQ.Queue
	l.Debug("exited")
	return nil
}

func (j *ProcX) getWorkRabbitMQ() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "getWorkRabbitMQ",
		"driver": j.DriverName,
	})
	l.Debug("getWorkRabbitMQ")
	m, err := client.ReceiveMessageRabbitMQ()
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("received message")
	if m == nil {
		l.Debug("no message")
		return nil, nil
	}
	l.Debug("message received")
	body := string(m.Body)
	return &body, nil
}

func (j *ProcX) clearWorkRabbitMQ() error {
	l := log.WithFields(log.Fields{
		"action": "clearWorkRabbitMQ",
		"driver": j.DriverName,
	})
	l.Debug("clearWorkRabbitMQ")
	client.RabbitMQClient.Close()
	return nil
}
