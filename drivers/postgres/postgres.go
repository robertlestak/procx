package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"github.com/robertlestak/procx/pkg/flags"
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
	RetrieveField *string
	RetrieveQuery *schema.SqlQuery
	ClearQuery    *schema.SqlQuery
	FailQuery     *schema.SqlQuery
	data          map[string]any
}

func (d *Postgres) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "postgres",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment variables")
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
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"PSQL_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery.Query = os.Getenv(prefix + "PSQL_RETRIEVE_QUERY")
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"PSQL_CLEAR_QUERY") != "" {
		d.ClearQuery.Query = os.Getenv(prefix + "PSQL_CLEAR_QUERY")
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"PSQL_FAIL_QUERY") != "" {
		d.FailQuery.Query = os.Getenv(prefix + "PSQL_FAIL_QUERY")
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
	if os.Getenv(prefix+"PSQL_RETRIEVE_FIELD") != "" {
		v := os.Getenv(prefix + "PSQL_RETRIEVE_FIELD")
		d.RetrieveField = &v
	}
	return nil
}

func (d *Postgres) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "postgres",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
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
	d.RetrieveField = flags.PsqlRetrieveField
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if *flags.PsqlRetrieveQuery != "" {
		d.RetrieveQuery.Query = *flags.PsqlRetrieveQuery
	}
	if len(rps) > 0 {
		d.RetrieveQuery.Params = rps
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if *flags.PsqlClearQuery != "" {
		d.ClearQuery.Query = *flags.PsqlClearQuery
	}
	if len(cps) > 0 {
		d.ClearQuery.Params = cps
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if *flags.PsqlFailQuery != "" {
		d.FailQuery.Query = *flags.PsqlFailQuery
	}
	if len(fps) > 0 {
		d.FailQuery.Params = fps
	}
	return nil
}

func (d *Postgres) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "postgres",
		"fn":  "Init",
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

func (d *Postgres) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "postgres",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from psql")
	var result string
	if d.RetrieveQuery == nil || d.RetrieveQuery.Query == "" {
		l.Error("RetrieveQuery is nil or empty")
		return nil, errors.New("RetrieveQuery is nil or empty")
	}
	r, err := d.Client.Query(d.RetrieveQuery.Query, d.RetrieveQuery.Params...)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	if r.Err() != nil {
		l.Error(r.Err())
		return nil, r.Err()
	}
	m, err := schema.RowsToMap(r)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	if len(m) == 0 {
		l.Debug("No work found")
		return nil, nil
	}
	d.data = m
	if d.RetrieveField != nil && *d.RetrieveField != "" {
		result = fmt.Sprintf("%s", schema.HandleField(m[*d.RetrieveField]))
	} else {
		jd, err := schema.MapStringAnyToJSON(m)
		if err != nil {
			l.Error(err)
			return nil, err
		}
		result = string(jd)
	}
	l.Debug("Got work")
	// if result is empty, return nil
	if result == "" {
		l.Debug("result is empty")
		return nil, nil
	}
	return strings.NewReader(result), nil
}

func (d *Postgres) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "postgres",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from psql")
	var err error
	if d.ClearQuery == nil || d.ClearQuery.Query == "" {
		return nil
	}
	d.ClearQuery.Params = schema.ReplaceParamsMap(d.data, d.ClearQuery.Params)
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
		"pkg": "postgres",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure in psql")
	var err error
	if d.FailQuery == nil || d.FailQuery.Query == "" {
		return nil
	}
	d.FailQuery.Params = schema.ReplaceParamsMap(d.data, d.FailQuery.Params)
	_, err = d.Client.Exec(d.FailQuery.Query, d.FailQuery.Params...)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Handled failure")
	return nil
}

func (d *Postgres) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "postgres",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up psql")
	err := d.Client.Close()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleaned up")
	return nil
}
