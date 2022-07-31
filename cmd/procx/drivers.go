package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/robertlestak/procx/internal/drivers/aws"
	"github.com/robertlestak/procx/internal/drivers/cassandra"
	"github.com/robertlestak/procx/internal/drivers/centauri"
	"github.com/robertlestak/procx/internal/drivers/gcp"
	"github.com/robertlestak/procx/internal/drivers/mongo"
	"github.com/robertlestak/procx/internal/drivers/mysql"
	"github.com/robertlestak/procx/internal/drivers/postgres"
	"github.com/robertlestak/procx/internal/drivers/rabbitmq"
	"github.com/robertlestak/procx/internal/drivers/redis"
	"github.com/robertlestak/procx/pkg/procx"
	log "github.com/sirupsen/logrus"
)

func initAWSSQSDriver(j *procx.ProcX) error {
	if flagSQSQueueURL != nil && *flagSQSQueueURL != "" {
		j.Driver = &aws.SQS{
			Region:  *flagAWSRegion,
			RoleARN: *flagSQSRoleARN,
			Queue:   *flagSQSQueueURL,
		}
	}
	return nil
}

func initCentauriDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverCentauriNet {
		if flagCentauriKey == nil || (flagCentauriKey != nil && *flagCentauriKey == "") {
			return errors.New("key required")
		}
		kd := []byte(*flagCentauriKey)
		j.Driver = &centauri.Centauri{
			URL:        *flagCentauriPeerURL,
			Channel:    flagCentauriChannel,
			PrivateKey: kd,
		}
	}
	return nil
}

func initRabbitDriver(j *procx.ProcX) error {
	if flagRabbitMQURL != nil && *flagRabbitMQURL != "" {
		j.Driver = &rabbitmq.RabbitMQ{
			URL:   *flagRabbitMQURL,
			Queue: *flagRabbitMQQueue,
		}
	}
	return nil
}

func initRedisListDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverRedisList {
		j.Driver = &redis.RedisList{
			Host:     *flagRedisHost,
			Port:     *flagRedisPort,
			Password: *flagRedisPassword,
			Key:      *flagRedisKey,
		}
	}
	return nil
}

func initRedisPubSubDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverRedisSubscription {
		j.Driver = &redis.RedisPubSub{
			Host:     *flagRedisHost,
			Port:     *flagRedisPort,
			Password: *flagRedisPassword,
			Key:      *flagRedisKey,
		}
	}
	return nil
}

func initGCPPubSubDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverGCPPubSub {
		j.Driver = &gcp.GCPPubSub{
			ProjectID:        *flagGCPProjectID,
			SubscriptionName: *flagGCPSubscription,
		}
	}
	return nil
}

func initMongoDriver(j *procx.ProcX) error {
	if flagDriver != nil && procx.DriverName(*flagDriver) == procx.DriverMongoDB {
		pv, err := strconv.Atoi(*flagMongoPort)
		if err != nil {
			return err
		}
		j.Driver = &mongo.Mongo{
			Host:          *flagMongoHost,
			Port:          pv,
			User:          *flagMongoUser,
			Password:      *flagMongoPassword,
			DB:            *flagMongoDatabase,
			Collection:    *flagMongoCollection,
			RetrieveQuery: flagMongoRetrieveQuery,
			ClearQuery:    flagMongoClearQuery,
			FailQuery:     flagMongoFailQuery,
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
		driver := &postgres.Postgres{
			Host:    *flagPsqlHost,
			Port:    pv,
			User:    *flagPsqlUser,
			Pass:    *flagPsqlPassword,
			Db:      *flagPsqlDatabase,
			SslMode: *flagPsqlSSLMode,
		}
		if *flagPsqlQueryKey {
			driver.QueryKey = flagPsqlQueryKey
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
			driver.FailQuery = fq
		}

		j.Driver = driver
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
		driver := &mysql.Mysql{
			Host: *flagMysqlHost,
			Port: pv,
			User: *flagMysqlUser,
			Pass: *flagMysqlPassword,
			Db:   *flagMysqlDatabase,
		}
		if *flagMysqlQueryKey {
			driver.QueryKey = flagMysqlQueryKey
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
			driver.FailQuery = fq
		}

		j.Driver = driver
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
		driver := &cassandra.Cassandra{
			Hosts:       hosts,
			User:        *flagCassandraUser,
			Password:    *flagCassandraPassword,
			Keyspace:    *flagCassandraKeyspace,
			Consistency: *flagCassandraConsistency,
		}
		if *flagCassandraQueryKey {
			driver.QueryKey = flagCassandraQueryKey
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
			driver.FailQuery = fq
		}

		j.Driver = driver
	}
	return nil
}

func initDriver(j *procx.ProcX) (*procx.ProcX, error) {
	l := log.WithFields(log.Fields{
		"app": "procx",
	})
	l.Debug("starting")
	var err error
	switch procx.DriverName(*flagDriver) {
	case procx.DriverAWSSQS:
		err = initAWSSQSDriver(j)
	case procx.DriverCassandraDB:
		err = initCassandraDriver(j)
	case procx.DriverCentauriNet:
		err = initCentauriDriver(j)
	case procx.DriverMySQL:
		err = initMysqlDriver(j)
	case procx.DriverMongoDB:
		err = initMongoDriver(j)
	case procx.DriverGCPPubSub:
		err = initGCPPubSubDriver(j)
	case procx.DriverRabbit:
		err = initRabbitDriver(j)
	case procx.DriverPostgres:
		err = initPsqlDriver(j)
	case procx.DriverRedisList:
		err = initRedisListDriver(j)
	case procx.DriverRedisSubscription:
		err = initRedisPubSubDriver(j)
	default:
		err = procx.ErrDriverNotFound
	}
	if err != nil {
		return nil, err
	}
	l.Debug("exited")
	return j, nil
}
