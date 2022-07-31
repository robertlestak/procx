package client

import (
	"time"

	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"
)

var (
	CassandraClient *gocql.Session
)

func CreateCassandraClient(hosts []string, user string, pass string, consistency string, keyspace string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreateCassandraClient",
	})
	l.Debug("Initializing cassandra client")

	cluster := gocql.NewCluster(hosts...)
	// parse consistency string
	consistencyLevel := gocql.ParseConsistency(consistency)
	cluster.Consistency = consistencyLevel
	if keyspace != "" {
		cluster.Keyspace = keyspace
	}
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = time.Second * 10
	if user != "" || pass != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{Username: user, Password: pass}
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	CassandraClient = session
	return nil
}

func GetWorkCassandra(query string, params []any, queryKey *bool) (*string, *string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWorkCassandra",
	})
	l.Debug("Getting work from cassandra")
	var err error
	var result string
	var key string
	if queryKey != nil && *queryKey {
		err = CassandraClient.Query(query).Scan(&key, &result)
	} else {
		err = CassandraClient.Query(query).Scan(&result)
	}
	if err != nil {
		// if the queue is empty, return nil
		if err == gocql.ErrNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	l.Debug("Got work")
	return &result, &key, nil
}

func ClearWorkCassandra(query string, params []any, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkCassandra",
	})
	l.Debug("Clearing work from cassandra")
	var err error
	if key != nil && *key != "" {
		// loop through params and if we find {{key}}, replace it with the key
		for i, v := range params {
			if v == "{{key}}" {
				params[i] = *key
			}
		}
	}
	err = CassandraClient.Query(query, params...).Exec()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func HandleFailureCassandra(query string, params []any, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "HandleFailureCassandra",
	})
	l.Debug("handling failure for cassandra")
	var err error
	if key != nil && *key != "" {
		// loop through params and if we find {{key}}, replace it with the key
		for i, v := range params {
			if v == "{{key}}" {
				params[i] = *key
			}
		}
	}
	err = CassandraClient.Query(query, params...).Exec()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("handled failure")
	return nil
}
