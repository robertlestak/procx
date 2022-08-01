package gcp

import (
	"context"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type GCPPubSub struct {
	Client           *pubsub.Client
	ProjectID        string
	SubscriptionName string
}

func (d *GCPPubSub) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "LoadEnv",
	})
	l.Debug("LoadEnv")
	if os.Getenv(prefix+"GCP_PROJECT_ID") != "" {
		d.ProjectID = os.Getenv(prefix + "GCP_PROJECT_ID")
	}
	if os.Getenv(prefix+"GCP_SUBSCRIPTION") != "" {
		d.SubscriptionName = os.Getenv(prefix + "GCP_SUBSCRIPTION")
	}
	return nil
}

func (d *GCPPubSub) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.ProjectID = *flags.GCPProjectID
	d.SubscriptionName = *flags.GCPSubscription
	return nil
}

func (d *GCPPubSub) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "Init",
	})
	l.Debug("Initializing gcp pubsub driver")
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, d.ProjectID)
	if err != nil {
		return err
	}
	d.Client = client
	return nil
}

func (d *GCPPubSub) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from gcp pubsub driver")
	ctx := context.Background()
	sub := d.Client.Subscription(d.SubscriptionName)
	var msgData *pubsub.Message
	msgChan := make(chan *pubsub.Message)
	// receive first message from subscription
	go func() {
		err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
			//m.Ack()
			msgChan <- m
		})
		if err != nil {
			log.Error(err)
		}
	}()
	msgData = <-msgChan
	if msgData == nil {
		return nil, nil
	}
	sd := string(msgData.Data)
	return &sd, nil
}

func (d *GCPPubSub) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from gcp pubsub driver")
	return nil
}

func (d *GCPPubSub) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure in gcp pubsub driver")
	return nil
}

func (d *GCPPubSub) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up gcp pubsub driver")
	return nil
}
