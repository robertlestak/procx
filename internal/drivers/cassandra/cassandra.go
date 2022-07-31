package cassandra

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/robertlestak/procx/pkg/procx"
	log "github.com/sirupsen/logrus"
)

type Cassandra struct {
	Client        *gocql.Session
	Hosts         []string
	User          string
	Password      string
	Consistency   string
	Keyspace      string
	QueryKey      *bool
	Key           *string
	RetrieveQuery *procx.SqlQuery
	ClearQuery    *procx.SqlQuery
	FailQuery     *procx.SqlQuery
}

func (d *Cassandra) Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreateCassandraClient",
	})
	l.Debug("Initializing cassandra client")

	cluster := gocql.NewCluster(d.Hosts...)
	// parse consistency string
	consistencyLevel := gocql.ParseConsistency(d.Consistency)
	cluster.Consistency = consistencyLevel
	if d.Keyspace != "" {
		cluster.Keyspace = d.Keyspace
	}
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = time.Second * 10
	if d.User != "" || d.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{Username: d.User, Password: d.Password}
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	d.Client = session
	return nil
}

func (d *Cassandra) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWork",
	})
	l.Debug("Getting work from cassandra")
	var err error
	var result string
	var key string
	if d.QueryKey != nil && *d.QueryKey {
		err = d.Client.Query(d.RetrieveQuery.Query).Scan(&key, &result)
	} else {
		err = d.Client.Query(d.RetrieveQuery.Query).Scan(&result)
	}
	if err != nil {
		// if the queue is empty, return nil
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	l.Debug("Got work")
	d.Key = &key
	return &result, nil
}

func (d *Cassandra) ClearWork() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkCassandra",
	})
	l.Debug("Clearing work from cassandra")
	var err error
	if d.ClearQuery == nil {
		return nil
	}
	if d.ClearQuery.Query == "" {
		return nil
	}
	if d.Key != nil && *d.Key != "" {
		// loop through params and if we find {{key}}, replace it with the key
		for i, v := range d.ClearQuery.Params {
			if v == "{{key}}" {
				d.ClearQuery.Params[i] = *d.Key
			}
		}
	}
	err = d.Client.Query(d.ClearQuery.Query, d.ClearQuery.Params...).Exec()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func (d *Cassandra) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "HandleFailureCassandra",
	})
	l.Debug("handling failure for cassandra")
	var err error
	if d.FailQuery == nil {
		return nil
	}
	if d.FailQuery.Query == "" {
		return nil
	}
	if d.Key != nil && *d.Key != "" {
		// loop through params and if we find {{key}}, replace it with the key
		for i, v := range d.FailQuery.Params {
			if v == "{{key}}" {
				d.FailQuery.Params[i] = *d.Key
			}
		}
	}
	err = d.Client.Query(d.FailQuery.Query, d.FailQuery.Params...).Exec()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("handled failure")
	return nil
}
