package procx

import (
	"errors"

	"github.com/robertlestak/procx/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *ProcX) InitCassandra() error {
	l := log.WithFields(log.Fields{
		"action": "InitCassandra",
		"driver": j.DriverName,
	})
	l.Debug("InitCassandra")
	err := client.CreateCassandraClient(
		j.Driver.Cassandra.Hosts,
		j.Driver.Cassandra.User,
		j.Driver.Cassandra.Password,
		j.Driver.Cassandra.Consistency,
		j.Driver.Cassandra.Keyspace,
	)
	if err != nil {
		return err
	}
	l.Debug("exited")
	return nil
}

func (q *ProcX) HandleFailureCassandra() error {
	l := log.WithFields(log.Fields{
		"action": "HandleFailureCassandra",
		"driver": q.DriverName,
	})
	l.Debug("HandleFailureCassandra")
	if q.Driver.Cassandra.FailureQuery == nil {
		l.Debug("no handle failure query")
		return nil
	}
	if err := client.HandleFailureCassandra(
		q.Driver.Cassandra.FailureQuery.Query,
		q.Driver.Cassandra.FailureQuery.Params,
		q.Driver.Cassandra.Key,
	); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (q *ProcX) GetWorkCassandra() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "GetWorkCassandra",
		"driver": q.DriverName,
	})
	l.Debug("GetWorkCassandra")
	if q.Driver.Cassandra.RetrieveQuery == nil {
		l.Error("RetrieveQuery is nil")
		return nil, errors.New("RetrieveQuery is nil")
	}
	m, key, err := client.GetWorkCassandra(
		q.Driver.Cassandra.RetrieveQuery.Query,
		q.Driver.Cassandra.RetrieveQuery.Params,
		q.Driver.Cassandra.QueryReturnsKey,
	)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("message received")
	if key != nil {
		q.Driver.Cassandra.Key = key
	}
	if m == nil {
		l.Debug("no work")
		return nil, nil
	}
	return m, nil
}

func (q *ProcX) ClearWorkCassandra() error {
	l := log.WithFields(log.Fields{
		"action": "ClearWorkCassandra",
		"driver": q.DriverName,
	})
	l.Debug("ClearWorkCassandra")
	if q.Driver.Cassandra.ClearQuery == nil {
		l.Debug("no clear query")
		return nil
	}
	err := client.ClearWorkCassandra(
		q.Driver.Cassandra.ClearQuery.Query,
		q.Driver.Cassandra.ClearQuery.Params,
		q.Driver.Cassandra.Key,
	)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}
