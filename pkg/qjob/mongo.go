package qjob

import (
	"errors"

	"github.com/robertlestak/qjob/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *QJob) InitMongo() error {
	l := log.WithFields(log.Fields{
		"action": "InitMongo",
		"driver": j.DriverName,
	})
	l.Debug("InitMongo")
	err := client.CreateMongoClient(
		j.Driver.Mongo.Host,
		j.Driver.Mongo.Port,
		j.Driver.Mongo.User,
		j.Driver.Mongo.Password,
		j.Driver.Mongo.DBName,
	)
	if err != nil {
		return err
	}
	l.Debug("exited")
	return nil
}

func (q *QJob) HandleFailureMongo() error {
	l := log.WithFields(log.Fields{
		"action": "HandleFailureMongo",
		"driver": q.DriverName,
	})
	l.Debug("HandleFailureMongo")
	if q.Driver.Mongo.FailureQuery == nil {
		l.Debug("no handle failure query")
		return nil
	}
	if err := client.HandleFailureMongo(
		q.Driver.Mongo.DBName,
		q.Driver.Mongo.Collection,
		*q.Driver.Mongo.FailureQuery,
		q.Driver.Mongo.Key,
	); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (q *QJob) GetWorkMongo() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "InitMongo",
		"driver": q.DriverName,
	})
	l.Debug("InitMongo")
	if q.Driver.Mongo.RetrieveQuery == nil {
		l.Error("RetrieveQuery is nil")
		return nil, errors.New("RetrieveQuery is nil")
	}
	m, key, err := client.GetWorkMongo(
		q.Driver.Mongo.DBName,
		q.Driver.Mongo.Collection,
		*q.Driver.Mongo.RetrieveQuery,
	)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("message received")
	if key != nil {
		q.Driver.Mongo.Key = key
	}
	if m == nil {
		l.Debug("no work")
		return nil, nil
	}
	return m, nil
}

func (q *QJob) ClearWorkMongo() error {
	l := log.WithFields(log.Fields{
		"action": "ClearWorkMongo",
		"driver": q.DriverName,
	})
	l.Debug("ClearWorkMongo")
	if q.Driver.Mongo.ClearQuery == nil {
		l.Debug("no clear query")
		return nil
	}
	err := client.ClearWorkMongo(
		q.Driver.Mongo.DBName,
		q.Driver.Mongo.Collection,
		*q.Driver.Mongo.ClearQuery,
		q.Driver.Mongo.Key,
	)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}
