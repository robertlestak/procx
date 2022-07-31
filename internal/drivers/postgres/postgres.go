package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/robertlestak/procx/pkg/procx"

	log "github.com/sirupsen/logrus"
)

type Postgres struct {
	Client        *sql.DB
	Host          string
	Port          int
	User          string
	Pass          string
	Db            string
	SslMode       string
	Key           *string
	QueryKey      *bool
	RetrieveQuery *procx.SqlQuery
	ClearQuery    *procx.SqlQuery
	FailQuery     *procx.SqlQuery
}

func (d *Postgres) Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreatePsqlClient",
	})
	l.Debug("Initializing psql client")
	var err error
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Pass, d.Host, d.Port, d.Db, d.SslMode)
	d.Client, err = sql.Open("postgres", connStr)
	if err != nil {
		l.Error(err)
		return err
	}
	err = d.Client.Ping()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Connected to psql")
	return nil
}

func (d *Postgres) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWorkPsql",
	})
	l.Debug("Getting work from psql")
	var err error
	var result string
	var key string
	if d.RetrieveQuery == nil || d.RetrieveQuery.Query == "" {
		l.Error("RetrieveQuery is nil or empty")
		return nil, errors.New("RetrieveQuery is nil or empty")
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

func (d *Postgres) ClearWork() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkPsql",
	})
	l.Debug("Clearing work from psql")
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

func (d *Postgres) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "HandleFailurePsql",
	})
	l.Debug("Handling failure in psql")
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
	l.Debug("Handled failure")
	return nil
}
