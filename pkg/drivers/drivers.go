package drivers

import (
	"errors"

	"github.com/robertlestak/procx/drivers/activemq"
	"github.com/robertlestak/procx/drivers/aws"
	"github.com/robertlestak/procx/drivers/cassandra"
	"github.com/robertlestak/procx/drivers/centauri"
	"github.com/robertlestak/procx/drivers/elasticsearch"
	"github.com/robertlestak/procx/drivers/fs"
	"github.com/robertlestak/procx/drivers/gcp"
	"github.com/robertlestak/procx/drivers/http"
	"github.com/robertlestak/procx/drivers/kafka"
	"github.com/robertlestak/procx/drivers/local"
	"github.com/robertlestak/procx/drivers/mongodb"
	"github.com/robertlestak/procx/drivers/mssql"
	"github.com/robertlestak/procx/drivers/mysql"
	"github.com/robertlestak/procx/drivers/nats"
	"github.com/robertlestak/procx/drivers/nfs"
	"github.com/robertlestak/procx/drivers/nsq"
	"github.com/robertlestak/procx/drivers/postgres"
	"github.com/robertlestak/procx/drivers/pulsar"
	"github.com/robertlestak/procx/drivers/rabbitmq"
	"github.com/robertlestak/procx/drivers/redis"
)

type DriverName string

var (
	ActiveMQ          DriverName = "activemq"
	AWSS3             DriverName = "aws-s3"
	AWSSQS            DriverName = "aws-sqs"
	AWSDynamoDB       DriverName = "aws-dynamo"
	CassandraDB       DriverName = "cassandra"
	Centauri          DriverName = "centauri"
	Elasticsearch     DriverName = "elasticsearch"
	FS                DriverName = "fs"
	HTTP              DriverName = "http"
	Kafka             DriverName = "kafka"
	GCPBQ             DriverName = "gcp-bq"
	GCPFirestore      DriverName = "gcp-firestore"
	GCPGCS            DriverName = "gcp-gcs"
	GCPPubSub         DriverName = "gcp-pubsub"
	MongoDB           DriverName = "mongodb"
	MSSql             DriverName = "mssql"
	MySQL             DriverName = "mysql"
	Nats              DriverName = "nats"
	NSQ               DriverName = "nsq"
	NFS               DriverName = "nfs"
	Postgres          DriverName = "postgres"
	Pulsar            DriverName = "pulsar"
	Rabbit            DriverName = "rabbitmq"
	RedisList         DriverName = "redis-list"
	RedisSubscription DriverName = "redis-pubsub"
	RedisStream       DriverName = "redis-stream"
	Local             DriverName = "local"
	ErrDriverNotFound            = errors.New("driver not found")
)

// Get returns the driver with the given name.
func GetDriver(name DriverName) Driver {
	switch name {
	case ActiveMQ:
		return &activemq.ActiveMQ{}
	case AWSS3:
		return &aws.S3{}
	case AWSSQS:
		return &aws.SQS{}
	case AWSDynamoDB:
		return &aws.Dynamo{}
	case CassandraDB:
		return &cassandra.Cassandra{}
	case Centauri:
		return &centauri.Centauri{}
	case Elasticsearch:
		return &elasticsearch.Elasticsearch{}
	case FS:
		return &fs.FS{}
	case GCPBQ:
		return &gcp.BQ{}
	case GCPFirestore:
		return &gcp.GCPFirestore{}
	case GCPGCS:
		return &gcp.GCS{}
	case GCPPubSub:
		return &gcp.GCPPubSub{}
	case HTTP:
		return &http.HTTP{}
	case Kafka:
		return &kafka.Kafka{}
	case MongoDB:
		return &mongodb.Mongo{}
	case MSSql:
		return &mssql.MSSql{}
	case MySQL:
		return &mysql.Mysql{}
	case Nats:
		return &nats.NATS{}
	case NSQ:
		return &nsq.NSQ{}
	case NFS:
		return &nfs.NFS{}
	case Postgres:
		return &postgres.Postgres{}
	case Pulsar:
		return &pulsar.Pulsar{}
	case Rabbit:
		return &rabbitmq.RabbitMQ{}
	case RedisList:
		return &redis.RedisList{}
	case RedisSubscription:
		return &redis.RedisPubSub{}
	case RedisStream:
		return &redis.RedisStream{}
	case Local:
		return &local.Local{}
	}
	return nil
}
