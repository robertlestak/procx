package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/robertlestak/qjob/internal/qjob"
	log "github.com/sirupsen/logrus"
)

var (
	Version           = "dev"
	flagDriver        = flag.String("driver", "", "driver to use. (aws-sqs, rabbitmq, local)")
	flagHostEnv       = flag.Bool("hostenv", false, "use host environment")
	flagAWSRegion     = flag.String("aws-region", "", "AWS region")
	flagSQSRoleARN    = flag.String("aws-sqs-role-arn", "", "AWS SQS role ARN")
	flagSQSQueueURL   = flag.String("aws-sqs-queue-url", "", "AWS SQS queue URL")
	flagPassWorkAsArg = flag.Bool("pass-work-as-arg", false, "pass work as an argument")
	flagRabbitMQURL   = flag.String("rabbitmq-url", "", "RabbitMQ URL")
	flagRabbitMQQueue = flag.String("rabbitmq-queue", "", "RabbitMQ queue")
	flagDaemon        = flag.Bool("daemon", false, "run as daemon")
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func initDriver(j *qjob.QJob) error {
	l := log.WithFields(log.Fields{
		"app": "qjob",
	})
	l.Debug("starting")
	if flagSQSQueueURL != nil {
		j.Driver = &qjob.Driver{
			Name: qjob.DriverAWSSQS,
			AWS: &qjob.DriverAWS{
				Region:      *flagAWSRegion,
				RoleARN:     *flagSQSRoleARN,
				SQSQueueURL: *flagSQSQueueURL,
			},
		}
	}
	if flagRabbitMQURL != nil {
		j.Driver = &qjob.Driver{
			Name: qjob.DriverRabbit,
			RabbitMQ: &qjob.DriverRabbitMQ{
				URL:   *flagRabbitMQURL,
				Queue: *flagRabbitMQQueue,
			},
		}
	}
	l.Debug("exited")
	return nil
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
	if err := initDriver(j); err != nil {
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
