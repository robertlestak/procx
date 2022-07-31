package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/robertlestak/procx/pkg/procx"
	log "github.com/sirupsen/logrus"
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func printVersion() {
	fmt.Printf("procx version %s\n", Version)
}

func runOnce() {
	l := log.WithFields(log.Fields{
		"app": "procx",
	})
	l.Debug("starting")
	args := flag.Args()
	j := &procx.ProcX{
		DriverName:    procx.DriverName(*flagDriver),
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
		"app": "procx",
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
