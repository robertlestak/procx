package local

import (
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Local struct {
}

func (d *Local) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "local",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	return nil
}

func (d *Local) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "local",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	return nil
}

func (d *Local) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "local",
		"fn":  "Init",
	})
	l.Debug("Initializing local driver")
	return nil
}

func (d *Local) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "local",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from local")
	w := os.Getenv("PROCX_PAYLOAD")
	if w == "" {
		return nil, nil
	}
	return strings.NewReader(w), nil
}

func (d *Local) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "local",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from local")
	return nil
}

func (d *Local) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "local",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from local")
	return nil
}

func (d *Local) Cleanup() error {
	return nil
}
