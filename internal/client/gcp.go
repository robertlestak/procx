package client

import (
	"context"

	"cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"
)

var (
	GCPPubSubClient *pubsub.Client
)

func CreateGCPPubSubClient(projectID string) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	GCPPubSubClient = client
	return nil
}

func ReceiveMessageGCPPubSub(subName string) (*pubsub.Message, error) {
	ctx := context.Background()
	sub := GCPPubSubClient.Subscription(subName)
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
	return msgData, nil
}
