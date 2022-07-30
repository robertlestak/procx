package client

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

var (
	MysqlClient *sql.DB
)

func CreateMySqlClient(host string, port int, user string, pass string, db string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreateMySqlClient",
	})
	l.Debug("Initializing mysql client")
	var err error
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pass, host, port, db)
	l.Debug("Connecting to mysql: ", connStr)
	MysqlClient, err = sql.Open("mysql", connStr)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Initialized mysql client")
	// ping the database to check if it is alive
	err = MysqlClient.Ping()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Pinged mysql client")
	return nil
}

func GetWorkMysql(query string, params []any, queryKey *bool) (*string, *string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWorkMysql",
	})
	l.Debug("Getting work from mysql")
	var err error
	var result string
	var key string
	if queryKey != nil && *queryKey {
		err = MysqlClient.QueryRow(query, params...).Scan(&key, &result)
	} else {
		err = MysqlClient.QueryRow(query, params...).Scan(&result)
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

func ClearWorkMysql(query string, params []any, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkMysql",
	})
	l.Debug("Clearing work from mysql")
	var err error
	if key != nil && *key != "" {
		// loop through params and if we find {{key}}, replace it with the key
		for i, v := range params {
			if v == "{{key}}" {
				params[i] = *key
			}
		}
	}
	_, err = MysqlClient.Exec(query, params...)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func HandleFailureMysql(query string, params []any, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "HandleFailureMysql",
	})
	l.Debug("handling failure for mysql")
	var err error
	if key != nil && *key != "" {
		// loop through params and if we find {{key}}, replace it with the key
		for i, v := range params {
			if v == "{{key}}" {
				params[i] = *key
			}
		}
	}
	_, err = MysqlClient.Exec(query, params...)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("handled failure")
	return nil
}
