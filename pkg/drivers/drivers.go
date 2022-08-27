package drivers

import (
	"errors"

	"github.com/robertlestak/procx/drivers/activemq"
	"github.com/robertlestak/procx/drivers/aws"
	"github.com/robertlestak/procx/drivers/cassandra"
	"github.com/robertlestak/procx/drivers/centauri"
	"github.com/robertlestak/procx/drivers/cockroach"
	"github.com/robertlestak/procx/drivers/couchbase"
	"github.com/robertlestak/procx/drivers/elasticsearch"
	"github.com/robertlestak/procx/drivers/etcd"
	"github.com/robertlestak/procx/drivers/fs"
	"github.com/robertlestak/procx/drivers/gcp"
	"github.com/robertlestak/procx/drivers/github"
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
	"github.com/robertlestak/procx/drivers/scylla"
	"github.com/robertlestak/procx/drivers/smb"
)

type DriverName string

var (
	ActiveMQ          DriverName = "activemq"
	AWSS3             DriverName = "aws-s3"
	AWSSQS            DriverName = "aws-sqs"
	AWSDynamoDB       DriverName = "aws-dynamo"
	CassandraDB       DriverName = "cassandra"
	Centauri          DriverName = "centauri"
	CockroachDB       DriverName = "cockroach"
	Couchbase         DriverName = "couchbase"
	Elasticsearch     DriverName = "elasticsearch"
	Etcd              DriverName = "etcd"
	FS                DriverName = "fs"
	HTTP              DriverName = "http"
	Kafka             DriverName = "kafka"
	GCPBQ             DriverName = "gcp-bq"
	GCPFirestore      DriverName = "gcp-firestore"
	GCPGCS            DriverName = "gcp-gcs"
	GCPPubSub         DriverName = "gcp-pubsub"
	GitHub            DriverName = "github"
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
	SMB               DriverName = "smb"
	Scylla            DriverName = "scylla"
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
	case CockroachDB:
		return &cockroach.CockroachDB{}
	case Couchbase:
		return &couchbase.Couchbase{}
	case Elasticsearch:
		return &elasticsearch.Elasticsearch{}
	case Etcd:
		return &etcd.Etcd{}
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
	case GitHub:
		return &github.GitHub{}
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
	case SMB:
		return &smb.SMB{}
	case Scylla:
		return &scylla.Scylla{}
	case Local:
		return &local.Local{}
	}
	return nil
}
