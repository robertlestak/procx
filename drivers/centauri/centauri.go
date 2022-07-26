package centauri

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"sort"

	"github.com/robertlestak/centauri/pkg/agent"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type Centauri struct {
	URL        string
	PrivateKey []byte
	Channel    *string
	Key        *string
}

func (d *Centauri) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "centauri",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	if os.Getenv(prefix+"CENTAURI_PEER_URL") != "" {
		d.URL = os.Getenv(prefix + "CENTAURI_PEER_URL")
	}
	if os.Getenv(prefix+"CENTAURI_CHANNEL") != "" {
		v := os.Getenv(prefix + "CENTAURI_CHANNEL")
		d.Channel = &v
	}
	if os.Getenv(prefix+"CENTAURI_KEY") != "" {
		v := os.Getenv(prefix + "CENTAURI_KEY")
		d.PrivateKey = []byte(v)
	}
	if os.Getenv(prefix+"CENTAURI_KEY_BASE64") != "" {
		v := os.Getenv(prefix + "CENTAURI_KEY_BASE64")
		dec, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			l.Errorf("error decoding base64: %v", err)
			return err
		}
		d.PrivateKey = dec
	}
	return nil
}

func (d *Centauri) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "centauri",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.URL = *flags.CentauriPeerURL
	d.Channel = flags.CentauriChannel
	if flags.CentauriKeyBase64 != nil && *flags.CentauriKeyBase64 != "" {
		kd, err := base64.StdEncoding.DecodeString(*flags.CentauriKeyBase64)
		if err != nil {
			l.Errorf("error decoding base64: %v", err)
			return err
		}
		d.PrivateKey = kd
	} else if flags.CentauriKey != nil && *flags.CentauriKey != "" {
		if flags.CentauriKey == nil || (flags.CentauriKey != nil && *flags.CentauriKey == "") {
			return errors.New("key required")
		}
		kd := []byte(*flags.CentauriKey)
		d.PrivateKey = kd
	}
	return nil
}

func (d *Centauri) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "centauri",
		"fn":  "Init",
	})
	l.Debug("Initializing centauri driver")
	if d.PrivateKey == nil {
		l.Error("private key is nil")
		return errors.New("private key is nil")
	}
	agent.ServerAddrs = []string{d.URL}
	if err := agent.LoadPrivateKey(d.PrivateKey); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *Centauri) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "centauri",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from centauri")
	msgs, err := agent.CheckPendingMessages(*d.Channel)
	if err != nil {
		l.Errorf("error checking pending messages: %v", err)
		return nil, err
	}
	if len(msgs) == 0 {
		l.Debug("no pending messages")
		return nil, nil
	}
	l.Debugf("pending messages: %v", msgs)
	// sort by created at
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
	})
	// get first message
	msg := msgs[0]
	id := msg.ID
	d.Key = &id
	l.Debugf("message: %v", msg)
	m, err := agent.GetMessage(*d.Channel, id)
	if err != nil {
		l.Errorf("error getting message %s: %v", id, err)
		return nil, err
	}
	if m == nil {
		l.Errorf("message %s not found", id)
		return nil, errors.New("message not found")
	}
	m, err = agent.DecryptMessageData(m)
	if err != nil {
		l.Errorf("error getting message %s: %v", id, err)
		return nil, err
	}
	l.Debugf("message: %v", m)
	l.Debugf("result: %s", string(m.Data))
	return bytes.NewReader(m.Data), nil
}

func (d *Centauri) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "centauri",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from centauri")
	if d.Key == nil || *d.Key == "" {
		l.Error("key is nil")
		return errors.New("key is nil")
	}
	err := agent.ConfirmMessageReceive(*d.Channel, *d.Key)
	if err != nil {
		l.Errorf("error deleting message: %v", err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func (d *Centauri) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "centauri",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	if d.Key == nil || *d.Key == "" {
		l.Error("key is nil")
		return errors.New("key is nil")
	}
	l.Debug("Handled failure")
	return nil
}

func (d *Centauri) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "centauri",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up")
	l.Debug("Cleaned up")
	return nil
}
