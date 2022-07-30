package client

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

var (
	RedisClient *redis.Client
)

func CreateRedisClient(host, port, password string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreateRedisClient",
		"host":    host,
		"port":    port,
	})
	l.Debug("Initializing redis client")
	RedisClient = redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%s", host, port),
		Password:    password, // no password set
		DB:          0,        // use default DB
		DialTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Second,
	})
	cmd := RedisClient.Ping()
	if cmd.Err() != nil {
		l.Error("Failed to connect to redis")
		return cmd.Err()
	}
	l.Debug("Connected to redis")
	return nil
}

func ReceiveMessageRedisSubscription(channel string) (string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ReceiveMessageRedisSubscription",
		"channel": channel,
	})
	l.Debug("Receiving message from redis subscription")
	sub := RedisClient.Subscribe(channel)
	defer sub.Close()
	for {
		msg, err := sub.ReceiveMessage()
		if err != nil {
			// If the queue is empty, return nil
			if err == redis.Nil {
				l.Debug("Queue is empty")
				return "", nil
			}
			l.WithError(err).Error("Failed to receive message")
			continue
		}
		l.Debug("Received message")
		return msg.Payload, nil
	}
}

func ReceiveMessageRedisList(queue string) (string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ReceiveMessageRedisList",
		"queue":   queue,
	})
	l.Debug("Receiving message from redis list")
	msg, err := RedisClient.LPop(queue).Result()
	if err != nil {
		// If the queue is empty, return nil
		if err == redis.Nil {
			l.Debug("Queue is empty")
			return "", nil
		}
		l.WithError(err).Error("Failed to receive message")
		return "", err
	}
	l.Debug("Received message")
	return msg, nil
}
