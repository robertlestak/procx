package mssql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type MSSql struct {
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
	data          []map[string]any
}

func (d *MSSql) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "mssql",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment variables")
	if os.Getenv(prefix+"MSSQL_HOST") != "" {
		d.Host = os.Getenv(prefix + "MSSQL_HOST")
	}
	if os.Getenv(prefix+"MSSQL_PORT") != "" {
		pv, err := strconv.Atoi(os.Getenv(prefix + "MSSQL_PORT"))
		if err != nil {
			return err
		}
		d.Port = pv
	}
	if os.Getenv(prefix+"MSSQL_USER") != "" {
		d.User = os.Getenv(prefix + "MSSQL_USER")
	}
	if os.Getenv(prefix+"MSSQL_PASSWORD") != "" {
		d.Pass = os.Getenv(prefix + "MSSQL_PASSWORD")
	}
	if os.Getenv(prefix+"MSSQL_DATABASE") != "" {
		d.Db = os.Getenv(prefix + "MSSQL_DATABASE")
	}
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"MSSQL_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery.Query = os.Getenv(prefix + "MSSQL_RETRIEVE_QUERY")
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"MSSQL_CLEAR_QUERY") != "" {
		d.ClearQuery.Query = os.Getenv(prefix + "MSSQL_CLEAR_QUERY")
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"MSSQL_FAIL_QUERY") != "" {
		d.FailQuery.Query = os.Getenv(prefix + "MSSQL_FAIL_QUERY")
	}
	if os.Getenv(prefix+"MSSQL_RETRIEVE_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"MSSQL_RETRIEVE_PARAMS"), ",")
		for _, v := range p {
			d.RetrieveQuery.Params = append(d.RetrieveQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"MSSQL_CLEAR_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"MSSQL_CLEAR_PARAMS"), ",")
		for _, v := range p {
			d.ClearQuery.Params = append(d.ClearQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"MSSQL_FAIL_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"MSSQL_FAIL_PARAMS"), ",")
		for _, v := range p {
			d.FailQuery.Params = append(d.FailQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"MSSQL_RETRIEVE_FIELD") != "" {
		v := os.Getenv(prefix + "MSSQL_RETRIEVE_FIELD")
		d.RetrieveField = &v
	}
	return nil
}

func (d *MSSql) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "mssql",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	pv, err := strconv.Atoi(*flags.MSSqlPort)
	if err != nil {
		return err
	}
	var rps []any
	var cps []any
	var fps []any
	if *flags.MSSqlRetrieveParams != "" {
		s := strings.Split(*flags.MSSqlRetrieveParams, ",")
		for _, v := range s {
			rps = append(rps, v)
		}
	}
	if *flags.MSSqlClearParams != "" {
		s := strings.Split(*flags.MSSqlClearParams, ",")
		for _, v := range s {
			cps = append(cps, v)
		}
	}
	if *flags.MSSqlFailParams != "" {
		s := strings.Split(*flags.MSSqlFailParams, ",")
		for _, v := range s {
			fps = append(fps, v)
		}
	}
	d.Host = *flags.MSSqlHost
	d.Port = pv
	d.User = *flags.MSSqlUser
	d.Pass = *flags.MSSqlPassword
	d.Db = *flags.MSSqlDatabase
	d.RetrieveField = flags.MSSqlRetrieveField
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if *flags.MSSqlRetrieveQuery != "" {
		d.RetrieveQuery.Query = *flags.MSSqlRetrieveQuery
	}
	if len(rps) > 0 {
		d.RetrieveQuery.Params = rps
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if *flags.MSSqlClearQuery != "" {
		d.ClearQuery.Query = *flags.MSSqlClearQuery
	}
	if len(cps) > 0 {
		d.ClearQuery.Params = cps
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if *flags.MSSqlFailQuery != "" {
		d.FailQuery.Query = *flags.MSSqlFailQuery
	}
	if len(fps) > 0 {
		d.FailQuery.Params = fps
	}
	return nil
}

func (d *MSSql) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "mssql",
		"fn":  "Init",
	})
	l.Debug("Initializing mssql client")
	var err error
	connStr := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s", d.Host, d.User, d.Pass, d.Port, d.Db)
	l.Debug("Connecting to mssql: ", connStr)
	d.Client, err = sql.Open("mssql", connStr)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Initialized mssql client")
	// ping the database to check if it is alive
	err = d.Client.Ping()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Pinged mssql client")
	return nil
}

func (d *MSSql) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "mssql",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from mssql")
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
	m, err := schema.RowsToMapSlice(r)
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
		bd, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		jv := gjson.GetBytes(bd, *d.RetrieveField)
		result = fmt.Sprintf("%s", jv.Value())
	} else {
		jd, err := schema.SliceMapStringAnyToJSON(m)
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

func (d *MSSql) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "mssql",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from mssql")
	var err error
	if d.ClearQuery == nil || d.ClearQuery.Query == "" {
		return nil
	}
	d.ClearQuery.Params = schema.ReplaceParamsSliceMap(d.data, d.ClearQuery.Params)
	_, err = d.Client.Exec(d.ClearQuery.Query, d.ClearQuery.Params...)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func (d *MSSql) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "mssql",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failed work from mssql")
	var err error
	if d.FailQuery == nil || d.FailQuery.Query == "" {
		return nil
	}
	d.FailQuery.Params = schema.ReplaceParamsSliceMap(d.data, d.FailQuery.Params)
	_, err = d.Client.Exec(d.FailQuery.Query, d.FailQuery.Params...)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func (d *MSSql) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "mssql",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up mssql client")
	err := d.Client.Close()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleaned up mssql client")
	return nil
}
