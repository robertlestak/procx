package cassandra

import (
	"os"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/robertlestak/procx/internal/flags"
	"github.com/robertlestak/procx/pkg/schema"
	log "github.com/sirupsen/logrus"
)

type Cassandra struct {
	Client        *gocql.Session
	Hosts         []string
	User          string
	Password      string
	Consistency   string
	Keyspace      string
	QueryKey      *bool
	Key           *string
	RetrieveQuery *schema.SqlQuery
	ClearQuery    *schema.SqlQuery
	FailQuery     *schema.SqlQuery
}

func (d *Cassandra) LoadEnv(prefix string) error {
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
	if os.Getenv(prefix+"CASSANDRA_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery = &schema.SqlQuery{Query: os.Getenv(prefix + "CASSANDRA_RETRIEVE_QUERY")}
	}
	if os.Getenv(prefix+"CASSANDRA_RETRIEVE_PARAMS") != "" {
		for _, s := range strings.Split(os.Getenv(prefix+"CASSANDRA_RETRIEVE_PARAMS"), ",") {
			d.RetrieveQuery.Params = append(d.RetrieveQuery.Params, s)
		}
	}
	if os.Getenv(prefix+"CASSANDRA_CLEAR_QUERY") != "" {
		d.ClearQuery = &schema.SqlQuery{Query: os.Getenv(prefix + "CASSANDRA_CLEAR_QUERY")}
	}
	if os.Getenv(prefix+"CASSANDRA_CLEAR_PARAMS") != "" {
		for _, s := range strings.Split(os.Getenv(prefix+"CASSANDRA_CLEAR_PARAMS"), ",") {
			d.ClearQuery.Params = append(d.ClearQuery.Params, s)
		}
	}
	if os.Getenv(prefix+"CASSANDRA_FAIL_QUERY") != "" {
		d.FailQuery = &schema.SqlQuery{Query: os.Getenv(prefix + "CASSANDRA_FAIL_QUERY")}
	}
	if os.Getenv(prefix+"CASSANDRA_FAIL_PARAMS") != "" {
		for _, s := range strings.Split(os.Getenv(prefix+"CASSANDRA_FAIL_PARAMS"), ",") {
			d.FailQuery.Params = append(d.FailQuery.Params, s)
		}
	}
	if os.Getenv(prefix+"CASSANDRA_QUERY_KEY") != "" {
		tval := os.Getenv(prefix+"CASSANDRA_QUERY_KEY") == "true"
		d.QueryKey = &tval
	}
	return nil
}

func (d *Cassandra) LoadFlags() error {
	var hosts []string
	if *flags.FlagCassandraHosts != "" {
		s := strings.Split(*flags.FlagCassandraHosts, ",")
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
	if *flags.FlagCassandraRetrieveParams != "" {
		s := strings.Split(*flags.FlagCassandraRetrieveParams, ",")
		for _, v := range s {
			rps = append(rps, v)
		}
	}
	if *flags.FlagCassandraClearParams != "" {
		s := strings.Split(*flags.FlagCassandraClearParams, ",")
		for _, v := range s {
			cps = append(cps, v)
		}
	}
	if *flags.FlagCassandraFailParams != "" {
		s := strings.Split(*flags.FlagCassandraFailParams, ",")
		for _, v := range s {
			fps = append(fps, v)
		}
	}
	d.Hosts = hosts
	d.User = *flags.FlagCassandraUser
	d.Password = *flags.FlagCassandraPassword
	d.Keyspace = *flags.FlagCassandraKeyspace
	d.Consistency = *flags.FlagCassandraConsistency
	if *flags.FlagCassandraQueryKey {
		d.QueryKey = flags.FlagCassandraQueryKey
	}
	if *flags.FlagCassandraRetrieveQuery != "" {
		rq := &schema.SqlQuery{
			Query:  *flags.FlagCassandraRetrieveQuery,
			Params: rps,
		}
		d.RetrieveQuery = rq
	}
	if *flags.FlagCassandraClearQuery != "" {
		cq := &schema.SqlQuery{
			Query:  *flags.FlagCassandraClearQuery,
			Params: cps,
		}
		d.ClearQuery = cq
	}
	if *flags.FlagCassandraFailQuery != "" {
		fq := &schema.SqlQuery{
			Query:  *flags.FlagCassandraFailQuery,
			Params: fps,
		}
		d.FailQuery = fq
	}
	return nil
}

func (d *Cassandra) Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreateCassandraClient",
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

func (d *Cassandra) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWork",
	})
	l.Debug("Getting work from cassandra")
	var err error
	var result string
	var key string
	if d.QueryKey != nil && *d.QueryKey {
		err = d.Client.Query(d.RetrieveQuery.Query).Scan(&key, &result)
	} else {
		err = d.Client.Query(d.RetrieveQuery.Query).Scan(&result)
	}
	if err != nil {
		// if the queue is empty, return nil
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	l.Debug("Got work")
	d.Key = &key
	return &result, nil
}

func (d *Cassandra) ClearWork() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkCassandra",
	})
	l.Debug("Clearing work from cassandra")
	var err error
	if d.ClearQuery == nil {
		return nil
	}
	if d.ClearQuery.Query == "" {
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
		"package": "cache",
		"method":  "HandleFailureCassandra",
	})
	l.Debug("handling failure for cassandra")
	var err error
	if d.FailQuery == nil {
		return nil
	}
	if d.FailQuery.Query == "" {
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
	err = d.Client.Query(d.FailQuery.Query, d.FailQuery.Params...).Exec()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("handled failure")
	return nil
}
