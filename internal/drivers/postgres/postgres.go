package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"github.com/robertlestak/procx/internal/flags"
	"github.com/robertlestak/procx/pkg/schema"

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
	RetrieveQuery *schema.SqlQuery
	ClearQuery    *schema.SqlQuery
	FailQuery     *schema.SqlQuery
}

func (d *Postgres) LoadEnv(prefix string) error {

	if os.Getenv(prefix+"PSQL_HOST") != "" {
		d.Host = os.Getenv(prefix + "PSQL_HOST")
	}
	if os.Getenv(prefix+"PSQL_PORT") != "" {
		pval, err := strconv.Atoi(os.Getenv(prefix + "PSQL_PORT"))
		if err != nil {
			return err
		}
		d.Port = pval
	}
	if os.Getenv(prefix+"PSQL_USER") != "" {
		d.User = os.Getenv(prefix + "PSQL_USER")
	}
	if os.Getenv(prefix+"PSQL_PASSWORD") != "" {
		d.Pass = os.Getenv(prefix + "PSQL_PASSWORD")
	}
	if os.Getenv(prefix+"PSQL_DATABASE") != "" {
		d.Db = os.Getenv(prefix + "PSQL_DATABASE")
	}
	if os.Getenv(prefix+"PSQL_SSL_MODE") != "" {
		d.SslMode = os.Getenv(prefix + "PSQL_SSL_MODE")
	}
	if os.Getenv(prefix+"PSQL_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery = &schema.SqlQuery{
			Query: os.Getenv(prefix + "PSQL_RETRIEVE_QUERY"),
		}
	}
	if os.Getenv(prefix+"PSQL_CLEAR_QUERY") != "" {
		d.ClearQuery = &schema.SqlQuery{
			Query: os.Getenv(prefix + "PSQL_CLEAR_QUERY"),
		}
	}
	if os.Getenv(prefix+"PSQL_FAIL_QUERY") != "" {
		d.FailQuery = &schema.SqlQuery{
			Query: os.Getenv(prefix + "PSQL_FAIL_QUERY"),
		}
	}
	if os.Getenv(prefix+"PSQL_RETRIEVE_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"PSQL_RETRIEVE_PARAMS"), ",")
		for _, v := range p {
			d.RetrieveQuery.Params = append(d.RetrieveQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"PSQL_CLEAR_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"PSQL_CLEAR_PARAMS"), ",")
		for _, v := range p {
			d.ClearQuery.Params = append(d.ClearQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"PSQL_FAIL_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"PSQL_FAIL_PARAMS"), ",")
		for _, v := range p {
			d.FailQuery.Params = append(d.FailQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"PSQL_QUERY_KEY") != "" {
		tval := os.Getenv(prefix+"PSQL_QUERY_KEY") == "true"
		d.QueryKey = &tval
	}
	return nil
}

func (d *Postgres) LoadFlags() error {
	pv, err := strconv.Atoi(*flags.PsqlPort)
	if err != nil {
		return err
	}
	var rps []any
	var cps []any
	var fps []any
	if *flags.PsqlRetrieveParams != "" {
		s := strings.Split(*flags.PsqlRetrieveParams, ",")
		for _, v := range s {
			rps = append(rps, v)
		}
	}
	if *flags.PsqlClearParams != "" {
		s := strings.Split(*flags.PsqlClearParams, ",")
		for _, v := range s {
			cps = append(cps, v)
		}
	}
	if *flags.PsqlFailParams != "" {
		s := strings.Split(*flags.PsqlFailParams, ",")
		for _, v := range s {
			fps = append(fps, v)
		}
	}
	d.Host = *flags.PsqlHost
	d.Port = pv
	d.User = *flags.PsqlUser
	d.Pass = *flags.PsqlPassword
	d.Db = *flags.PsqlDatabase
	d.SslMode = *flags.PsqlSSLMode
	if *flags.PsqlQueryKey {
		d.QueryKey = flags.PsqlQueryKey
	}
	if *flags.PsqlRetrieveQuery != "" {
		rq := &schema.SqlQuery{
			Query:  *flags.PsqlRetrieveQuery,
			Params: rps,
		}
		d.RetrieveQuery = rq
	}
	if *flags.PsqlClearQuery != "" {
		cq := &schema.SqlQuery{
			Query:  *flags.PsqlClearQuery,
			Params: cps,
		}
		d.ClearQuery = cq
	}
	if *flags.PsqlFailQuery != "" {
		fq := &schema.SqlQuery{
			Query:  *flags.PsqlFailQuery,
			Params: fps,
		}
		d.FailQuery = fq
	}
	return nil
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
