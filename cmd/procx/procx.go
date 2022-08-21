package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/robertlestak/procx/pkg/drivers"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/procx"
	log "github.com/sirupsen/logrus"
)

var (
	Version      = "dev"
	AppName      = "procx"
	EnvKeyPrefix = fmt.Sprintf("%s_", strings.ToUpper(AppName))
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func printVersion() {
	fmt.Printf(AppName+" version %s\n", Version)
}

func printUsage() {
	fmt.Printf("Usage: %s [options] [process]\n", AppName)
	flags.FlagSet.PrintDefaults()
}

func LoadEnv(prefix string) error {
	if os.Getenv(prefix+"DRIVER") != "" {
		d := os.Getenv(prefix + "DRIVER")
		flags.Driver = &d
	}
	if os.Getenv(prefix+"HOSTENV") != "" {
		h := os.Getenv(prefix + "HOSTENV")
		t := h == "true"
		flags.HostEnv = &t
	}
	if os.Getenv(prefix+"PASS_WORK_AS_ARG") != "" {
		r := os.Getenv(prefix + "PASS_WORK_AS_ARG")
		t := r == "true"
		flags.PassWorkAsArg = &t
	}
	if os.Getenv(prefix+"PASS_WORK_AS_STDIN") != "" {
		r := os.Getenv(prefix + "PASS_WORK_AS_STDIN")
		t := r == "true"
		flags.PassWorkAsStdin = &t
	}
	if os.Getenv(prefix+"DAEMON") != "" {
		r := os.Getenv(prefix + "DAEMON")
		t := r == "true"
		flags.Daemon = &t
	}
	if os.Getenv(prefix+"DAEMON_INTERVAL") != "" {
		r := os.Getenv(prefix + "DAEMON_INTERVAL")
		i, err := strconv.Atoi(r)
		if err != nil {
			return err
		}
		flags.DaemonInterval = &i
	}
	if os.Getenv(prefix+"PAYLOAD_FILE") != "" {
		r := os.Getenv(prefix + "PAYLOAD_FILE")
		flags.PayloadFile = &r
	}
	if os.Getenv(prefix+"KEEP_PAYLOAD_FILE") != "" {
		r := os.Getenv(prefix + "KEEP_PAYLOAD_FILE")
		t := r == "true"
		flags.KeepPayloadFile = &t
	}
	return nil
}

func run(j *procx.ProcX) {
	l := log.WithFields(log.Fields{
		"app": AppName,
		"fn":  "run",
	})
	l.Debug("start")
	if err := j.DoWork(); err != nil {
		l.Errorf("failed to do work: %s", err)
		os.Exit(1)
	}
	if j.PayloadFile != "" && !j.KeepPayloadFile {
		l.Debug("removing payload file")
		if err := os.Remove(j.PayloadFile); err != nil {
			l.WithError(err).Error("failed to remove payload file")
		}
	}
}

func cleanup(j *procx.ProcX) error {
	l := log.WithFields(log.Fields{
		"app": AppName,
		"fn":  "cleanup",
	})
	l.Debug("cleanup")
	if err := j.Driver.Cleanup(); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func main() {
	l := log.WithFields(log.Fields{
		"app": AppName,
		"fn":  "main",
	})
	l.Debug("start")
	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			printVersion()
			os.Exit(0)
		}
		if os.Args[1] == "--help" || os.Args[1] == "-h" {
			printUsage()
			os.Exit(0)
		}
	}
	flags.FlagSet.Parse(os.Args[1:])
	if err := LoadEnv(EnvKeyPrefix); err != nil {
		l.Error(err)
		os.Exit(1)
	}
	l.Debug("parsed flags")
	j := &procx.ProcX{
		DriverName:      drivers.DriverName(*flags.Driver),
		HostEnv:         *flags.HostEnv,
		PassWorkAsArg:   *flags.PassWorkAsArg,
		PassWorkAsStdin: *flags.PassWorkAsStdin,
		PayloadFile:     *flags.PayloadFile,
		KeepPayloadFile: *flags.KeepPayloadFile,
	}
	if err := j.Init(EnvKeyPrefix); err != nil {
		l.WithError(err).Error("InitDriver")
		os.Exit(1)
	}
	if *flags.Daemon {
		l.Debug("running as daemon")
		for {
			run(j)
			time.Sleep(time.Millisecond * time.Duration(*flags.DaemonInterval))
		}
	} else {
		run(j)
	}
	if err := cleanup(j); err != nil {
		l.WithError(err).Error("cleanup")
		os.Exit(1)
	}
	l.Debug("exited")
}
