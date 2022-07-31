package procx

import (
	"github.com/robertlestak/procx/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *ProcX) InitAWSSQS() error {
	l := log.WithFields(log.Fields{
		"app": "procx",
	})
	l.Debug("starting")
	c, err := client.CreateSQSClient(j.Driver.AWS.Region, j.Driver.AWS.RoleARN)
	if err != nil {
		return err
	}
	client.SQSClient = c
	client.SQSQueueURL = j.Driver.AWS.SQSQueueURL
	l.Debug("exited")
	return nil
}

func (j *ProcX) getWorkSQS() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "getWorkSQS",
		"driver": j.DriverName,
	})
	l.Debug("getWorkSQS")
	m, err := client.ReceiveMessageSQS()
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
	client.SQSReceiptHandle = *m.ReceiptHandle
	return m.Body, nil
}

func (j *ProcX) clearWorkSQS() error {
	l := log.WithFields(log.Fields{
		"action": "clearWorkSQS",
		"driver": j.DriverName,
	})
	l.Debug("clearWorkSQS")
	err := client.DeleteMessageSQS()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("message deleted")
	return nil
}
