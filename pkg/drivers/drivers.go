package drivers

import (
	"errors"

	"github.com/robertlestak/procx/internal/drivers/aws"
	"github.com/robertlestak/procx/internal/drivers/cassandra"
	"github.com/robertlestak/procx/internal/drivers/centauri"
	"github.com/robertlestak/procx/internal/drivers/gcp"
	"github.com/robertlestak/procx/internal/drivers/local"
	"github.com/robertlestak/procx/internal/drivers/mongo"
	"github.com/robertlestak/procx/internal/drivers/mysql"
	"github.com/robertlestak/procx/internal/drivers/postgres"
	"github.com/robertlestak/procx/internal/drivers/rabbitmq"
	"github.com/robertlestak/procx/internal/drivers/redis"
)

type DriverName string

var (
	DriverAWSSQS            DriverName = "aws-sqs"
	DriverCassandraDB       DriverName = "cassandra"
	DriverCentauriNet       DriverName = "centauri"
	DriverGCPPubSub         DriverName = "gcp-pubsub"
	DriverPostgres          DriverName = "postgres"
	DriverMongoDB           DriverName = "mongodb"
	DriverMySQL             DriverName = "mysql"
	DriverRabbit            DriverName = "rabbitmq"
	DriverRedisSubscription DriverName = "redis-pubsub"
	DriverRedisList         DriverName = "redis-list"
	DriverLocal             DriverName = "local"
	ErrDriverNotFound                  = errors.New("driver not found")
)

func GetDriver(name DriverName) Driver {
	switch name {
	case DriverAWSSQS:
		return &aws.SQS{}
	case DriverCassandraDB:
		return &cassandra.Cassandra{}
	case DriverCentauriNet:
		return &centauri.Centauri{}
	case DriverGCPPubSub:
		return &gcp.GCPPubSub{}
	case DriverMongoDB:
		return &mongo.Mongo{}
	case DriverMySQL:
		return &mysql.Mysql{}
	case DriverPostgres:
		return &postgres.Postgres{}
	case DriverRabbit:
		return &rabbitmq.RabbitMQ{}
	case DriverRedisSubscription:
		return &redis.RedisPubSub{}
	case DriverRedisList:
		return &redis.RedisList{}
	case DriverLocal:
		return &local.Local{}
	}
	return nil
}

type Driver interface {
	LoadEnv(string) error
	LoadFlags() error
	Init() error
	GetWork() (*string, error)
	ClearWork() error
	HandleFailure() error
}
