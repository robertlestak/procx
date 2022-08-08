package drivers

import (
	"errors"

	"github.com/robertlestak/procx/drivers/aws"
	"github.com/robertlestak/procx/drivers/cassandra"
	"github.com/robertlestak/procx/drivers/centauri"
	"github.com/robertlestak/procx/drivers/elasticsearch"
	"github.com/robertlestak/procx/drivers/gcp"
	"github.com/robertlestak/procx/drivers/kafka"
	"github.com/robertlestak/procx/drivers/local"
	"github.com/robertlestak/procx/drivers/mongodb"
	"github.com/robertlestak/procx/drivers/mysql"
	"github.com/robertlestak/procx/drivers/nats"
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
	DriverElasticsearch     DriverName = "elasticsearch"
	DriverKafka             DriverName = "kafka"
	DriverGCPBQ             DriverName = "gcp-bq"
	DriverGCPGCS            DriverName = "gcp-gcs"
	DriverGCPPubSub         DriverName = "gcp-pubsub"
	DriverPostgres          DriverName = "postgres"
	DriverMongoDB           DriverName = "mongodb"
	DriverMySQL             DriverName = "mysql"
	DriverNats              DriverName = "nats"
	DriverNFS               DriverName = "nfs"
	DriverRabbit            DriverName = "rabbitmq"
	DriverRedisList         DriverName = "redis-list"
	DriverRedisSubscription DriverName = "redis-pubsub"
	DriverRedisStream       DriverName = "redis-stream"
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
	case DriverElasticsearch:
		return &elasticsearch.Elasticsearch{}
	case DriverGCPBQ:
		return &gcp.BQ{}
	case DriverGCPGCS:
		return &gcp.GCS{}
	case DriverGCPPubSub:
		return &gcp.GCPPubSub{}
	case DriverKafka:
		return &kafka.Kafka{}
	case DriverMongoDB:
		return &mongodb.Mongo{}
	case DriverMySQL:
		return &mysql.Mysql{}
	case DriverNats:
		return &nats.Nats{}
	case DriverNFS:
		return &nfs.NFS{}
	case DriverPostgres:
		return &postgres.Postgres{}
	case DriverRabbit:
		return &rabbitmq.RabbitMQ{}
	case DriverRedisList:
		return &redis.RedisList{}
	case DriverRedisSubscription:
		return &redis.RedisPubSub{}
	case DriverRedisStream:
		return &redis.RedisStream{}
	case DriverLocal:
		return &local.Local{}
	}
	return nil
}
