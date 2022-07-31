package local

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type Local struct {
}

func (d *Local) LoadEnv(prefix string) error {
	return nil
}

func (d *Local) LoadFlags() error {
	return nil
}

func (d *Local) Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "Init",
	})
	l.Debug("Initializing rabbitmq driver")
	return nil
}

func (d *Local) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWork",
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
		"package": "cache",
		"method":  "ClearWork",
	})
	l.Debug("Clearing work from rabbitmq")
	return nil
}

func (d *Local) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWork",
	})
	l.Debug("Clearing work from rabbitmq")
	return nil
}
