package cassandra

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Cassandra struct {
	Client        *gocql.Session
	Hosts         []string
	User          string
	Password      string
	Consistency   string
	Keyspace      string
	RetrieveField *string
	RetrieveQuery *schema.SqlQuery
	ClearQuery    *schema.SqlQuery
	FailQuery     *schema.SqlQuery
	data          []map[string]any
}

func (d *Cassandra) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "cassandra",
		"fn":  "LoadEnv",
	})
	l.Debug("loading env")
	if os.Getenv(prefix+"CASSANDRA_HOSTS") != "" {
		d.Hosts = strings.Split(os.Getenv(prefix+"CASSANDRA_HOSTS"), ",")
	}
	if os.Getenv(prefix+"CASSANDRA_KEYSPACE") != "" {
		d.Keyspace = os.Getenv(prefix + "CASSANDRA_KEYSPACE")
	}
	if os.Getenv(prefix+"CASSANDRA_USER") != "" {
		d.User = os.Getenv(prefix + "CASSANDRA_USER")
	}
	if os.Getenv(prefix+"CASSANDRA_PASSWORD") != "" {
		d.Password = os.Getenv(prefix + "CASSANDRA_PASSWORD")
	}
	if os.Getenv(prefix+"CASSANDRA_CONSISTENCY") != "" {
		d.Consistency = os.Getenv(prefix + "CASSANDRA_CONSISTENCY")
	}
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"CASSANDRA_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery.Query = os.Getenv(prefix + "CASSANDRA_RETRIEVE_QUERY")
	}
	if os.Getenv(prefix+"CASSANDRA_RETRIEVE_PARAMS") != "" {
		for _, s := range strings.Split(os.Getenv(prefix+"CASSANDRA_RETRIEVE_PARAMS"), ",") {
			d.RetrieveQuery.Params = append(d.RetrieveQuery.Params, s)
		}
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"CASSANDRA_CLEAR_QUERY") != "" {
		d.ClearQuery.Query = os.Getenv(prefix + "CASSANDRA_CLEAR_QUERY")
	}
	if os.Getenv(prefix+"CASSANDRA_CLEAR_PARAMS") != "" {
		for _, s := range strings.Split(os.Getenv(prefix+"CASSANDRA_CLEAR_PARAMS"), ",") {
			d.ClearQuery.Params = append(d.ClearQuery.Params, s)
		}
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"CASSANDRA_FAIL_QUERY") != "" {
		d.FailQuery.Query = os.Getenv(prefix + "CASSANDRA_FAIL_QUERY")
	}
	if os.Getenv(prefix+"CASSANDRA_FAIL_PARAMS") != "" {
		for _, s := range strings.Split(os.Getenv(prefix+"CASSANDRA_FAIL_PARAMS"), ",") {
			d.FailQuery.Params = append(d.FailQuery.Params, s)
		}
	}
	if os.Getenv(prefix+"CASSANDRA_RETRIEVE_FIELD") != "" {
		v := os.Getenv(prefix + "CASSANDRA_RETRIEVE_FIELD")
		d.RetrieveField = &v
	}
	return nil
}

func (d *Cassandra) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "cassandra",
		"fn":  "LoadFlags",
	})
	l.Debug("loading flags")
	var hosts []string
	if *flags.CassandraHosts != "" {
		s := strings.Split(*flags.CassandraHosts, ",")
		for _, v := range s {
			v = strings.TrimSpace(v)
			if v != "" {
				hosts = append(hosts, v)
			}
		}
	}
	var rps []any
	var cps []any
	var fps []any
	if *flags.CassandraRetrieveParams != "" {
		s := strings.Split(*flags.CassandraRetrieveParams, ",")
		for _, v := range s {
			rps = append(rps, v)
		}
	}
	if *flags.CassandraClearParams != "" {
		s := strings.Split(*flags.CassandraClearParams, ",")
		for _, v := range s {
			cps = append(cps, v)
		}
	}
	if *flags.CassandraFailParams != "" {
		s := strings.Split(*flags.CassandraFailParams, ",")
		for _, v := range s {
			fps = append(fps, v)
		}
	}
	d.Hosts = hosts
	d.User = *flags.CassandraUser
	d.Password = *flags.CassandraPassword
	d.Keyspace = *flags.CassandraKeyspace
	d.Consistency = *flags.CassandraConsistency
	d.RetrieveField = flags.CassandraRetrieveField
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if *flags.CassandraRetrieveQuery != "" {
		d.RetrieveQuery.Query = *flags.CassandraRetrieveQuery
	}
	if len(rps) > 0 {
		d.RetrieveQuery.Params = rps
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if *flags.CassandraClearQuery != "" {
		d.ClearQuery.Query = *flags.CassandraClearQuery
	}
	if len(cps) > 0 {
		d.ClearQuery.Params = cps
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if *flags.CassandraFailQuery != "" {
		d.FailQuery.Query = *flags.CassandraFailQuery
	}
	if len(fps) > 0 {
		d.FailQuery.Params = fps
	}
	return nil
}

func (d *Cassandra) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "cassandra",
		"fn":  "Init",
	})
	l.Debug("Initializing cassandra client")

	cluster := gocql.NewCluster(d.Hosts...)
	// parse consistency string
	consistencyLevel := gocql.ParseConsistency(d.Consistency)
	cluster.Consistency = consistencyLevel
	if d.Keyspace != "" {
		cluster.Keyspace = d.Keyspace
	}
	cluster.ProtoVersion = 4
	cluster.ConnectTimeout = time.Second * 10
	if d.User != "" || d.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{Username: d.User, Password: d.Password}
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	d.Client = session
	return nil
}

func (d *Cassandra) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "cassandra",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from cassandra")
	var result string
	qry := d.Client.Query(d.RetrieveQuery.Query, d.RetrieveQuery.Params...)
	m, err := schema.CqlRowsToMapSlice(qry)
	if err != nil {
		// if the queue is empty, return nil
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	d.data = m
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
	l.Debug("Got work")
	// if result is empty, return nil
	if result == "" {
		l.Debug("result is empty")
		return nil, nil
	}
	return strings.NewReader(result), nil
}

func (d *Cassandra) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "cassandra",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from cassandra")
	var err error
	if d.ClearQuery == nil {
		return nil
	}
	if d.ClearQuery.Query == "" {
		return nil
	}
	d.ClearQuery.Params = schema.ReplaceParamsSliceMap(d.data, d.ClearQuery.Params)
	err = d.Client.Query(d.ClearQuery.Query, d.ClearQuery.Params...).Exec()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func (d *Cassandra) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "cassandra",
		"fn":  "HandleFailure",
	})
	l.Debug("handling failure for cassandra")
	var err error
	if d.FailQuery == nil {
		return nil
	}
	if d.FailQuery.Query == "" {
		return nil
	}
	d.FailQuery.Params = schema.ReplaceParamsSliceMap(d.data, d.FailQuery.Params)
	err = d.Client.Query(d.FailQuery.Query, d.FailQuery.Params...).Exec()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("handled failure")
	return nil
}

func (d *Cassandra) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "cassandra",
		"fn":  "Cleanup",
	})
	l.Debug("cleaning up cassandra")
	d.Client.Close()
	l.Debug("cleaned up cassandra")
	return nil
}
