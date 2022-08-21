package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/schema"
	log "github.com/sirupsen/logrus"
)

type Mysql struct {
	Client        *sql.DB
	Host          string
	Port          int
	User          string
	Pass          string
	Db            string
	RetrieveField *string
	RetrieveQuery *schema.SqlQuery
	ClearQuery    *schema.SqlQuery
	FailQuery     *schema.SqlQuery
	data          map[string]any
}

func (d *Mysql) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "mysql",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment variables")
	if os.Getenv(prefix+"MYSQL_HOST") != "" {
		d.Host = os.Getenv(prefix + "MYSQL_HOST")
	}
	if os.Getenv(prefix+"MYSQL_PORT") != "" {
		pv, err := strconv.Atoi(os.Getenv(prefix + "MYSQL_PORT"))
		if err != nil {
			return err
		}
		d.Port = pv
	}
	if os.Getenv(prefix+"MYSQL_USER") != "" {
		d.User = os.Getenv(prefix + "MYSQL_USER")
	}
	if os.Getenv(prefix+"MYSQL_PASSWORD") != "" {
		d.Pass = os.Getenv(prefix + "MYSQL_PASSWORD")
	}
	if os.Getenv(prefix+"MYSQL_DATABASE") != "" {
		d.Db = os.Getenv(prefix + "MYSQL_DATABASE")
	}
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"MYSQL_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery.Query = os.Getenv(prefix + "MYSQL_RETRIEVE_QUERY")
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"MYSQL_CLEAR_QUERY") != "" {
		d.ClearQuery.Query = os.Getenv(prefix + "MYSQL_CLEAR_QUERY")
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"MYSQL_FAIL_QUERY") != "" {
		d.FailQuery.Query = os.Getenv(prefix + "MYSQL_FAIL_QUERY")
	}
	if os.Getenv(prefix+"MYSQL_RETRIEVE_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"MYSQL_RETRIEVE_PARAMS"), ",")
		for _, v := range p {
			d.RetrieveQuery.Params = append(d.RetrieveQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"MYSQL_CLEAR_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"MYSQL_CLEAR_PARAMS"), ",")
		for _, v := range p {
			d.ClearQuery.Params = append(d.ClearQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"MYSQL_FAIL_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"MYSQL_FAIL_PARAMS"), ",")
		for _, v := range p {
			d.FailQuery.Params = append(d.FailQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"MYSQL_RETRIEVE_FIELD") != "" {
		v := os.Getenv(prefix + "MYSQL_RETRIEVE_FIELD")
		d.RetrieveField = &v
	}
	return nil
}

func (d *Mysql) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "mysql",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	pv, err := strconv.Atoi(*flags.MysqlPort)
	if err != nil {
		return err
	}
	var rps []any
	var cps []any
	var fps []any
	if *flags.MysqlRetrieveParams != "" {
		s := strings.Split(*flags.MysqlRetrieveParams, ",")
		for _, v := range s {
			rps = append(rps, v)
		}
	}
	if *flags.MysqlClearParams != "" {
		s := strings.Split(*flags.MysqlClearParams, ",")
		for _, v := range s {
			cps = append(cps, v)
		}
	}
	if *flags.MysqlFailParams != "" {
		s := strings.Split(*flags.MysqlFailParams, ",")
		for _, v := range s {
			fps = append(fps, v)
		}
	}
	d.Host = *flags.MysqlHost
	d.Port = pv
	d.User = *flags.MysqlUser
	d.Pass = *flags.MysqlPassword
	d.Db = *flags.MysqlDatabase
	d.RetrieveField = flags.MysqlRetrieveField
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if *flags.MysqlRetrieveQuery != "" {
		d.RetrieveQuery.Query = *flags.MysqlRetrieveQuery
	}
	if len(rps) > 0 {
		d.RetrieveQuery.Params = rps
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if *flags.MysqlClearQuery != "" {
		d.ClearQuery.Query = *flags.MysqlClearQuery
	}
	if len(cps) > 0 {
		d.ClearQuery.Params = cps
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if *flags.MysqlFailQuery != "" {
		d.FailQuery.Query = *flags.MysqlFailQuery
	}
	if len(fps) > 0 {
		d.FailQuery.Params = fps
	}
	return nil
}

func (d *Mysql) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "mysql",
		"fn":  "Init",
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

func (d *Mysql) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "mysql",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from mysql")
	var result string
	if d.RetrieveQuery == nil || d.RetrieveQuery.Query == "" {
		l.Error("query is empty")
		return nil, errors.New("query is empty")
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
	// if result is empty, return nil
	if result == "" {
		l.Debug("result is empty")
		return nil, nil
	}
	l.Debug("Got work")
	return strings.NewReader(result), nil
}

func (d *Mysql) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "mysql",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from mysql")
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

func (d *Mysql) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "mysql",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failed work from mysql")
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
	l.Debug("Cleared work")
	return nil
}

func (d *Mysql) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "mysql",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up mysql client")
	err := d.Client.Close()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleaned up mysql client")
	return nil
}
