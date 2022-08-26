package scylla

import (
	"encoding/json"
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

type Scylla struct {
	Client        *gocql.Session
	Hosts         []string
	User          string
	Password      string
	Consistency   string
	LocalDC       *string
	Keyspace      string
	RetrieveField *string
	RetrieveQuery *schema.SqlQuery
	ClearQuery    *schema.SqlQuery
	FailQuery     *schema.SqlQuery
	data          []map[string]any
}

func (d *Scylla) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "scylla",
		"fn":  "LoadEnv",
	})
	l.Debug("loading env")
	if os.Getenv(prefix+"SCYLLA_HOSTS") != "" {
		d.Hosts = strings.Split(os.Getenv(prefix+"SCYLLA_HOSTS"), ",")
	}
	if os.Getenv(prefix+"SCYLLA_KEYSPACE") != "" {
		d.Keyspace = os.Getenv(prefix + "SCYLLA_KEYSPACE")
	}
	if os.Getenv(prefix+"SCYLLA_USER") != "" {
		d.User = os.Getenv(prefix + "SCYLLA_USER")
	}
	if os.Getenv(prefix+"SCYLLA_PASSWORD") != "" {
		d.Password = os.Getenv(prefix + "SCYLLA_PASSWORD")
	}
	if os.Getenv(prefix+"SCYLLA_CONSISTENCY") != "" {
		d.Consistency = os.Getenv(prefix + "SCYLLA_CONSISTENCY")
	}
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"SCYLLA_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery.Query = os.Getenv(prefix + "SCYLLA_RETRIEVE_QUERY")
	}
	if os.Getenv(prefix+"SCYLLA_RETRIEVE_PARAMS") != "" {
		for _, s := range strings.Split(os.Getenv(prefix+"SCYLLA_RETRIEVE_PARAMS"), ",") {
			d.RetrieveQuery.Params = append(d.RetrieveQuery.Params, s)
		}
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"SCYLLA_CLEAR_QUERY") != "" {
		d.ClearQuery.Query = os.Getenv(prefix + "SCYLLA_CLEAR_QUERY")
	}
	if os.Getenv(prefix+"SCYLLA_CLEAR_PARAMS") != "" {
		for _, s := range strings.Split(os.Getenv(prefix+"SCYLLA_CLEAR_PARAMS"), ",") {
			d.ClearQuery.Params = append(d.ClearQuery.Params, s)
		}
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"SCYLLA_FAIL_QUERY") != "" {
		d.FailQuery.Query = os.Getenv(prefix + "SCYLLA_FAIL_QUERY")
	}
	if os.Getenv(prefix+"SCYLLA_FAIL_PARAMS") != "" {
		for _, s := range strings.Split(os.Getenv(prefix+"SCYLLA_FAIL_PARAMS"), ",") {
			d.FailQuery.Params = append(d.FailQuery.Params, s)
		}
	}
	if os.Getenv(prefix+"SCYLLA_RETRIEVE_FIELD") != "" {
		v := os.Getenv(prefix + "SCYLLA_RETRIEVE_FIELD")
		d.RetrieveField = &v
	}
	if os.Getenv(prefix+"SCYLLA_LOCAL_DC") != "" {
		v := os.Getenv(prefix + "SCYLLA_LOCAL_DC")
		d.LocalDC = &v
	}
	return nil
}

func (d *Scylla) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "scylla",
		"fn":  "LoadFlags",
	})
	l.Debug("loading flags")
	var hosts []string
	if *flags.ScyllaHosts != "" {
		s := strings.Split(*flags.ScyllaHosts, ",")
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
	if *flags.ScyllaRetrieveParams != "" {
		s := strings.Split(*flags.ScyllaRetrieveParams, ",")
		for _, v := range s {
			rps = append(rps, v)
		}
	}
	if *flags.ScyllaClearParams != "" {
		s := strings.Split(*flags.ScyllaClearParams, ",")
		for _, v := range s {
			cps = append(cps, v)
		}
	}
	if *flags.ScyllaFailParams != "" {
		s := strings.Split(*flags.ScyllaFailParams, ",")
		for _, v := range s {
			fps = append(fps, v)
		}
	}
	d.Hosts = hosts
	d.User = *flags.ScyllaUser
	d.Password = *flags.ScyllaPassword
	d.Keyspace = *flags.ScyllaKeyspace
	d.Consistency = *flags.ScyllaConsistency
	d.LocalDC = flags.ScyllaLocalDC
	d.RetrieveField = flags.ScyllaRetrieveField
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if *flags.ScyllaRetrieveQuery != "" {
		d.RetrieveQuery.Query = *flags.ScyllaRetrieveQuery
	}
	if len(rps) > 0 {
		d.RetrieveQuery.Params = rps
	}
	if d.ClearQuery == nil {
		d.ClearQuery = &schema.SqlQuery{}
	}
	if *flags.ScyllaClearQuery != "" {
		d.ClearQuery.Query = *flags.ScyllaClearQuery
	}
	if len(cps) > 0 {
		d.ClearQuery.Params = cps
	}
	if d.FailQuery == nil {
		d.FailQuery = &schema.SqlQuery{}
	}
	if *flags.ScyllaFailQuery != "" {
		d.FailQuery.Query = *flags.ScyllaFailQuery
	}
	if len(fps) > 0 {
		d.FailQuery.Params = fps
	}
	return nil
}

func (d *Scylla) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "scylla",
		"fn":  "Init",
	})
	l.Debug("Initializing scylla client")

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
	fallback := gocql.RoundRobinHostPolicy()
	if d.LocalDC != nil && *d.LocalDC != "" {
		fallback = gocql.DCAwareRoundRobinPolicy(*d.LocalDC)
	}
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(fallback)
	if *d.LocalDC != "" {
		cluster.Consistency = gocql.LocalQuorum
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}
	d.Client = session
	return nil
}

func (d *Scylla) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "scylla",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from scylla")
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
		result = gjson.GetBytes(bd, *d.RetrieveField).String()
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

func (d *Scylla) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "scylla",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from scylla")
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

func (d *Scylla) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "scylla",
		"fn":  "HandleFailure",
	})
	l.Debug("handling failure for scylla")
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

func (d *Scylla) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "scylla",
		"fn":  "Cleanup",
	})
	l.Debug("cleaning up scylla")
	d.Client.Close()
	l.Debug("cleaned up scylla")
	return nil
}
