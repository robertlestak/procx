package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robertlestak/procx/internal/flags"
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
	Key           *string
	QueryKey      *bool
	RetrieveQuery *schema.SqlQuery
	ClearQuery    *schema.SqlQuery
	FailQuery     *schema.SqlQuery
}

func (d *Mysql) LoadEnv(prefix string) error {
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
	if os.Getenv(prefix+"MYSQL_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery = &schema.SqlQuery{
			Query: os.Getenv(prefix + "MYSQL_RETRIEVE_QUERY"),
		}
	}
	if os.Getenv(prefix+"MYSQL_CLEAR_QUERY") != "" {
		d.ClearQuery = &schema.SqlQuery{
			Query: os.Getenv(prefix + "MYSQL_CLEAR_QUERY"),
		}
	}
	if os.Getenv(prefix+"MYSQL_FAIL_QUERY") != "" {
		d.FailQuery = &schema.SqlQuery{
			Query: os.Getenv(prefix + "MYSQL_FAIL_QUERY"),
		}
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
	if os.Getenv(prefix+"MYSQL_QUERY_KEY") != "" {
		v := os.Getenv(prefix+"MYSQL_QUERY_KEY") == "true"
		d.QueryKey = &v
	}
	return nil
}

func (d *Mysql) LoadFlags() error {
	pv, err := strconv.Atoi(*flags.FlagMysqlPort)
	if err != nil {
		return err
	}
	var rps []any
	var cps []any
	var fps []any
	if *flags.FlagMysqlRetrieveParams != "" {
		s := strings.Split(*flags.FlagMysqlRetrieveParams, ",")
		for _, v := range s {
			rps = append(rps, v)
		}
	}
	if *flags.FlagMysqlClearParams != "" {
		s := strings.Split(*flags.FlagMysqlClearParams, ",")
		for _, v := range s {
			cps = append(cps, v)
		}
	}
	if *flags.FlagMysqlFailParams != "" {
		s := strings.Split(*flags.FlagMysqlFailParams, ",")
		for _, v := range s {
			fps = append(fps, v)
		}
	}
	d.Host = *flags.FlagMysqlHost
	d.Port = pv
	d.User = *flags.FlagMysqlUser
	d.Pass = *flags.FlagMysqlPassword
	d.Db = *flags.FlagMysqlDatabase
	if *flags.FlagMysqlQueryKey {
		d.QueryKey = flags.FlagMysqlQueryKey
	}
	if *flags.FlagMysqlRetrieveQuery != "" {
		rq := &schema.SqlQuery{
			Query:  *flags.FlagMysqlRetrieveQuery,
			Params: rps,
		}
		d.RetrieveQuery = rq
	}
	if *flags.FlagMysqlClearQuery != "" {
		cq := &schema.SqlQuery{
			Query:  *flags.FlagMysqlClearQuery,
			Params: cps,
		}
		d.ClearQuery = cq
	}
	if *flags.FlagMysqlFailQuery != "" {
		fq := &schema.SqlQuery{
			Query:  *flags.FlagMysqlFailQuery,
			Params: fps,
		}
		d.FailQuery = fq
	}
	return nil
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
