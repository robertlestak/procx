package local

import (
	"os"

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
	l.Debug("Initializing rabbitmq driver")
	return nil
}

func (d *Local) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"pkg": "local",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from rabbitmq")
	w := os.Getenv("PROCX_PAYLOAD")
	if w == "" {
		return nil, nil
	}
	return &w, nil
}

func (d *Local) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "local",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from rabbitmq")
	return nil
}

func (d *Local) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "local",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from rabbitmq")
	return nil
}

func (d *Local) Cleanup() error {
	return nil
}
