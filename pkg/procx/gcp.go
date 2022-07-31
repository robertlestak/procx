package procx

import (
	"github.com/robertlestak/procx/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *ProcX) InitGCPPubSub() error {
	l := log.WithFields(log.Fields{
		"app": "procx",
	})
	l.Debug("starting")
	err := client.CreateGCPPubSubClient(j.Driver.GCP.ProjectID)
	if err != nil {
		return err
	}
	l.Debug("exited")
	return nil
}

func (j *ProcX) getWorkGCPPubSub() (*string, error) {
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
	j.Driver.GCP.pubSubMessage = m
	return &body, nil
}

func (j *ProcX) clearWorkGCPPubSub() error {
	l := log.WithFields(log.Fields{
		"action": "clearWorkGCPPubSub",
		"driver": j.DriverName,
	})
	l.Debug("clearWorkGCPPubSub")
	j.Driver.GCP.pubSubMessage.Ack()
	l.Debug("acknowledged")
	return nil
}
