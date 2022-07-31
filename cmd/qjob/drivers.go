package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/robertlestak/qjob/pkg/qjob"
	log "github.com/sirupsen/logrus"
)

func initAWSDriver(j *qjob.QJob) {
	if flagSQSQueueURL != nil && *flagSQSQueueURL != "" {
		j.Driver = &qjob.Driver{
			Name: qjob.DriverAWSSQS,
			AWS: &qjob.DriverAWS{
				Region:      *flagAWSRegion,
				RoleARN:     *flagSQSRoleARN,
				SQSQueueURL: *flagSQSQueueURL,
			},
		}
	}
}

func initCentauriDriver(j *qjob.QJob) error {
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverCentauriNet {
		if flagCentauriKey == nil || (flagCentauriKey != nil && *flagCentauriKey == "") {
			return errors.New("key required")
		}
		kd := []byte(*flagCentauriKey)
		j.Driver = &qjob.Driver{
			Name: qjob.DriverCentauriNet,
			Centauri: &qjob.DriverCentauri{
				PeerURL:    *flagCentauriPeerURL,
				Channel:    flagCentauriChannel,
				PrivateKey: kd,
			},
		}
	}
	return nil
}

func initRabbitDriver(j *qjob.QJob) {
	if flagRabbitMQURL != nil && *flagRabbitMQURL != "" {
		j.Driver = &qjob.Driver{
			Name: qjob.DriverRabbit,
			RabbitMQ: &qjob.DriverRabbitMQ{
				URL:   *flagRabbitMQURL,
				Queue: *flagRabbitMQQueue,
			},
		}
	}
}

func initRedisDriver(j *qjob.QJob) {
	if flagRedisHost != nil && *flagRedisHost != "" {
		j.Driver = &qjob.Driver{
			Name: qjob.DriverName(*flagDriver),
			Redis: &qjob.DriverRedis{
				Host:     *flagRedisHost,
				Port:     *flagRedisPort,
				Password: *flagRedisPassword,
				Key:      *flagRedisKey,
			},
		}
	}
}

func initGCPDriver(j *qjob.QJob) {
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverGCPPubSub {
		j.Driver = &qjob.Driver{
			Name: qjob.DriverGCPPubSub,
			GCP: &qjob.DriverGCP{
				ProjectID:        *flagGCPProjectID,
				SubscriptionName: *flagGCPSubscription,
			},
		}
	}
}

func initMongoDriver(j *qjob.QJob) error {
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverMongoDB {
		pv, err := strconv.Atoi(*flagMongoPort)
		if err != nil {
			return err
		}
		j.Driver = &qjob.Driver{
			Name: qjob.DriverMongoDB,
			Mongo: &qjob.DriverMongo{
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

func initPsqlDriver(j *qjob.QJob) error {
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverPostgres {
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
		driver := &qjob.DriverPsql{
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
			rq := &qjob.SqlQuery{
				Query:  *flagPsqlRetrieveQuery,
				Params: rps,
			}
			driver.RetrieveQuery = rq
		}
		if *flagPsqlClearQuery != "" {
			cq := &qjob.SqlQuery{
				Query:  *flagPsqlClearQuery,
				Params: cps,
			}
			driver.ClearQuery = cq
		}
		if *flagPsqlFailQuery != "" {
			fq := &qjob.SqlQuery{
				Query:  *flagPsqlFailQuery,
				Params: fps,
			}
			driver.FailureQuery = fq
		}

		j.Driver = &qjob.Driver{
			Name: qjob.DriverPostgres,
			Psql: driver,
		}
	}
	return nil
}

func initMysqlDriver(j *qjob.QJob) error {
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverMySQL {
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
		driver := &qjob.DriverMysql{
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
			rq := &qjob.SqlQuery{
				Query:  *flagMysqlRetrieveQuery,
				Params: rps,
			}
			driver.RetrieveQuery = rq
		}
		if *flagMysqlClearQuery != "" {
			cq := &qjob.SqlQuery{
				Query:  *flagMysqlClearQuery,
				Params: cps,
			}
			driver.ClearQuery = cq
		}
		if *flagMysqlFailQuery != "" {
			fq := &qjob.SqlQuery{
				Query:  *flagMysqlFailQuery,
				Params: fps,
			}
			driver.FailureQuery = fq
		}

		j.Driver = &qjob.Driver{
			Name:  qjob.DriverMySQL,
			Mysql: driver,
		}
	}
	return nil
}

func initDriver(j *qjob.QJob) (*qjob.QJob, error) {
	l := log.WithFields(log.Fields{
		"app": "qjob",
	})
	l.Debug("starting")
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverAWSSQS {
		initAWSDriver(j)
	}
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverCentauriNet {
		if err := initCentauriDriver(j); err != nil {
			return nil, err
		}
	}
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverRabbit {
		initRabbitDriver(j)
	}
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverRedisList {
		initRedisDriver(j)
	}
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverRedisSubscription {
		initRedisDriver(j)
	}
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverGCPPubSub {
		initGCPDriver(j)
	}
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverPostgres {
		if err := initPsqlDriver(j); err != nil {
			return nil, err
		}
	}
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverMySQL {
		if err := initMysqlDriver(j); err != nil {
			return nil, err
		}
	}
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverMongoDB {
		if err := initMongoDriver(j); err != nil {
			return nil, err
		}
	}
	l.Debug("exited")
	return j, nil
}
