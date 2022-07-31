package procx

import (
	"io"
	"os"
	"os/exec"

	"github.com/robertlestak/procx/internal/flags"
	"github.com/robertlestak/procx/pkg/drivers"
	log "github.com/sirupsen/logrus"
)

type ProcX struct {
	DriverName    drivers.DriverName `json:"driverName"`
	Driver        drivers.Driver     `json:"driver"`
	PassWorkAsArg bool               `json:"passWorkAsArg"`
	HostEnv       bool               `json:"hostEnv"`
	Bin           string             `json:"bin"`
	Args          []string           `json:"args"`
	work          string             `json:"-"`
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

func (j *ProcX) Init(envKeyPrefix string) error {
	l := log.WithFields(log.Fields{
		"action": "Init",
	})
	l.Debug("Init")
	if j.DriverName == "" {
		l.Error("no driver specified")
		return drivers.ErrDriverNotFound
	}
	l.Debug("driver specified")
	j.Driver = drivers.GetDriver(j.DriverName)
	if j.Driver == nil {
		l.Error("driver not found")
		return drivers.ErrDriverNotFound
	}
	j.ParseArgs(flags.FlagSet.Args())
	if err := j.Driver.LoadFlags(); err != nil {
		l.WithError(err).Error("LoadFlags")
		return err
	}
	if err := j.Driver.LoadEnv(envKeyPrefix); err != nil {
		l.WithError(err).Error("LoadEnv")
		return err
	}
	if err := j.Driver.Init(); err != nil {
		l.WithError(err).Error("Init")
		return err
	}
	return nil
}

func (j *ProcX) DoWork() error {
	l := log.WithFields(log.Fields{
		"action": "DoWork",
		"driver": j.DriverName,
	})
	l.Debug("DoWork")
	// execute
	if j.Bin == "" {
		l.Error("no bin specified")
		os.Exit(1)
	}
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
