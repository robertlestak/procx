package gcp

import (
	"context"

	"cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"
)

type GCPPubSub struct {
	Client           *pubsub.Client
	ProjectID        string
	SubscriptionName string
}

func (d *GCPPubSub) Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "Init",
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
	return nil
}

func (d *GCPPubSub) HandleFailure() error {
	return nil
}
