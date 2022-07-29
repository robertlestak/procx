package qjob

import (
	"github.com/robertlestak/qjob/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *QJob) InitAWSSQS() error {
	l := log.WithFields(log.Fields{
		"app": "qjob",
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

func (j *QJob) getWorkSQS() (*string, error) {
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

func (j *QJob) clearWorkSQS() error {
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
