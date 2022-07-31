package qjob

import (
	"github.com/robertlestak/qjob/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *QJob) InitCentauri() error {
	l := log.WithFields(log.Fields{
		"action": "InitCentauri",
		"driver": j.DriverName,
	})
	l.Debug("InitCentauri")
	err := client.CreateCentariClient(
		j.Driver.Centauri.PeerURL,
		j.Driver.Centauri.PrivateKey,
	)
	if err != nil {
		return err
	}
	l.Debug("exited")
	return nil
}

func (q *QJob) handleFailureCentauri() error {
	l := log.WithFields(log.Fields{
		"action": "handleFailureCentauri",
		"driver": q.DriverName,
	})
	l.Debug("handleFailureCentauri")
	client.HandleFailureCentauri(
		*q.Driver.Centauri.Channel,
		q.Driver.Centauri.Key,
	)
	return nil
}

func (q *QJob) getWorkCentauri() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "getWorkCentauri",
		"driver": q.DriverName,
	})
	l.Debug("getWorkCentauri")
	m, key, err := client.GetWorkCentauri(
		*q.Driver.Centauri.Channel,
	)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("message received")
	if key != nil {
		q.Driver.Centauri.Key = key
	}
	if m == nil {
		l.Debug("no work")
		return nil, nil
	}
	return m, nil
}

func (q *QJob) clearWorkCentauri() error {
	l := log.WithFields(log.Fields{
		"action": "clearWorkCentauri",
		"driver": q.DriverName,
	})
	l.Debug("clearWorkCentauri")
	err := client.ClearWorkCentauri(
		*q.Driver.Centauri.Channel,
		q.Driver.Centauri.Key,
	)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}
