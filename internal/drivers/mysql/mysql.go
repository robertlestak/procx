package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robertlestak/procx/pkg/procx"
	log "github.com/sirupsen/logrus"
)

type Mysql struct {
	Client        *sql.DB
	Host          string
	Port          int
	User          string
	Pass          string
	Db            string
	Key           *string
	QueryKey      *bool
	RetrieveQuery *procx.SqlQuery
	ClearQuery    *procx.SqlQuery
	FailQuery     *procx.SqlQuery
}

func (d *Mysql) Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreateMySqlClient",
	})
	l.Debug("Initializing mysql client")
	var err error
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", d.User, d.Pass, d.Host, d.Port, d.Db)
	l.Debug("Connecting to mysql: ", connStr)
	d.Client, err = sql.Open("mysql", connStr)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Initialized mysql client")
	// ping the database to check if it is alive
	err = d.Client.Ping()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Pinged mysql client")
	return nil
}

func (d *Mysql) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWorkMysql",
	})
	l.Debug("Getting work from mysql")
	var err error
	var result string
	var key string
	if d.RetrieveQuery == nil || d.RetrieveQuery.Query == "" {
		l.Error("query is empty")
		return nil, errors.New("query is empty")
	}
	if d.QueryKey != nil && *d.QueryKey {
		err = d.Client.QueryRow(d.RetrieveQuery.Query, d.RetrieveQuery.Params...).Scan(&key, &result)
	} else {
		err = d.Client.QueryRow(d.RetrieveQuery.Query, d.RetrieveQuery.Params...).Scan(&result)
	}
	if err != nil {
		// if the queue is empty, return nil
		if err == sql.ErrNoRows {
			l.Debug("Queue is empty")
			return nil, nil
		}
		l.Error(err)
		return nil, err
	}
	d.Key = &key
	l.Debug("Got work")
	return &result, nil
}

func (d *Mysql) ClearWork() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkMysql",
	})
	l.Debug("Clearing work from mysql")
	var err error
	if d.ClearQuery == nil || d.ClearQuery.Query == "" {
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
	_, err = d.Client.Exec(d.ClearQuery.Query, d.ClearQuery.Params...)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func (d *Mysql) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "HandleFailure",
	})
	l.Debug("Handling failed work from mysql")
	var err error
	if d.FailQuery == nil || d.FailQuery.Query == "" {
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
	_, err = d.Client.Exec(d.FailQuery.Query, d.FailQuery.Params...)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}
