package cockroach

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

type CockroachDB struct {
	Client        *sql.DB
	Host          string
	Port          int
	User          string
	Pass          string
	Db            string
	SslMode       string
	SSLRootCert   *string
	SSLCert       *string
	SSLKey        *string
	RetrieveField *string
	RoutingID     *string
	RetrieveQuery *schema.SqlQuery
	ClearQuery    *schema.SqlQuery
	FailQuery     *schema.SqlQuery
	data          map[string]any
}

func (d *CockroachDB) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "cockroach",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment variables")
	if os.Getenv(prefix+"COCKROACH_HOST") != "" {
		d.Host = os.Getenv(prefix + "COCKROACH_HOST")
	}
	if os.Getenv(prefix+"COCKROACH_PORT") != "" {
		pval, err := strconv.Atoi(os.Getenv(prefix + "COCKROACH_PORT"))
		if err != nil {
			return err
		}
		d.Port = pval
	}
	if os.Getenv(prefix+"COCKROACH_USER") != "" {
		d.User = os.Getenv(prefix + "COCKROACH_USER")
	}
	if os.Getenv(prefix+"COCKROACH_PASSWORD") != "" {
		d.Pass = os.Getenv(prefix + "COCKROACH_PASSWORD")
	}
	if os.Getenv(prefix+"COCKROACH_DATABASE") != "" {
		d.Db = os.Getenv(prefix + "COCKROACH_DATABASE")
	}
	if os.Getenv(prefix+"COCKROACH_SSL_MODE") != "" {
		d.SslMode = os.Getenv(prefix + "COCKROACH_SSL_MODE")
	}
	if os.Getenv(prefix+"COCKROACH_TLS_ROOT_CERT") != "" {
		v := os.Getenv(prefix + "COCKROACH_TLS_ROOT_CERT")
		d.SSLRootCert = &v
	}
	if os.Getenv(prefix+"COCKROACH_TLS_CERT") != "" {
		v := os.Getenv(prefix + "COCKROACH_TLS_CERT")
		d.SSLCert = &v
	}
	if os.Getenv(prefix+"COCKROACH_TLS_KEY") != "" {
		v := os.Getenv(prefix + "COCKROACH_TLS_KEY")
		d.SSLKey = &v
	}
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"COCKROACH_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery.Query = os.Getenv(prefix + "COCKROACH_RETRIEVE_QUERY")
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"COCKROACH_CLEAR_QUERY") != "" {
		d.ClearQuery.Query = os.Getenv(prefix + "COCKROACH_CLEAR_QUERY")
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"COCKROACH_FAIL_QUERY") != "" {
		d.FailQuery.Query = os.Getenv(prefix + "COCKROACH_FAIL_QUERY")
	}
	if os.Getenv(prefix+"COCKROACH_RETRIEVE_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"COCKROACH_RETRIEVE_PARAMS"), ",")
		for _, v := range p {
			d.RetrieveQuery.Params = append(d.RetrieveQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"COCKROACH_CLEAR_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"COCKROACH_CLEAR_PARAMS"), ",")
		for _, v := range p {
			d.ClearQuery.Params = append(d.ClearQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"COCKROACH_FAIL_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"COCKROACH_FAIL_PARAMS"), ",")
		for _, v := range p {
			d.FailQuery.Params = append(d.FailQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"COCKROACH_RETRIEVE_FIELD") != "" {
		v := os.Getenv(prefix + "COCKROACH_RETRIEVE_FIELD")
		d.RetrieveField = &v
	}
	if os.Getenv(prefix+"COCKROACH_ROUTING_ID") != "" {
		v := os.Getenv(prefix + "COCKROACH_ROUTING_ID")
		d.RoutingID = &v
	}
	return nil
}

func (d *CockroachDB) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "cockroach",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	pv, err := strconv.Atoi(*flags.CockroachDBPort)
	if err != nil {
		return err
	}
	var rps []any
	var cps []any
	var fps []any
	if *flags.CockroachDBRetrieveParams != "" {
		s := strings.Split(*flags.CockroachDBRetrieveParams, ",")
		for _, v := range s {
			rps = append(rps, v)
		}
	}
	if *flags.CockroachDBClearParams != "" {
		s := strings.Split(*flags.CockroachDBClearParams, ",")
		for _, v := range s {
			cps = append(cps, v)
		}
	}
	if *flags.CockroachDBFailParams != "" {
		s := strings.Split(*flags.CockroachDBFailParams, ",")
		for _, v := range s {
			fps = append(fps, v)
		}
	}
	d.Host = *flags.CockroachDBHost
	d.Port = pv
	d.User = *flags.CockroachDBUser
	d.Pass = *flags.CockroachDBPassword
	d.Db = *flags.CockroachDBDatabase
	d.SslMode = *flags.CockroachDBSSLMode
	d.SSLRootCert = flags.CockroachDBTLSRootCert
	d.SSLCert = flags.CockroachDBTLSCert
	d.SSLKey = flags.CockroachDBTLSKey
	d.RetrieveField = flags.CockroachDBRetrieveField
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if *flags.CockroachDBRetrieveQuery != "" {
		d.RetrieveQuery.Query = *flags.CockroachDBRetrieveQuery
	}
	if len(rps) > 0 {
		d.RetrieveQuery.Params = rps
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if *flags.CockroachDBClearQuery != "" {
		d.ClearQuery.Query = *flags.CockroachDBClearQuery
	}
	if len(cps) > 0 {
		d.ClearQuery.Params = cps
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if *flags.CockroachDBFailQuery != "" {
		d.FailQuery.Query = *flags.CockroachDBFailQuery
	}
	if len(fps) > 0 {
		d.FailQuery.Params = fps
	}
	return nil
}

func (d *CockroachDB) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "cockroach",
		"fn":  "Init",
	})
	l.Debug("Initializing cockroachdb client")
	var err error
	var opts string
	var connStr string = "postgresql://"
	if d.RoutingID != nil && *d.RoutingID != "" {
		opts = "&options=--cluster%3D" + *d.RoutingID
	}
	if d.User != "" && d.Pass != "" {
		connStr += fmt.Sprintf("%s:%s@%s:%d/%s",
			d.User, d.Pass, d.Host, d.Port, d.Db)
	} else if d.User != "" && d.Pass == "" {
		connStr += fmt.Sprintf("%s@%s:%d/%s",
			d.User, d.Host, d.Port, d.Db)
	}
	connStr += "?sslmode=" + d.SslMode
	if d.SSLRootCert != nil && *d.SSLRootCert != "" {
		connStr += "&sslrootcert=" + *d.SSLRootCert
	}
	if d.SSLCert != nil && *d.SSLCert != "" {
		connStr += "&sslcert=" + *d.SSLCert
	}
	if d.SSLKey != nil && *d.SSLKey != "" {
		connStr += "&sslkey=" + *d.SSLKey
	}
	connStr += opts
	l.Debugf("Connecting to %s", connStr)
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
	l.Debug("Connected to cockroachdb")
	return nil
}

func (d *CockroachDB) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "cockroach",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from cockroachdb")
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

func (d *CockroachDB) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "cockroach",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from cockroachdb")
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

func (d *CockroachDB) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "cockroach",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure in cockroachdb")
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

func (d *CockroachDB) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "cockroach",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up cockroachdb")
	err := d.Client.Close()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleaned up")
	return nil
}
