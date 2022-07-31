package procx

import (
	"errors"

	"github.com/robertlestak/procx/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *ProcX) InitPsql() error {
	l := log.WithFields(log.Fields{
		"action": "InitPsql",
		"driver": j.DriverName,
	})
	l.Debug("InitPsql")
	err := client.CreatePsqlClient(
		j.Driver.Psql.Host,
		j.Driver.Psql.Port,
		j.Driver.Psql.User,
		j.Driver.Psql.Password,
		j.Driver.Psql.DBName,
		j.Driver.Psql.SSLMode,
	)
	if err != nil {
		return err
	}
	l.Debug("exited")
	return nil
}

func (q *ProcX) handleFailurePsql() error {
	l := log.WithFields(log.Fields{
		"action": "handleFailurePsql",
		"driver": q.DriverName,
	})
	l.Debug("handleFailurePsql")
	if q.Driver.Psql.FailureQuery == nil {
		l.Debug("no handle failure query")
		return nil
	}
	if err := client.HandleFailurePsql(
		q.Driver.Psql.FailureQuery.Query,
		q.Driver.Psql.FailureQuery.Params,
		q.Driver.Psql.Key,
	); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (q *ProcX) getWorkPsql() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "getWorkPsql",
		"driver": q.DriverName,
	})
	l.Debug("getWorkPsql")
	if q.Driver.Psql.RetrieveQuery == nil {
		l.Error("RetrieveQuery is nil")
		return nil, errors.New("RetrieveQuery is nil")
	}
	m, key, err := client.GetWorkPsql(
		q.Driver.Psql.RetrieveQuery.Query,
		q.Driver.Psql.RetrieveQuery.Params,
		q.Driver.Psql.QueryReturnsKey,
	)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("message received")
	if key != nil {
		q.Driver.Psql.Key = key
	}
	if m == nil {
		l.Debug("no work")
		return nil, nil
	}
	return m, nil
}

func (q *ProcX) clearWorkPsql() error {
	l := log.WithFields(log.Fields{
		"action": "clearWorkPsql",
		"driver": q.DriverName,
	})
	l.Debug("clearWorkPsql")
	if q.Driver.Psql.ClearQuery == nil {
		l.Debug("no clear query")
		return nil
	}
	err := client.ClearWorkPsql(
		q.Driver.Psql.ClearQuery.Query,
		q.Driver.Psql.ClearQuery.Params,
		q.Driver.Psql.Key,
	)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}
