package procx

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/robertlestak/procx/pkg/drivers"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type ProcX struct {
	DriverName      drivers.DriverName `json:"driverName"`
	Driver          drivers.Driver     `json:"driver"`
	PassWorkAsArg   bool               `json:"passWorkAsArg"`
	PassWorkAsStdin bool               `json:"passWorkAsStdin"`
	PayloadFile     string             `json:"payloadFile"`
	KeepPayloadFile bool               `json:"KeepPayloadFile"`
	HostEnv         bool               `json:"hostEnv"`
	Bin             string             `json:"bin"`
	Args            []string           `json:"args"`
	work            io.Reader          `json:"-"`
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
		"fn": "Init",
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
		"fn":     "DoWork",
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
	j.work = work
	l.Debug("work received")
	// execute
	if j.Bin == "" {
		// print work to stdout
		if _, err := io.Copy(os.Stdout, work); err != nil {
			l.WithError(err).Error("Copy")
			return err
		}
	} else {
		err = j.Exec(os.Stdout, os.Stderr)
		if err != nil {
			l.Error(err)
			if err := j.Driver.HandleFailure(); err != nil {
				l.Error(err)
			}
			return err
		}
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

func (j *ProcX) PayloadString() string {
	d, err := ioutil.ReadAll(j.work)
	if err != nil {
		return ""
	}
	j.work = ioutil.NopCloser(bytes.NewBuffer(d))
	return string(d)
}

// Exec will execute the given script, streaming the output to the provided
// io.Writers. If the script exits with a non-zero exit code, an error will be
// returned. If the script exits with a zero exit code, no error will be
// returned.
func (j *ProcX) Exec(stdout, stderr io.Writer) error {
	l := log.WithFields(log.Fields{
		"fn":     "Exec",
		"driver": j.DriverName,
	})
	l.Debug("Exec")
	// if passing work as arg, add it to args
	if j.PassWorkAsArg {
		l.Debug("passing work as arg")
		j.Args = append(j.Args, j.PayloadString())
	}
	cmd := exec.Command(j.Bin, j.Args...)
	// set the stdout and stderr pipes
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if j.HostEnv {
		l.Debug("setting host env")
		cmd.Env = os.Environ()
	}
	// if the payload file is set, set the work to the file contents
	if j.PayloadFile != "" {
		l.Debug("writing payload to file")
		f, err := os.Create(j.PayloadFile)
		if err != nil {
			l.Error(err)
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, j.work)
		if err != nil {
			l.Error(err)
			return err
		}
	}
	if j.PassWorkAsStdin {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			l.Error(err)
			return err
		}
		go func() {
			defer stdin.Close()
			io.Copy(stdin, j.work)
		}()
	}
	// if there is no payload file, and the payload is not passed as an arg nor stdin,
	// pass the payload as env var
	if j.PayloadFile == "" && !j.PassWorkAsArg && !j.PassWorkAsStdin {
		l.Debug("exporting work")
		// do not export payload to environment if output is file
		// to prevent buffer overflow in the environment on large payloads
		cmd.Env = append(cmd.Env, "PROCX_PAYLOAD="+j.PayloadString())
	}
	// execute the command
	err := cmd.Start()
	if err != nil {
		l.Error(err)
		return err
	}
	err = cmd.Wait()
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}
