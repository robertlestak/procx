package qjob

import (
	"errors"
	"io"
	"os"
	"os/exec"

	"github.com/robertlestak/qjob/internal/client"
	log "github.com/sirupsen/logrus"
)

var (
	DriverAWSSQS      DriverName = "aws-sqs"
	DriverLocal       DriverName = "local"
	ErrDriverNotFound            = errors.New("driver not found")
)

type DriverName string

type DriverAWS struct {
	Region      string
	RoleARN     string
	SQSQueueURL string
}

type Driver struct {
	Name DriverName
	AWS  *DriverAWS
}

type Driver2 interface {
	Init() error
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

func (j *QJob) InitAWSSQS() error {
	l := log.WithFields(log.Fields{
		"app": "qjob",
	})
	l.Debug("starting")
	c, err := client.CreateSQSClient(j.Driver.AWS.Region, j.Driver.AWS.RoleARN)
	if err != nil {
		return err
	}
	client.SQSClient = c
	client.SQSQueueURL = j.Driver.AWS.SQSQueueURL
	l.Debug("exited")
	return nil
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
	case DriverLocal:
		return nil
	default:
		return ErrDriverNotFound
	}
}

func (j *QJob) getWorkSQS() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "getWorkSQS",
		"driver": j.DriverName,
	})
	l.Debug("getWorkSQS")
	m, err := client.ReceiveMessageSQS()
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("received message")
	if m == nil {
		l.Debug("no message")
		return nil, nil
	}
	l.Debug("message received")
	client.SQSReceiptHandle = *m.ReceiptHandle
	return m.Body, nil
}

func (j *QJob) clearWorkSQS() error {
	l := log.WithFields(log.Fields{
		"action": "clearWorkSQS",
		"driver": j.DriverName,
	})
	l.Debug("clearWorkSQS")
	err := client.DeleteMessageSQS()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("message deleted")
	return nil
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
	case DriverLocal:
		w := os.Getenv("QJOB_PAYLOAD")
		return &w, nil
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
	case DriverLocal:
		return nil
	default:
		return ErrDriverNotFound
	}
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
