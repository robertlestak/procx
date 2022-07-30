package qjob

import (
	"github.com/robertlestak/qjob/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *QJob) InitGCPPubSub() error {
	l := log.WithFields(log.Fields{
		"app": "qjob",
	})
	l.Debug("starting")
	err := client.CreateGCPPubSubClient(j.Driver.GCP.ProjectID)
	if err != nil {
		return err
	}
	l.Debug("exited")
	return nil
}

func (j *QJob) getWorkGCPPubSub() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "getWorkGCPPubSub",
		"driver": j.DriverName,
	})
	l.Debug("getWorkGCPPubSub")
	m, err := client.ReceiveMessageGCPPubSub(j.Driver.GCP.SubscriptionName)
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
	body := string(m.Data)
	j.Driver.GCP.PubSubMessage = m
	return &body, nil
}

func (j *QJob) clearWorkGCPPubSub() error {
	l := log.WithFields(log.Fields{
		"action": "clearWorkGCPPubSub",
		"driver": j.DriverName,
	})
	l.Debug("clearWorkGCPPubSub")
	j.Driver.GCP.PubSubMessage.Ack()
	l.Debug("acknowledged")
	return nil
}
