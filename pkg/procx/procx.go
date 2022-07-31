package procx

import (
	"errors"
	"io"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type DriverName string

var (
	DriverAWSSQS            DriverName = "aws-sqs"
	DriverCassandraDB       DriverName = "cassandra"
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

type SqlQuery struct {
	Query  string `json:"query"`
	Params []any  `json:"params"`
}

type Driver interface {
	Init() error
	GetWork() (*string, error)
	ClearWork() error
	HandleFailure() error
}

type ProcX struct {
	DriverName    DriverName `json:"driverName"`
	Driver        Driver     `json:"driver2"`
	PassWorkAsArg bool       `json:"passWorkAsArg"`
	HostEnv       bool       `json:"hostEnv"`
	Bin           string     `json:"bin"`
	Args          []string   `json:"args"`
	work          string     `json:"-"`
}

func (j *ProcX) ParseArgs(args []string) {
	if len(args) == 0 {
		return
	}
	j.Bin = args[0]
	if len(args) > 1 {
		j.Args = args[1:]
	}
}

func (j *ProcX) DoWork() error {
	l := log.WithFields(log.Fields{
		"action": "DoWork",
		"driver": j.DriverName,
	})
	l.Debug("DoWork")
	work, err := j.Driver.GetWork()
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
		if err := j.Driver.HandleFailure(); err != nil {
			l.Error(err)
		}
		return err
	}
	l.Debug("work completed")
	err = j.Driver.ClearWork()
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
func (j *ProcX) Exec(stdout, stderr io.Writer) error {
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
	cmd.Env = append(cmd.Env, "PROCX_PAYLOAD="+j.work)
	// execute the command
	return cmd.Run()
}
