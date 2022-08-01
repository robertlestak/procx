package drivers

import (
	"errors"

	"github.com/robertlestak/procx/drivers/aws"
	"github.com/robertlestak/procx/drivers/cassandra"
	"github.com/robertlestak/procx/drivers/centauri"
	"github.com/robertlestak/procx/drivers/gcp"
	"github.com/robertlestak/procx/drivers/local"
	"github.com/robertlestak/procx/drivers/mongo"
	"github.com/robertlestak/procx/drivers/mysql"
	"github.com/robertlestak/procx/drivers/postgres"
	"github.com/robertlestak/procx/drivers/rabbitmq"
	"github.com/robertlestak/procx/drivers/redis"
)

// DriverName is the unique name of a driver.
type DriverName string

var (
	DriverAWSSQS            DriverName = "aws-sqs"
	DriverAWSDynamoDB       DriverName = "aws-dynamo"
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

// GetDriver returns the driver with the given name.
func GetDriver(name DriverName) Driver {
	switch name {
	case DriverAWSSQS:
		return &aws.SQS{}
	case DriverAWSDynamoDB:
		return &aws.Dynamo{}
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
