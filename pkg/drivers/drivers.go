package drivers

import (
	"errors"

	"github.com/robertlestak/procx/drivers/aws"
	"github.com/robertlestak/procx/drivers/cassandra"
	"github.com/robertlestak/procx/drivers/centauri"
	"github.com/robertlestak/procx/drivers/gcp"
	"github.com/robertlestak/procx/drivers/local"
	"github.com/robertlestak/procx/drivers/mongodb"
	"github.com/robertlestak/procx/drivers/mysql"
	"github.com/robertlestak/procx/drivers/nfs"
	"github.com/robertlestak/procx/drivers/postgres"
	"github.com/robertlestak/procx/drivers/rabbitmq"
	"github.com/robertlestak/procx/drivers/redis"
)

// DriverName is the unique name of a driver.
type DriverName string

var (
	DriverAWSS3             DriverName = "aws-s3"
	DriverAWSSQS            DriverName = "aws-sqs"
	DriverAWSDynamoDB       DriverName = "aws-dynamo"
	DriverCassandraDB       DriverName = "cassandra"
	DriverCentauriNet       DriverName = "centauri"
	DriverGCPGCS            DriverName = "gcp-gcs"
	DriverGCPPubSub         DriverName = "gcp-pubsub"
	DriverPostgres          DriverName = "postgres"
	DriverMongoDB           DriverName = "mongodb"
	DriverMySQL             DriverName = "mysql"
	DriverNFS               DriverName = "nfs"
	DriverRabbit            DriverName = "rabbitmq"
	DriverRedisSubscription DriverName = "redis-pubsub"
	DriverRedisList         DriverName = "redis-list"
	DriverLocal             DriverName = "local"
	ErrDriverNotFound                  = errors.New("driver not found")
)

// GetDriver returns the driver with the given name.
func GetDriver(name DriverName) Driver {
	switch name {
	case DriverAWSS3:
		return &aws.S3{}
	case DriverAWSSQS:
		return &aws.SQS{}
	case DriverAWSDynamoDB:
		return &aws.Dynamo{}
	case DriverCassandraDB:
		return &cassandra.Cassandra{}
	case DriverCentauriNet:
		return &centauri.Centauri{}
	case DriverGCPGCS:
		return &gcp.GCS{}
	case DriverGCPPubSub:
		return &gcp.GCPPubSub{}
	case DriverMongoDB:
		return &mongodb.Mongo{}
	case DriverMySQL:
		return &mysql.Mysql{}
	case DriverNFS:
		return &nfs.NFS{}
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
