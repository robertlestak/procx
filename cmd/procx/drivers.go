package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/robertlestak/procx/pkg/procx"
	log "github.com/sirupsen/logrus"
)

func initAWSDriver(j *procx.ProcX) {
	if flagSQSQueueURL != nil && *flagSQSQueueURL != "" {
		j.Driver = &procx.Driver{
			Name: procx.DriverAWSSQS,
			AWS: &procx.DriverAWS{
				Region:      *flagAWSRegion,
				RoleARN:     *flagSQSRoleARN,
				SQSQueueURL: *flagSQSQueueURL,
			},
		}
	}
}

func initCentauriDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverCentauriNet {
		if flagCentauriKey == nil || (flagCentauriKey != nil && *flagCentauriKey == "") {
			return errors.New("key required")
		}
		kd := []byte(*flagCentauriKey)
		j.Driver = &procx.Driver{
			Name: procx.DriverCentauriNet,
			Centauri: &procx.DriverCentauri{
				PeerURL:    *flagCentauriPeerURL,
				Channel:    flagCentauriChannel,
				PrivateKey: kd,
			},
		}
	}
	return nil
}

func initRabbitDriver(j *procx.ProcX) {
	if flagRabbitMQURL != nil && *flagRabbitMQURL != "" {
		j.Driver = &procx.Driver{
			Name: procx.DriverRabbit,
			RabbitMQ: &procx.DriverRabbitMQ{
				URL:   *flagRabbitMQURL,
				Queue: *flagRabbitMQQueue,
			},
		}
	}
}

func initRedisDriver(j *procx.ProcX) {
	if flagRedisHost != nil && *flagRedisHost != "" {
		j.Driver = &procx.Driver{
			Name: procx.DriverName(*flagDriver),
			Redis: &procx.DriverRedis{
				Host:     *flagRedisHost,
				Port:     *flagRedisPort,
				Password: *flagRedisPassword,
				Key:      *flagRedisKey,
			},
		}
	}
}

func initGCPDriver(j *procx.ProcX) {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverGCPPubSub {
		j.Driver = &procx.Driver{
			Name: procx.DriverGCPPubSub,
			GCP: &procx.DriverGCP{
				ProjectID:        *flagGCPProjectID,
				SubscriptionName: *flagGCPSubscription,
			},
		}
	}
}

func initMongoDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverMongoDB {
		pv, err := strconv.Atoi(*flagMongoPort)
		if err != nil {
			return err
		}
		j.Driver = &procx.Driver{
			Name: procx.DriverMongoDB,
			Mongo: &procx.DriverMongo{
				Host:          *flagMongoHost,
				Port:          pv,
				User:          *flagMongoUser,
				Password:      *flagMongoPassword,
				DBName:        *flagMongoDatabase,
				Collection:    *flagMongoCollection,
				RetrieveQuery: flagMongoRetrieveQuery,
				ClearQuery:    flagMongoClearQuery,
				FailureQuery:  flagMongoFailQuery,
			},
		}
	}
	return nil
}

func initPsqlDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverPostgres {
		pv, err := strconv.Atoi(*flagPsqlPort)
		if err != nil {
			return err
		}
		var rps []any
		var cps []any
		var fps []any
		if *flagPsqlRetrieveParams != "" {
			s := strings.Split(*flagPsqlRetrieveParams, ",")
			for _, v := range s {
				rps = append(rps, v)
			}
		}
		if *flagPsqlClearParams != "" {
			s := strings.Split(*flagPsqlClearParams, ",")
			for _, v := range s {
				cps = append(cps, v)
			}
		}
		if *flagPsqlFailParams != "" {
			s := strings.Split(*flagPsqlFailParams, ",")
			for _, v := range s {
				fps = append(fps, v)
			}
		}
		driver := &procx.DriverPsql{
			Host:     *flagPsqlHost,
			Port:     pv,
			User:     *flagPsqlUser,
			Password: *flagPsqlPassword,
			DBName:   *flagPsqlDatabase,
			SSLMode:  *flagPsqlSSLMode,
		}
		if *flagPsqlQueryKey {
			driver.QueryReturnsKey = flagPsqlQueryKey
		}
		if *flagPsqlRetrieveQuery != "" {
			rq := &procx.SqlQuery{
				Query:  *flagPsqlRetrieveQuery,
				Params: rps,
			}
			driver.RetrieveQuery = rq
		}
		if *flagPsqlClearQuery != "" {
			cq := &procx.SqlQuery{
				Query:  *flagPsqlClearQuery,
				Params: cps,
			}
			driver.ClearQuery = cq
		}
		if *flagPsqlFailQuery != "" {
			fq := &procx.SqlQuery{
				Query:  *flagPsqlFailQuery,
				Params: fps,
			}
			driver.FailureQuery = fq
		}

		j.Driver = &procx.Driver{
			Name: procx.DriverPostgres,
			Psql: driver,
		}
	}
	return nil
}

func initMysqlDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverMySQL {
		pv, err := strconv.Atoi(*flagMysqlPort)
		if err != nil {
			return err
		}
		var rps []any
		var cps []any
		var fps []any
		if *flagMysqlRetrieveParams != "" {
			s := strings.Split(*flagMysqlRetrieveParams, ",")
			for _, v := range s {
				rps = append(rps, v)
			}
		}
		if *flagMysqlClearParams != "" {
			s := strings.Split(*flagMysqlClearParams, ",")
			for _, v := range s {
				cps = append(cps, v)
			}
		}
		if *flagMysqlFailParams != "" {
			s := strings.Split(*flagMysqlFailParams, ",")
			for _, v := range s {
				fps = append(fps, v)
			}
		}
		driver := &procx.DriverMysql{
			Host:     *flagMysqlHost,
			Port:     pv,
			User:     *flagMysqlUser,
			Password: *flagMysqlPassword,
			DBName:   *flagMysqlDatabase,
		}
		if *flagMysqlQueryKey {
			driver.QueryReturnsKey = flagMysqlQueryKey
		}
		if *flagMysqlRetrieveQuery != "" {
			rq := &procx.SqlQuery{
				Query:  *flagMysqlRetrieveQuery,
				Params: rps,
			}
			driver.RetrieveQuery = rq
		}
		if *flagMysqlClearQuery != "" {
			cq := &procx.SqlQuery{
				Query:  *flagMysqlClearQuery,
				Params: cps,
			}
			driver.ClearQuery = cq
		}
		if *flagMysqlFailQuery != "" {
			fq := &procx.SqlQuery{
				Query:  *flagMysqlFailQuery,
				Params: fps,
			}
			driver.FailureQuery = fq
		}

		j.Driver = &procx.Driver{
			Name:  procx.DriverMySQL,
			Mysql: driver,
		}
	}
	return nil
}

func initCassandraDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverCassandraDB {
		var hosts []string
		if *flagCassandraHosts != "" {
			s := strings.Split(*flagCassandraHosts, ",")
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
		if *flagCassandraRetrieveParams != "" {
			s := strings.Split(*flagCassandraRetrieveParams, ",")
			for _, v := range s {
				rps = append(rps, v)
			}
		}
		if *flagCassandraClearParams != "" {
			s := strings.Split(*flagCassandraClearParams, ",")
			for _, v := range s {
				cps = append(cps, v)
			}
		}
		if *flagCassandraFailParams != "" {
			s := strings.Split(*flagCassandraFailParams, ",")
			for _, v := range s {
				fps = append(fps, v)
			}
		}
		driver := &procx.DriverCassandra{
			Hosts:       hosts,
			User:        *flagCassandraUser,
			Password:    *flagCassandraPassword,
			Keyspace:    *flagCassandraKeyspace,
			Consistency: *flagCassandraConsistency,
		}
		if *flagCassandraQueryKey {
			driver.QueryReturnsKey = flagCassandraQueryKey
		}
		if *flagCassandraRetrieveQuery != "" {
			rq := &procx.SqlQuery{
				Query:  *flagCassandraRetrieveQuery,
				Params: rps,
			}
			driver.RetrieveQuery = rq
		}
		if *flagCassandraClearQuery != "" {
			cq := &procx.SqlQuery{
				Query:  *flagCassandraClearQuery,
				Params: cps,
			}
			driver.ClearQuery = cq
		}
		if *flagCassandraFailQuery != "" {
			fq := &procx.SqlQuery{
				Query:  *flagCassandraFailQuery,
				Params: fps,
			}
			driver.FailureQuery = fq
		}

		j.Driver = &procx.Driver{
			Name:      procx.DriverCassandraDB,
			Cassandra: driver,
		}
	}
	return nil
}

func initDriver(j *procx.ProcX) (*procx.ProcX, error) {
	l := log.WithFields(log.Fields{
		"app": "procx",
	})
	l.Debug("starting")
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverAWSSQS {
		initAWSDriver(j)
	}
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverCentauriNet {
		if err := initCentauriDriver(j); err != nil {
			return nil, err
		}
	}
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverCassandraDB {
		if err := initCassandraDriver(j); err != nil {
			return nil, err
		}
	}
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverRabbit {
		initRabbitDriver(j)
	}
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverRedisList {
		initRedisDriver(j)
	}
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverRedisSubscription {
		initRedisDriver(j)
	}
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverGCPPubSub {
		initGCPDriver(j)
	}
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverPostgres {
		if err := initPsqlDriver(j); err != nil {
			return nil, err
		}
	}
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverMySQL {
		if err := initMysqlDriver(j); err != nil {
			return nil, err
		}
	}
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverMongoDB {
		if err := initMongoDriver(j); err != nil {
			return nil, err
		}
	}
	l.Debug("exited")
	return j, nil
}
