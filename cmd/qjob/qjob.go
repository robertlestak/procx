package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/robertlestak/qjob/pkg/qjob"
	log "github.com/sirupsen/logrus"
)

var (
	Version                = "dev"
	flagDriver             = flag.String("driver", "", "driver to use. (aws-sqs, gcp-pubsub, postgres, rabbitmq, redis-list, redis-pubsub, local)")
	flagHostEnv            = flag.Bool("hostenv", false, "use host environment")
	flagAWSRegion          = flag.String("aws-region", "", "AWS region")
	flagAWSLoadConfig      = flag.Bool("aws-load-config", false, "load AWS config from ~/.aws/config")
	flagSQSRoleARN         = flag.String("aws-sqs-role-arn", "", "AWS SQS role ARN")
	flagSQSQueueURL        = flag.String("aws-sqs-queue-url", "", "AWS SQS queue URL")
	flagGCPProjectID       = flag.String("gcp-project-id", "", "GCP project ID")
	flagGCPSubscription    = flag.String("gcp-pubsub-subscription", "", "GCP Pub/Sub subscription name")
	flagPassWorkAsArg      = flag.Bool("pass-work-as-arg", false, "pass work as an argument")
	flagPsqlHost           = flag.String("psql-host", "", "PostgreSQL host")
	flagPsqlPort           = flag.String("psql-port", "5432", "PostgreSQL port")
	flagPsqlUser           = flag.String("psql-user", "", "PostgreSQL user")
	flagPsqlPassword       = flag.String("psql-password", "", "PostgreSQL password")
	flagPsqlDatabase       = flag.String("psql-database", "", "PostgreSQL database")
	flagPsqlSSLMode        = flag.String("psql-ssl-mode", "disable", "PostgreSQL SSL mode")
	flagPsqlQueryKey       = flag.Bool("psql-query-key", false, "PostgreSQL query returns key as first column and value as second column")
	flagPsqlRetrieveQuery  = flag.String("psql-retrieve-query", "", "PostgreSQL retrieve query")
	flagPsqlRetrieveParams = flag.String("psql-retrieve-params", "", "PostgreSQL retrieve params")
	flagPsqlClearQuery     = flag.String("psql-clear-query", "", "PostgreSQL clear query")
	flagPsqlClearParams    = flag.String("psql-clear-params", "", "PostgreSQL clear params")
	flagPsqlFailQuery      = flag.String("psql-fail-query", "", "PostgreSQL fail query")
	flagPsqlFailParams     = flag.String("psql-fail-params", "", "PostgreSQL fail params")
	flagRabbitMQURL        = flag.String("rabbitmq-url", "", "RabbitMQ URL")
	flagRabbitMQQueue      = flag.String("rabbitmq-queue", "", "RabbitMQ queue")
	flagRedisHost          = flag.String("redis-host", "", "Redis host")
	flagRedisPort          = flag.String("redis-port", "6379", "Redis port")
	flagRedisPassword      = flag.String("redis-password", "", "Redis password")
	flagRedisKey           = flag.String("redis-key", "", "Redis key")
	flagDaemon             = flag.Bool("daemon", false, "run as daemon")
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func initDriver(j *qjob.QJob) (*qjob.QJob, error) {
	l := log.WithFields(log.Fields{
		"app": "qjob",
	})
	l.Debug("starting")
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
	if flagRabbitMQURL != nil && *flagRabbitMQURL != "" {
		j.Driver = &qjob.Driver{
			Name: qjob.DriverRabbit,
			RabbitMQ: &qjob.DriverRabbitMQ{
				URL:   *flagRabbitMQURL,
				Queue: *flagRabbitMQQueue,
			},
		}
	}
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
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverGCPPubSub {
		j.Driver = &qjob.Driver{
			Name: qjob.DriverGCPPubSub,
			GCP: &qjob.DriverGCP{
				ProjectID:        *flagGCPProjectID,
				SubscriptionName: *flagGCPSubscription,
			},
		}
	}
	if flagDriver != nil && qjob.DriverName(*flagDriver) == qjob.DriverPostgres {
		pv, err := strconv.Atoi(*flagPsqlPort)
		if err != nil {
			return nil, err
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
	l.Debug("exited")
	return j, nil
}

func parseEnvToFlags() {
	if os.Getenv("QJOB_DRIVER") != "" {
		d := os.Getenv("QJOB_DRIVER")
		flagDriver = &d
	}
	if os.Getenv("QJOB_HOSTENV") != "" {
		h := os.Getenv("QJOB_HOSTENV")
		t := h == "true"
		flagHostEnv = &t
	}
	if os.Getenv("QJOB_AWS_REGION") != "" {
		r := os.Getenv("QJOB_AWS_REGION")
		flagAWSRegion = &r
	}
	if os.Getenv("QJOB_AWS_SQS_ROLE_ARN") != "" {
		r := os.Getenv("QJOB_AWS_SQS_ROLE_ARN")
		flagSQSRoleARN = &r
	}
	if os.Getenv("QJOB_AWS_SQS_QUEUE_URL") != "" {
		r := os.Getenv("QJOB_AWS_SQS_QUEUE_URL")
		flagSQSQueueURL = &r
	}
	if os.Getenv("QJOB_PASS_WORK_AS_ARG") != "" {
		r := os.Getenv("QJOB_PASS_WORK_AS_ARG")
		t := r == "true"
		flagPassWorkAsArg = &t
	}
	if os.Getenv("QJOB_RABBITMQ_URL") != "" {
		r := os.Getenv("QJOB_RABBITMQ_URL")
		flagRabbitMQURL = &r
	}
	if os.Getenv("QJOB_RABBITMQ_QUEUE") != "" {
		r := os.Getenv("QJOB_RABBITMQ_QUEUE")
		flagRabbitMQQueue = &r
	}
	if os.Getenv("QJOB_DAEMON") != "" {
		r := os.Getenv("QJOB_DAEMON")
		t := r == "true"
		flagDaemon = &t
	}
	if os.Getenv("QJOB_AWS_LOAD_CONFIG") != "" || os.Getenv("AWS_SDK_LOAD_CONFIG") != "" {
		r := os.Getenv("QJOB_AWS_LOAD_CONFIG")
		t := r == "true"
		flagAWSLoadConfig = &t
	}
	if os.Getenv("QJOB_REDIS_HOST") != "" {
		r := os.Getenv("QJOB_REDIS_HOST")
		flagRedisHost = &r
	}
	if os.Getenv("QJOB_REDIS_PORT") != "" {
		r := os.Getenv("QJOB_REDIS_PORT")
		flagRedisPort = &r
	}
	if os.Getenv("QJOB_REDIS_PASSWORD") != "" {
		r := os.Getenv("QJOB_REDIS_PASSWORD")
		flagRedisPassword = &r
	}
	if os.Getenv("QJOB_REDIS_KEY") != "" {
		r := os.Getenv("QJOB_REDIS_KEY")
		flagRedisKey = &r
	}
	if os.Getenv("QJOB_GCP_PROJECT_ID") != "" {
		r := os.Getenv("QJOB_GCP_PROJECT_ID")
		flagGCPProjectID = &r
	}
	if os.Getenv("QJOB_GCP_SUBSCRIPTION") != "" {
		r := os.Getenv("QJOB_GCP_SUBSCRIPTION")
		flagGCPSubscription = &r
	}
	if os.Getenv("QJOB_PSQL_HOST") != "" {
		r := os.Getenv("QJOB_PSQL_HOST")
		flagPsqlHost = &r
	}
	if os.Getenv("QJOB_PSQL_PORT") != "" {
		r := os.Getenv("QJOB_PSQL_PORT")
		flagPsqlPort = &r
	}
	if os.Getenv("QJOB_PSQL_USER") != "" {
		r := os.Getenv("QJOB_PSQL_USER")
		flagPsqlUser = &r
	}
	if os.Getenv("QJOB_PSQL_PASSWORD") != "" {
		r := os.Getenv("QJOB_PSQL_PASSWORD")
		flagPsqlPassword = &r
	}
	if os.Getenv("QJOB_PSQL_DATABASE") != "" {
		r := os.Getenv("QJOB_PSQL_DATABASE")
		flagPsqlDatabase = &r
	}
	if os.Getenv("QJOB_PSQL_SSL_MODE") != "" {
		r := os.Getenv("QJOB_PSQL_SSL_MODE")
		flagPsqlSSLMode = &r
	}
	if os.Getenv("QJOB_PSQL_RETRIEVE_QUERY") != "" {
		r := os.Getenv("QJOB_PSQL_RETRIEVE_QUERY")
		flagPsqlRetrieveQuery = &r
	}
	if os.Getenv("QJOB_PSQL_CLEAR_QUERY") != "" {
		r := os.Getenv("QJOB_PSQL_CLEAR_QUERY")
		flagPsqlClearQuery = &r
	}
	if os.Getenv("QJOB_PSQL_FAIL_QUERY") != "" {
		r := os.Getenv("QJOB_PSQL_FAIL_QUERY")
		flagPsqlFailQuery = &r
	}
	if os.Getenv("QJOB_PSQL_RETRIEVE_PARAMS") != "" {
		r := os.Getenv("QJOB_PSQL_RETRIEVE_PARAMS")
		flagPsqlRetrieveParams = &r
	}
	if os.Getenv("QJOB_PSQL_CLEAR_PARAMS") != "" {
		r := os.Getenv("QJOB_PSQL_CLEAR_PARAMS")
		flagPsqlClearParams = &r
	}
	if os.Getenv("QJOB_PSQL_FAIL_PARAMS") != "" {
		r := os.Getenv("QJOB_PSQL_FAIL_PARAMS")
		flagPsqlFailParams = &r
	}
	if os.Getenv("QJOB_PSQL_QUERY_KEY") != "" {
		r := os.Getenv("QJOB_PSQL_QUERY_KEY")
		t := r == "true"
		flagPsqlQueryKey = &t
	}
	if *flagAWSLoadConfig {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
}

func printVersion() {
	fmt.Printf("qjob version %s\n", Version)
}

func runOnce() {
	l := log.WithFields(log.Fields{
		"app": "qjob",
	})
	l.Debug("starting")
	args := flag.Args()
	j := &qjob.QJob{
		DriverName:    qjob.DriverName(*flagDriver),
		HostEnv:       *flagHostEnv,
		PassWorkAsArg: *flagPassWorkAsArg,
	}
	var err error
	j, err = initDriver(j)
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}
	j.ParseArgs(args)
	l.Debug("parsed args")
	// execute
	if j.Bin == "" {
		l.Error("no bin specified")
		os.Exit(1)
	}
	if err := j.InitDriver(); err != nil {
		l.Errorf("failed to init driver: %s", err)
		os.Exit(1)
	}
	if err := j.DoWork(); err != nil {
		l.Errorf("failed to do work: %s", err)
		os.Exit(1)
	}
}

func main() {
	l := log.WithFields(log.Fields{
		"app": "qjob",
	})
	l.Debug("starting")
	if len(os.Args) < 2 {
		printVersion()
		flag.PrintDefaults()
		os.Exit(1)
	}
	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		printVersion()
		os.Exit(0)
	}
	flag.Parse()
	parseEnvToFlags()
	l.Debug("parsed flags")
	args := flag.Args()
	if len(args) == 0 {
		// print help
		printVersion()
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *flagDaemon {
		l.Debug("running as daemon")
		for {
			runOnce()
		}
	} else {
		runOnce()
	}
	l.Debug("exited")
}
