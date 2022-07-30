package client

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
)

var (
	PsqlClient *sql.DB
)

func CreatePsqlClient(host string, port int, user, pass, db, sslMode string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreatePsqlClient",
	})
	l.Debug("Initializing psql client")
	var err error
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", user, pass, host, port, db, sslMode)
	PsqlClient, err = sql.Open("postgres", connStr)
	if err != nil {
		l.Error(err)
		return err
	}
	err = PsqlClient.Ping()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Connected to psql")
	return nil
}

func GetWorkPsql(query string, params []any, queryKey *bool) (*string, *string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWorkPsql",
	})
	l.Debug("Getting work from psql")
	var err error
	var result string
	var key string
	if queryKey != nil && *queryKey {
		err = PsqlClient.QueryRow(query, params...).Scan(&key, &result)
	} else {
		err = PsqlClient.QueryRow(query, params...).Scan(&result)
	}
	if err != nil {
		// if the queue is empty, return nil
		if err == sql.ErrNoRows {
			l.Debug("Queue is empty")
			return nil, nil, nil
		}
		l.Error(err)
		return nil, nil, err
	}
	l.Debug("Got work")
	return &result, &key, nil
}

func ClearWorkPsql(query string, params []any, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkPsql",
	})
	l.Debug("Clearing work from psql")
	var err error
	if key != nil && *key != "" {
		// loop through params and if we find {{key}}, replace it with the key
		for i, v := range params {
			if v == "{{key}}" {
				params[i] = *key
			}
		}
	}
	_, err = PsqlClient.Exec(query, params...)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func HandleFailurePsql(query string, params []any, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "HandleFailurePsql",
	})
	l.Debug("handling failure for psql")
	var err error
	if key != nil && *key != "" {
		// loop through params and if we find {{key}}, replace it with the key
		for i, v := range params {
			if v == "{{key}}" {
				params[i] = *key
			}
		}
	}
	_, err = PsqlClient.Exec(query, params...)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("handled failure")
	return nil
}
