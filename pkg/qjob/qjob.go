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
	DriverCentauriNet       DriverName = "centauri"
	DriverGCPPubSub         DriverName = "gcp-pubsub"
	DriverPostgres          DriverName = "postgres"
	DriverMongoDB           DriverName = "mongodb"
	DriverMySQL             DriverName = "mysql"
	DriverRabbit            DriverName = "rabbitmq"
	DriverRedisSubscription DriverName = "redis-pubsub"
	DriverRedisList         DriverName = "redis-list"
	DriverLocal             DriverName = "local"
	ErrDriverNotFound                  = errors.New("driver not found")
)

type DriverName string

type DriverAWS struct {
	Region      string `json:"region"`
	RoleARN     string `json:"roleArn"`
	SQSQueueURL string `json:"sqsQueueUrl"`
}

type DriverGCP struct {
	ProjectID        string          `json:"projectId"`
	SubscriptionName string          `json:"subscriptionName"`
	pubSubMessage    *pubsub.Message `json:"-"`
}

type SqlQuery struct {
	Query  string `json:"query"`
	Params []any  `json:"params"`
}

type DriverCentauri struct {
	PrivateKey []byte  `json:"privateKey"`
	PeerURL    string  `json:"peerUrl"`
	Channel    *string `json:"channel"`
	Key        *string `json:"key"`
}

type DriverPsql struct {
	Host            string    `json:"host"`
	Port            int       `json:"port"`
	User            string    `json:"user"`
	Password        string    `json:"password"`
	DBName          string    `json:"dbName"`
	SSLMode         string    `json:"sslMode"`
	QueryReturnsKey *bool     `json:"queryReturnsKey"`
	RetrieveQuery   *SqlQuery `json:"retrieveQuery"`
	FailureQuery    *SqlQuery `json:"failureQuery"`
	ClearQuery      *SqlQuery `json:"clearQuery"`
	Key             *string   `json:"key"`
}

type DriverMysql struct {
	Host            string    `json:"host"`
	Port            int       `json:"port"`
	User            string    `json:"user"`
	Password        string    `json:"password"`
	DBName          string    `json:"dbName"`
	QueryReturnsKey *bool     `json:"queryReturnsKey"`
	RetrieveQuery   *SqlQuery `json:"retrieveQuery"`
	FailureQuery    *SqlQuery `json:"failureQuery"`
	ClearQuery      *SqlQuery `json:"clearQuery"`
	Key             *string   `json:"key"`
}

type DriverMongo struct {
	Host          string  `json:"host"`
	Port          int     `json:"port"`
	User          string  `json:"user"`
	Password      string  `json:"password"`
	DBName        string  `json:"dbName"`
	Collection    string  `json:"collection"`
	RetrieveQuery *string `json:"retrieveQuery"`
	FailureQuery  *string `json:"failureQuery"`
	ClearQuery    *string `json:"clearQuery"`
	Key           *string `json:"key"`
}

type DriverRabbitMQ struct {
	URL   string `json:"url"`
	Queue string `json:"queue"`
}

type DriverRedis struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

type Driver struct {
	Name     DriverName      `json:"name"`
	AWS      *DriverAWS      `json:"aws"`
	Centauri *DriverCentauri `json:"centauri"`
	GCP      *DriverGCP      `json:"gcp"`
	Psql     *DriverPsql     `json:"psql"`
	Mongo    *DriverMongo    `json:"mongo"`
	Mysql    *DriverMysql    `json:"mysql"`
	RabbitMQ *DriverRabbitMQ `json:"rabbitmq"`
	Redis    *DriverRedis    `json:"redis"`
}

type QJob struct {
	DriverName    DriverName `json:"driverName"`
	Driver        *Driver    `json:"driver"`
	PassWorkAsArg bool       `json:"passWorkAsArg"`
	HostEnv       bool       `json:"hostEnv"`
	Bin           string     `json:"bin"`
	Args          []string   `json:"args"`
	work          string     `json:"-"`
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
	case DriverCentauriNet:
		return j.InitCentauri()
	case DriverGCPPubSub:
		return j.InitGCPPubSub()
	case DriverPostgres:
		return j.InitPsql()
	case DriverMongoDB:
		return j.InitMongo()
	case DriverMySQL:
		return j.InitMysql()
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
	case DriverCentauriNet:
		return j.getWorkCentauri()
	case DriverGCPPubSub:
		return j.getWorkGCPPubSub()
	case DriverPostgres:
		return j.getWorkPsql()
	case DriverMongoDB:
		return j.GetWorkMongo()
	case DriverMySQL:
		return j.GetWorkMysql()
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
	case DriverCentauriNet:
		return j.clearWorkCentauri()
	case DriverPostgres:
		return j.clearWorkPsql()
	case DriverMongoDB:
		return j.ClearWorkMongo()
	case DriverMySQL:
		return j.ClearWorkMysql()
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
	case DriverCentauriNet:
		return j.handleFailureCentauri()
	case DriverRedisList:
		return j.handleFailureRedisList()
	case DriverPostgres:
		return j.handleFailurePsql()
	case DriverMySQL:
		return j.HandleFailureMysql()
	case DriverMongoDB:
		return j.HandleFailureMongo()
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
