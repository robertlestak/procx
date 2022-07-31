package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type RedisPubSub struct {
	Client   *redis.Client
	Host     string
	Port     string
	Password string
	Key      string
}

func (d *RedisPubSub) Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "Init",
	})
	l.Debug("Initializing redis pub/sub driver")
	d.Client = redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%s", d.Host, d.Port),
		Password:    d.Password, // no password set
		DB:          0,          // use default DB
		DialTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Second,
	})
	cmd := d.Client.Ping()
	if cmd.Err() != nil {
		l.Error("Failed to connect to redis")
		return cmd.Err()
	}
	l.Debug("Connected to redis")
	return nil
}

func (d *RedisPubSub) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWork",
	})
	l.Debug("Getting work from redis pub/sub")
	l.Debug("Receiving message from redis subscription")
	sub := d.Client.Subscribe(d.Key)
	defer sub.Close()
	for {
		msg, err := sub.ReceiveMessage()
		if err != nil {
			// If the queue is empty, return nil
			if err == redis.Nil {
				l.Debug("Queue is empty")
				return nil, nil
			}
			l.WithError(err).Error("Failed to receive message")
			continue
		}
		l.Debug("Received message")
		return &msg.Payload, nil
	}
}

func (d *RedisPubSub) ClearWork() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWork",
	})
	l.Debug("Clearing work from redis pub/sub")
	return nil
}

func (d *RedisPubSub) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "HandleFailure",
	})
	l.Debug("Handling failure")
	return nil
}
