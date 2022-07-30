package qjob

import (
	"errors"
	"io"
	"os"
	"os/exec"

	"cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"
)

var (
	DriverAWSSQS            DriverName = "aws-sqs"
	DriverGCPPubSub         DriverName = "gcp-pubsub"
	DriverRabbit            DriverName = "rabbitmq"
	DriverRedisSubscription DriverName = "redis-subscription"
	DriverRedisList         DriverName = "redis-list"
	DriverLocal             DriverName = "local"
	ErrDriverNotFound                  = errors.New("driver not found")
)

type DriverName string

type DriverAWS struct {
	Region      string
	RoleARN     string
	SQSQueueURL string
}

type DriverGCP struct {
	ProjectID        string
	SubscriptionName string
	PubSubMessage    *pubsub.Message
}

type DriverRabbitMQ struct {
	URL   string
	Queue string
}

type DriverRedis struct {
	Host     string
	Port     string
	Password string
	Key      string
}

type Driver struct {
	Name     DriverName
	AWS      *DriverAWS
	GCP      *DriverGCP
	RabbitMQ *DriverRabbitMQ
	Redis    *DriverRedis
}

type QJob struct {
	DriverName    DriverName
	Driver        *Driver
	PassWorkAsArg bool
	HostEnv       bool
	Bin           string
	Args          []string
	work          string
}

func (j *QJob) ParseArgs(args []string) {
	if len(args) == 0 {
		return
	}
	j.Bin = args[0]
	if len(args) > 1 {
		j.Args = args[1:]
	}
}

func (j *QJob) InitDriver() error {
	l := log.WithFields(log.Fields{
		"action": "InitDriver",
		"driver": j.DriverName,
	})
	l.Debug("InitDriver")
	switch j.DriverName {
	case DriverAWSSQS:
		return j.InitAWSSQS()
	case DriverGCPPubSub:
		return j.InitGCPPubSub()
	case DriverRabbit:
		return j.InitRabbitMQ()
	case DriverRedisSubscription:
		return j.InitRedis()
	case DriverRedisList:
		return j.InitRedis()
	case DriverLocal:
		return nil
	default:
		return ErrDriverNotFound
	}
}

func (j *QJob) GetWorkFromDriver() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "GetWorkFromDriver",
		"driver": j.DriverName,
	})
	l.Debug("GetWorkFromDriver")
	switch j.DriverName {
	case DriverAWSSQS:
		return j.getWorkSQS()
	case DriverGCPPubSub:
		return j.getWorkGCPPubSub()
	case DriverRabbit:
		return j.getWorkRabbitMQ()
	case DriverLocal:
		w := os.Getenv("QJOB_PAYLOAD")
		return &w, nil
	case DriverRedisList:
		return j.getWorkRedisList()
	case DriverRedisSubscription:
		return j.getWorkRedisSubscription()
	default:
		return nil, ErrDriverNotFound
	}
}

func (j *QJob) ClearWorkFromDriver() error {
	l := log.WithFields(log.Fields{
		"action": "ClearWorkFromDriver",
		"driver": j.DriverName,
	})
	l.Debug("ClearWorkFromDriver")
	switch j.DriverName {
	case DriverAWSSQS:
		return j.clearWorkSQS()
	case DriverRabbit:
		return j.clearWorkRabbitMQ()
	case DriverGCPPubSub:
		return j.clearWorkGCPPubSub()
	case DriverRedisList:
		return nil
	case DriverRedisSubscription:
		return nil
	case DriverLocal:
		return nil
	default:
		return ErrDriverNotFound
	}
}

func (j *QJob) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"action": "HandleFailure",
		"driver": j.DriverName,
	})
	l.Debug("HandleFailure")
	switch j.DriverName {
	case DriverRedisList:
		return j.handleFailureRedisList()
	}
	return nil
}

func (j *QJob) DoWork() error {
	l := log.WithFields(log.Fields{
		"action": "DoWork",
		"driver": j.DriverName,
	})
	l.Debug("DoWork")
	work, err := j.GetWorkFromDriver()
	if err != nil {
		l.Error(err)
		return err
	}
	if work == nil {
		l.Debug("no work")
		return nil
	}
	j.work = *work
	l.Debug("work received")
	err = j.Exec(os.Stdout, os.Stderr)
	if err != nil {
		l.Error(err)
		if err := j.HandleFailure(); err != nil {
			l.Error(err)
		}
		return err
	}
	l.Debug("work completed")
	err = j.ClearWorkFromDriver()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("work cleared")
	return nil
}

// Exec will execute the given script, streaming the output to the provided
// io.Writers. If the script exits with a non-zero exit code, an error will be
// returned. If the script exits with a zero exit code, no error will be
// returned.
func (j *QJob) Exec(stdout, stderr io.Writer) error {
	// create the command
	if j.PassWorkAsArg {
		j.Args = append(j.Args, j.work)
	}
	cmd := exec.Command(j.Bin, j.Args...)
	// set the stdout and stderr pipes
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if j.HostEnv {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "QJOB_PAYLOAD="+j.work)
	// execute the command
	return cmd.Run()
}
