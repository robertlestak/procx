package client

import (
	"errors"
	"sort"

	_ "github.com/lib/pq"
	"github.com/robertlestak/centauri/pkg/agent"
	log "github.com/sirupsen/logrus"
)

func CreateCentariClient(url string, privateKey []byte) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreateCentariClient",
	})
	l.Debug("Initializing centauri client")
	if privateKey == nil {
		l.Error("private key is nil")
		return errors.New("private key is nil")
	}
	agent.ServerAddrs = []string{url}
	if err := agent.LoadPrivateKey(privateKey); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func GetWorkCentauri(channel string) (*string, *string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWorkCentauri",
	})
	l.Debug("Getting work from centauri")
	msgs, err := agent.CheckPendingMessages(channel)
	if err != nil {
		l.Errorf("error checking pending messages: %v", err)
		return nil, nil, err
	}
	if len(msgs) == 0 {
		l.Debug("no pending messages")
		return nil, nil, nil
	}
	l.Debugf("pending messages: %v", msgs)
	// sort by created at
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
	})
	// get first message
	msg := msgs[0]
	id := msg.ID
	l.Debugf("message: %v", msg)
	m, err := agent.GetMessage(channel, id)
	if err != nil {
		l.Errorf("error getting message %s: %v", id, err)
		return nil, nil, err
	}
	if m == nil {
		l.Errorf("message %s not found", id)
		return nil, nil, errors.New("message not found")
	}
	m, err = agent.DecryptMessageData(m)
	if err != nil {
		l.Errorf("error getting message %s: %v", id, err)
		return nil, nil, err
	}
	var result string
	if m.Data != nil {
		result = string(m.Data)
	}
	return &result, &id, nil
}

func ClearWorkCentauri(channel string, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkCentauri",
	})
	l.Debug("Clearing work from centauri")
	if key == nil || *key == "" {
		l.Error("key is nil")
		return errors.New("key is nil")
	}
	err := agent.ConfirmMessageReceive(channel, *key)
	if err != nil {
		l.Errorf("error deleting message: %v", err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func HandleFailureCentauri(channel string, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "HandleFailureCentauri",
	})
	l.Debug("handling failure for centauri")
	if key == nil || *key == "" {
		l.Error("key is nil")
		return errors.New("key is nil")
	}
	l.Debug("handled failure")
	return nil
}
