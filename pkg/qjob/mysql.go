package qjob

import (
	"errors"

	"github.com/robertlestak/qjob/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *QJob) InitMysql() error {
	l := log.WithFields(log.Fields{
		"action": "InitMysql",
		"driver": j.DriverName,
	})
	l.Debug("InitMysql")
	err := client.CreateMySqlClient(
		j.Driver.Mysql.Host,
		j.Driver.Mysql.Port,
		j.Driver.Mysql.User,
		j.Driver.Mysql.Password,
		j.Driver.Mysql.DBName,
	)
	if err != nil {
		return err
	}
	l.Debug("exited")
	return nil
}

func (q *QJob) HandleFailureMysql() error {
	l := log.WithFields(log.Fields{
		"action": "HandleFailureMysql",
		"driver": q.DriverName,
	})
	l.Debug("HandleFailureMysql")
	if q.Driver.Mysql.FailureQuery == nil {
		l.Debug("no handle failure query")
		return nil
	}
	if err := client.HandleFailureMysql(
		q.Driver.Mysql.FailureQuery.Query,
		q.Driver.Mysql.FailureQuery.Params,
		q.Driver.Mysql.Key,
	); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (q *QJob) GetWorkMysql() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "GetWorkMysql",
		"driver": q.DriverName,
	})
	l.Debug("GetWorkMysql")
	if q.Driver.Mysql.RetrieveQuery == nil {
		l.Error("RetrieveQuery is nil")
		return nil, errors.New("RetrieveQuery is nil")
	}
	m, key, err := client.GetWorkMysql(
		q.Driver.Mysql.RetrieveQuery.Query,
		q.Driver.Mysql.RetrieveQuery.Params,
		q.Driver.Mysql.QueryReturnsKey,
	)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("message received")
	if key != nil {
		q.Driver.Mysql.Key = key
	}
	if m == nil {
		l.Debug("no work")
		return nil, nil
	}
	return m, nil
}

func (q *QJob) ClearWorkMysql() error {
	l := log.WithFields(log.Fields{
		"action": "ClearWorkMysql",
		"driver": q.DriverName,
	})
	l.Debug("ClearWorkMysql")
	if q.Driver.Mysql.ClearQuery == nil {
		l.Debug("no clear query")
		return nil
	}
	err := client.ClearWorkMysql(
		q.Driver.Mysql.ClearQuery.Query,
		q.Driver.Mysql.ClearQuery.Params,
		q.Driver.Mysql.Key,
	)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}
