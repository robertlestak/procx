package qjob

import (
	"github.com/robertlestak/qjob/internal/client"
	log "github.com/sirupsen/logrus"
)

func (j *QJob) InitRedis() error {
	l := log.WithFields(log.Fields{
		"action": "InitRedis",
		"driver": j.DriverName,
	})
	l.Debug("InitRedis")
	err := client.CreateRedisClient(j.Driver.Redis.Host, j.Driver.Redis.Port, j.Driver.Redis.Password)
	if err != nil {
		return err
	}
	l.Debug("exited")
	return nil
}

func (q *QJob) handleFailureRedisList() error {
	l := log.WithFields(log.Fields{
		"action": "handleFailureRedisList",
		"driver": q.DriverName,
	})
	l.Debug("handleFailureRedisList")
	if err := client.RedisClient.RPush(q.Driver.Redis.Key, q.work).Err(); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (q *QJob) getWorkRedisList() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "getWorkRedisList",
		"driver": q.DriverName,
	})
	l.Debug("getWorkRedisList")
	m, err := client.ReceiveMessageRedisList(q.Driver.Redis.Key)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("received message")
	if m == "" {
		l.Debug("no message")
		return nil, nil
	}
	l.Debug("message received")
	return &m, nil
}

func (q *QJob) getWorkRedisSubscription() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "getWorkRedisSubscription",
		"driver": q.DriverName,
	})
	l.Debug("getWorkRedisSubscription")
	m, err := client.ReceiveMessageRedisSubscription(q.Driver.Redis.Key)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("received message")
	if m == "" {
		l.Debug("no message")
		return nil, nil
	}
	l.Debug("message received")
	return &m, nil
}
