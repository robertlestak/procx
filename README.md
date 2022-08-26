# procx - cloud agnostic process execution

procx is a small process manager that can wrap around any existing application / process, and integrate with a job queue or data persistence layer to enable autoscaling job executions with no native code integration.

procx is a single compiled binary that can be packaged in your existing job code container, configured with environment variables or command line flags, and started as an entrypoint for any application.

The application will receive the work data agnostic to the data provider, with no need to integrate with the various upstream data sources. Data sources can be changed by simply changing procx's start up variables, and the application remains exactly the same.

Build once, run anywhere.

## Execution

procx is started with the path to the process to manage, and either environment variables or a set of command line flags.

```bash
# cli args
procx -driver redis-list ... /path/to/process
# or, env vars
export PROCX_DRIVER=redis-list
...
procx /path/to/process
```

procx will retrieve the next job from the queue and pass it to the process. Upon success (exit code 0), procx will mark the job as complete. Upon failure (exit code != 0), procx will mark the job as failed to be requeued.

By default, the subprocess spawned by procx will not have access to the host environment variables. This can be changed by setting the `-hostenv` flag.

By default, procx will connect to the data source, consume a single message, and then exit when the spawned process exits. If the `-daemon` flag is set, procx will connect to the data source and consume messages until the process is killed, or until a job fails.

### Payload

By default, procx will export the payload as an environment variable `PROCX_PAYLOAD`. If `-pass-work-as-arg` is set, the job payload string will be appended to the process arguments, and if `-pass-work-as-stdin` is set, the job payload will be piped to stdin of the process. Finally, if the `-payload-file` flag is set, the payload will be written to the specified file path. procx will clean up the file at the end of the job, unless you pass `-keep-payload-file`.

If no process is passed to `procx`, the payload will be printed to stdout.

### Relational Driver JSON Parsing

For drivers which are non-structured (ex. `fs`, `aws-s3`, `redis-list`, etc.), procx will pass the payload data as-is to the driver. However for drivers which enforce some relational schema such as SQL-based drivers, you will need to provide a query which will be run to retrieve the data, and optionally queries to run if the work completes successfully or fails. procx will parse the query output into an array of JSON objects and pass it to the process. You can select a specific JSON field by passing the driver's respective `-{driver}-retrieve-field` flag. You can then use `{{mustache}}` syntax to extract specific fields from the returned data and use them in your subsequent clear and fail queries. For example:

```bash
procx -driver postgres \
    ... \
    -psql-retrieve-query "SELECT id,name,work FROM jobs WHERE status=$1 LIMIT 1" \
    -psql-retrieve-params "pending" \
    -psql-clear-query "UPDATE jobs SET status=$1 WHERE id=$2" \
    -psql-clear-params "complete,{{0.id}}" \
    -pass-work-as-stdin \
    cat
# the above will print
# [{"id":1,"name":"John Doe","work":"This is my work"}]
# however if we use the -psql-retrieve-field=0.work flag, we can extract the 0'th work field, to just print: "This is my work"
```

## Drivers

Currently, the following drivers are supported:

- [ActiveMQ](#activemq) (`activemq`)
- [AWS DynamoDB](#aws-dynamodb) (`aws-dynamo`)
- [AWS S3](#aws-s3) (`aws-s3`)
- [AWS SQS](#aws-sqs) (`aws-sqs`)
- [Cassandra](#cassandra) (`cassandra`)
- [Centauri](#centauri) (`centauri`)
- [Cockroach](#cockroach) (`cockroach`)
- [Couchbase](#couchbase) (`couchbase`)
- [Elasticsearch](#elasticsearch) (`elasticsearch`)
- [FS](#fs) (`fs`)
- [GCP Big Query](#gcp-bq) (`gcp-bq`)
- [GCP Cloud Storage](#gcp-gcs) (`gcp-gcs`)
- [GCP Firestore](#gcp-firestore) (`gcp-firestore`)
- [GCP Pub/Sub](#gcp-pubsub) (`gcp-pubsub`)
- [GitHub](#github) (`github`)
- [HTTP](#http) (`http`)
- [Kafka](#kafka) (`kafka`)
- [PostgreSQL](#postgresql) (`postgres`)
- [Pulsar](#pulsar) (`pulsar`)
- [MongoDB](#mongodb) (`mongodb`)
- [MSSQL](#mssql) (`mssql`)
- [MySQL](#mysql) (`mysql`)
- [NATS](#nats) (`nats`)
- [NFS](#nfs) (`nfs`)
- [NSQ](#nsq) (`nsq`)
- [RabbitMQ](#rabbitmq) (`rabbitmq`)
- [Redis List](#redis-list) (`redis-list`)
- [Redis Pub/Sub](#redis-pubsub) (`redis-pubsub`)
- [Redis Stream](#redis-stream) (`redis-stream`)
- [Scylla](#scylla) (`scylla`)
- [SMB](#smb) (`smb`)
- [Local](#local) (`local`)

Plans to add more drivers in the future, and PRs are welcome.

See [Driver Examples](#driver-examples) for more information.

## Install

```bash
curl -SsL https://raw.githubusercontent.com/robertlestak/procx/main/scripts/install.sh | bash -e
```

### A note on permissions

Depending on the path of `INSTALL_DIR` and the permissions of the user running the installation script, you may get a Permission Denied error if you are trying to move the binary into a location which your current user does not have access to. This is most often the case when running the script as a non-root user yet trying to install into `/usr/local/bin`. To fix this, you can either:

Create a `$HOME/bin` directory in your current user home directory. This will be the default installation directory. Be sure to add this to your `$PATH` environment variable.

Use `sudo` to run the installation script, to install into `/usr/local/bin` 

```bash
curl -SsL https://raw.githubusercontent.com/robertlestak/procx/main/scripts/install.sh | sudo bash -e
```

### Build From Source

```bash
mkdir -p bin
go build -o bin/procx cmd/procx/*.go
```

#### Building for a Specific Driver

By default, the `procx` binary is compiled for all drivers. This is to enable a truly build-once-run-anywhere experience. However some users may want a smaller binary for embedded workloads. To enable this, you can run `make listdrivers` to get the full list of available drivers, and `make slim drivers="driver1 driver2 driver3 ..."` - listing each driver separated by a space - to build a slim binary with just the specified driver(s).

While building for a specific driver may seem contrary to the ethos of procx, the decoupling between the job queue and work still enables a write-once-run-anywhere experience, and simply requires DevOps to rebuild the image with your new drivers if you are shifting upstream data sources.

## Usage

```bash
Usage: procx [options] [process]
  -activemq-address string
    	ActiveMQ STOMP address
  -activemq-enable-tls
    	Enable TLS
  -activemq-name string
    	ActiveMQ name
  -activemq-tls-ca-file string
    	TLS CA
  -activemq-tls-cert-file string
    	TLS cert
  -activemq-tls-insecure
    	Enable TLS insecure
  -activemq-tls-key-file string
    	TLS key
  -activemq-type string
    	ActiveMQ type. Valid values are: topic, queue
  -aws-dynamo-clear-query string
    	AWS DynamoDB clear query
  -aws-dynamo-fail-query string
    	AWS DynamoDB fail query
  -aws-dynamo-include-next-token
    	AWS DynamoDB include next token as _nextToken in response
  -aws-dynamo-limit int
    	AWS DynamoDB limit
  -aws-dynamo-next-token string
    	AWS DynamoDB next token
  -aws-dynamo-retrieve-field string
    	AWS DynamoDB retrieve field
  -aws-dynamo-retrieve-query string
    	AWS DynamoDB retrieve query
  -aws-dynamo-unmarshal-json
    	AWS DynamoDB unmarshal JSON (default true)
  -aws-load-config
    	load AWS config from ~/.aws/config
  -aws-region string
    	AWS region
  -aws-role-arn string
    	AWS role ARN
  -aws-s3-bucket string
    	AWS S3 bucket
  -aws-s3-clear-bucket string
    	AWS S3 clear bucket, if clear op is mv
  -aws-s3-clear-key string
    	AWS S3 clear key, if clear op is mv. default is origional key name.
  -aws-s3-clear-key-template string
    	AWS S3 clear key template, if clear op is mv.
  -aws-s3-clear-op string
    	AWS S3 clear operation. Valid values: mv, rm
  -aws-s3-fail-bucket string
    	AWS S3 fail bucket, if fail op is mv
  -aws-s3-fail-key string
    	AWS S3 fail key, if fail op is mv. default is original key name.
  -aws-s3-fail-key-template string
    	AWS S3 fail key template, if fail op is mv.
  -aws-s3-fail-op string
    	AWS S3 fail operation. Valid values: mv, rm
  -aws-s3-key string
    	AWS S3 key
  -aws-s3-key-prefix string
    	AWS S3 key prefix
  -aws-s3-key-regex string
    	AWS S3 key regex
  -aws-sqs-queue-url string
    	AWS SQS queue URL
  -cassandra-clear-params string
    	Cassandra clear params
  -cassandra-clear-query string
    	Cassandra clear query
  -cassandra-consistency string
    	Cassandra consistency (default "QUORUM")
  -cassandra-fail-params string
    	Cassandra fail params
  -cassandra-fail-query string
    	Cassandra fail query
  -cassandra-hosts string
    	Cassandra hosts
  -cassandra-keyspace string
    	Cassandra keyspace
  -cassandra-password string
    	Cassandra password
  -cassandra-retrieve-field string
    	Cassandra retrieve field. If not set, all fields will be returned as a JSON object
  -cassandra-retrieve-params string
    	Cassandra retrieve params
  -cassandra-retrieve-query string
    	Cassandra retrieve query
  -cassandra-user string
    	Cassandra user
  -centauri-channel string
    	Centauri channel (default "default")
  -centauri-key string
    	Centauri key
  -centauri-key-base64 string
    	Centauri key base64
  -centauri-peer-url string
    	Centauri peer URL
  -cockroach-clear-params string
    	CockroachDB clear params
  -cockroach-clear-query string
    	CockroachDB clear query
  -cockroach-database string
    	CockroachDB database
  -cockroach-fail-params string
    	CockroachDB fail params
  -cockroach-fail-query string
    	CockroachDB fail query
  -cockroach-host string
    	CockroachDB host
  -cockroach-password string
    	CockroachDB password
  -cockroach-port string
    	CockroachDB port (default "26257")
  -cockroach-query-key
    	CockroachDB query returns key as first column and value as second column
  -cockroach-retrieve-field string
    	CockroachDB retrieve field. If not set, all fields will be returned as a JSON object
  -cockroach-retrieve-params string
    	CockroachDB retrieve params
  -cockroach-retrieve-query string
    	CockroachDB retrieve query
  -cockroach-routing-id string
    	CockroachDB routing id
  -cockroach-ssl-mode string
    	CockroachDB SSL mode (default "disable")
  -cockroach-tls-cert string
    	CockroachDB TLS cert
  -cockroach-tls-key string
    	CockroachDB TLS key
  -cockroach-tls-root-cert string
    	CockroachDB SSL root cert
  -cockroach-user string
    	CockroachDB user
  -couchbase-address string
    	Couchbase address
  -couchbase-bucket string
    	Couchbase bucket name
  -couchbase-clear-bucket string
    	Couchbase clear bucket, if op is set or merge. Default to the current bucket.
  -couchbase-clear-collection string
    	Couchbase clear collection, default to the current collection. (default "_default")
  -couchbase-clear-doc string
    	Couchbase clear doc, if op is set or merge
  -couchbase-clear-id string
    	Couchbase clear id, default to the current id.
  -couchbase-clear-op string
    	Couchbase clear op. one of: mv, rm, set, merge
  -couchbase-clear-scope string
    	Couchbase clear scope, default to the current scope. (default "_default")
  -couchbase-collection string
    	Couchbase collection (default "_default")
  -couchbase-enable-tls
    	Enable TLS
  -couchbase-fail-bucket string
    	Couchbase fail bucket, if op is set or merge. Default to the current bucket.
  -couchbase-fail-collection string
    	Couchbase fail collection, default to the current collection. (default "_default")
  -couchbase-fail-doc string
    	Couchbase fail doc, if op is set or merge
  -couchbase-fail-id string
    	Couchbase fail id, default to the current id.
  -couchbase-fail-op string
    	Couchbase fail op. one of: mv, rm, set, merge
  -couchbase-fail-scope string
    	Couchbase fail scope, default to the current scope. (default "_default")
  -couchbase-id string
    	Couchbase id
  -couchbase-password string
    	Couchbase password
  -couchbase-retrieve-params string
    	Couchbase retrieve params
  -couchbase-retrieve-query string
    	Couchbase retrieve query
  -couchbase-scope string
    	Couchbase scope (default "_default")
  -couchbase-tls-ca-file string
    	Couchbase TLS CA file
  -couchbase-tls-cert-file string
    	Couchbase TLS cert file
  -couchbase-tls-insecure
    	Enable TLS insecure
  -couchbase-tls-key-file string
    	Couchbase TLS key file
  -couchbase-user string
    	Couchbase user
  -daemon
    	run as daemon
  -daemon-interval int
    	daemon interval in milliseconds
  -driver string
    	driver to use. (activemq, aws-dynamo, aws-s3, aws-sqs, cassandra, centauri, cockroach, couchbase, elasticsearch, fs, gcp-bq, gcp-firestore, gcp-gcs, gcp-pubsub, github, http, kafka, local, mongodb, mssql, mysql, nats, nfs, nsq, postgres, pulsar, rabbitmq, redis-list, redis-pubsub, redis-stream, smb)
  -elasticsearch-address string
    	Elasticsearch address
  -elasticsearch-clear-doc string
    	Elasticsearch clear doc
  -elasticsearch-clear-index string
    	Elasticsearch clear index
  -elasticsearch-clear-op string
    	Elasticsearch clear op. Valid values are: delete, put, merge-put, move
  -elasticsearch-enable-tls
    	Elasticsearch enable TLS
  -elasticsearch-fail-doc string
    	Elasticsearch fail doc
  -elasticsearch-fail-index string
    	Elasticsearch fail index
  -elasticsearch-fail-op string
    	Elasticsearch fail op. Valid values are: delete, put, merge-put, move
  -elasticsearch-password string
    	Elasticsearch password
  -elasticsearch-retrieve-index string
    	Elasticsearch retrieve index
  -elasticsearch-retrieve-query string
    	Elasticsearch retrieve query
  -elasticsearch-tls-ca-file string
    	Elasticsearch TLS CA file
  -elasticsearch-tls-cert-file string
    	Elasticsearch TLS cert file
  -elasticsearch-tls-key-file string
    	Elasticsearch TLS key file
  -elasticsearch-tls-skip-verify
    	Elasticsearch TLS skip verify
  -elasticsearch-username string
    	Elasticsearch username
  -fs-clear-folder string
    	FS clear folder, if clear op is mv
  -fs-clear-key string
    	FS clear key, if clear op is mv. default is origional key name.
  -fs-clear-key-template string
    	FS clear key template, if clear op is mv.
  -fs-clear-op string
    	FS clear operation. Valid values: mv, rm
  -fs-fail-folder string
    	FS fail folder, if fail op is mv
  -fs-fail-key string
    	FS fail key, if fail op is mv. default is original key name.
  -fs-fail-key-template string
    	FS fail key template, if fail op is mv.
  -fs-fail-op string
    	FS fail operation. Valid values: mv, rm
  -fs-folder string
    	FS folder
  -fs-key string
    	FS key
  -fs-key-prefix string
    	FS key prefix
  -fs-key-regex string
    	FS key regex
  -gcp-bq-clear-query string
    	GCP BQ clear query
  -gcp-bq-fail-query string
    	GCP BQ fail query
  -gcp-bq-retrieve-field string
    	GCP BigQuery retrieve field
  -gcp-bq-retrieve-query string
    	GCP BQ retrieve query
  -gcp-firestore-clear-collection string
    	GCP Firestore clear collection
  -gcp-firestore-clear-op string
    	GCP Firestore clear op. Possible values: 'mv', 'rm', 'update'
  -gcp-firestore-clear-update string
    	GCP Firestore clear update object. Will be merged with document before update or move.
  -gcp-firestore-fail-collection string
    	GCP Firestore fail collection
  -gcp-firestore-fail-op string
    	GCP Firestore fail op. Possible values: 'mv', 'rm', 'update'
  -gcp-firestore-fail-update string
    	GCP Firestore fail update object. Will be merged with document before update or move.
  -gcp-firestore-retrieve-collection string
    	GCP Firestore retrieve collection
  -gcp-firestore-retrieve-document string
    	GCP Firestore retrieve document
  -gcp-firestore-retrieve-document-json-key string
    	GCP Firestore retrieve document JSON key
  -gcp-firestore-retrieve-query-op string
    	GCP Firestore retrieve query op
  -gcp-firestore-retrieve-query-order string
    	GCP Firestore retrieve query order. Valid values: asc, desc
  -gcp-firestore-retrieve-query-order-by string
    	GCP Firestore retrieve query order by key
  -gcp-firestore-retrieve-query-path string
    	GCP Firestore retrieve query path
  -gcp-firestore-retrieve-query-value string
    	GCP Firestore retrieve query value
  -gcp-gcs-bucket string
    	GCP GCS bucket
  -gcp-gcs-clear-bucket string
    	GCP GCS clear bucket, if clear op is mv
  -gcp-gcs-clear-key string
    	GCP GCS clear key, if clear op is mv. default is origional key name.
  -gcp-gcs-clear-key-template string
    	GCP GCS clear key template, if clear op is mv.
  -gcp-gcs-clear-op string
    	GCP GCS clear operation. Valid values: mv, rm
  -gcp-gcs-fail-bucket string
    	GCP GCS fail bucket, if fail op is mv
  -gcp-gcs-fail-key string
    	GCP GCS fail key, if fail op is mv. default is original key name.
  -gcp-gcs-fail-key-template string
    	GCP GCS fail key template, if fail op is mv.
  -gcp-gcs-fail-op string
    	GCP GCS fail operation. Valid values: mv, rm
  -gcp-gcs-key string
    	GCP GCS key
  -gcp-gcs-key-prefix string
    	GCP GCS key prefix
  -gcp-gcs-key-regex string
    	GCP GCS key regex
  -gcp-project-id string
    	GCP project ID
  -gcp-pubsub-subscription string
    	GCP Pub/Sub subscription name
  -github-base-branch string
    	base branch for PR
  -github-branch string
    	branch for PR.
  -github-clear-location string
    	clear operation location, if op is mv
  -github-clear-op string
    	clear operation. One of: [mv, rm]
  -github-commit-email string
    	commit email
  -github-commit-message string
    	commit message
  -github-commit-name string
    	commit name
  -github-fail-location string
    	fail operation location, if op is mv
  -github-fail-op string
    	fail operation. One of: [mv, rm]
  -github-file string
    	GitHub file
  -github-file-prefix string
    	GitHub file prefix
  -github-file-regex string
    	GitHub file regex
  -github-open-pr
    	open PR on changes. Default: false
  -github-owner string
    	GitHub owner
  -github-pr-body string
    	PR body
  -github-pr-title string
    	PR title
  -github-ref string
    	GitHub ref
  -github-repo string
    	GitHub repo
  -github-token string
    	GitHub token
  -hostenv
    	use host environment
  -http-clear-body string
    	HTTP clear body
  -http-clear-body-file string
    	HTTP clear body file
  -http-clear-content-type string
    	HTTP clear content type
  -http-clear-headers string
    	HTTP clear headers
  -http-clear-method string
    	HTTP clear method (default "GET")
  -http-clear-successful-status-codes string
    	HTTP clear successful status codes
  -http-clear-url string
    	HTTP clear url
  -http-enable-tls
    	HTTP enable tls
  -http-fail-body string
    	HTTP fail body
  -http-fail-body-file string
    	HTTP fail body file
  -http-fail-content-type string
    	HTTP fail content type
  -http-fail-headers string
    	HTTP fail headers
  -http-fail-method string
    	HTTP fail method (default "GET")
  -http-fail-successful-status-codes string
    	HTTP fail successful status codes
  -http-fail-url string
    	HTTP fail url
  -http-retrieve-body string
    	HTTP retrieve body
  -http-retrieve-body-file string
    	HTTP retrieve body file
  -http-retrieve-content-type string
    	HTTP retrieve content type
  -http-retrieve-headers string
    	HTTP retrieve headers
  -http-retrieve-key-json-selector string
    	HTTP retrieve key json selector
  -http-retrieve-method string
    	HTTP retrieve method (default "GET")
  -http-retrieve-successful-status-codes string
    	HTTP retrieve successful status codes
  -http-retrieve-url string
    	HTTP retrieve url
  -http-retrieve-work-json-selector string
    	HTTP retrieve work json selector
  -http-tls-ca-file string
    	HTTP tls ca file
  -http-tls-cert-file string
    	HTTP tls cert file
  -http-tls-insecure
    	HTTP tls insecure
  -http-tls-key-file string
    	HTTP tls key file
  -kafka-brokers string
    	Kafka brokers, comma separated
  -kafka-enable-sasl
    	Enable SASL
  -kafka-enable-tls
    	Enable TLS
  -kafka-group string
    	Kafka group
  -kafka-sasl-password string
    	Kafka SASL password
  -kafka-sasl-type string
    	Kafka SASL type. Can be either 'scram' or 'plain'
  -kafka-sasl-username string
    	Kafka SASL user
  -kafka-tls-ca-file string
    	Kafka TLS CA file
  -kafka-tls-cert-file string
    	Kafka TLS cert file
  -kafka-tls-insecure
    	Enable TLS insecure
  -kafka-tls-key-file string
    	Kafka TLS key file
  -kafka-topic string
    	Kafka topic
  -keep-payload-file
    	keep payload file after processing
  -mongo-auth-source string
    	MongoDB auth source
  -mongo-clear-query string
    	MongoDB clear query
  -mongo-collection string
    	MongoDB collection
  -mongo-database string
    	MongoDB database
  -mongo-enable-tls
    	Enable TLS
  -mongo-fail-query string
    	MongoDB fail query
  -mongo-host string
    	MongoDB host
  -mongo-limit int
    	MongoDB limit
  -mongo-password string
    	MongoDB password
  -mongo-port string
    	MongoDB port (default "27017")
  -mongo-retrieve-query string
    	MongoDB retrieve query
  -mongo-tls-ca-file string
    	Mongo TLS CA file
  -mongo-tls-cert-file string
    	Mongo TLS cert file
  -mongo-tls-insecure
    	Enable TLS insecure
  -mongo-tls-key-file string
    	Mongo TLS key file
  -mongo-user string
    	MongoDB user
  -mssql-clear-params string
    	MSSQL clear params
  -mssql-clear-query string
    	MSSQL clear query
  -mssql-database string
    	MSSQL database
  -mssql-fail-params string
    	MSSQL fail params
  -mssql-fail-query string
    	MSSQL fail query
  -mssql-host string
    	MSSQL host
  -mssql-password string
    	MSSQL password
  -mssql-port string
    	MSSQL port (default "1433")
  -mssql-retrieve-field string
    	MSSQL retrieve field. If not set, all fields will be returned as a JSON object
  -mssql-retrieve-params string
    	MSSQL retrieve params
  -mssql-retrieve-query string
    	MSSQL retrieve query
  -mssql-user string
    	MSSQL user
  -mysql-clear-params string
    	MySQL clear params
  -mysql-clear-query string
    	MySQL clear query
  -mysql-database string
    	MySQL database
  -mysql-fail-params string
    	MySQL fail params
  -mysql-fail-query string
    	MySQL fail query
  -mysql-host string
    	MySQL host
  -mysql-password string
    	MySQL password
  -mysql-port string
    	MySQL port (default "3306")
  -mysql-retrieve-field string
    	MySQL retrieve field. If not set, all fields will be returned as a JSON object
  -mysql-retrieve-params string
    	MySQL retrieve params
  -mysql-retrieve-query string
    	MySQL retrieve query
  -mysql-user string
    	MySQL user
  -nats-clear-response string
    	NATS clear response
  -nats-creds-file string
    	NATS creds file
  -nats-enable-tls
    	NATS enable TLS
  -nats-fail-response string
    	NATS fail response
  -nats-jwt-file string
    	NATS JWT file
  -nats-nkey-file string
    	NATS NKey file
  -nats-password string
    	NATS password
  -nats-queue-group string
    	NATS queue group
  -nats-subject string
    	NATS subject
  -nats-tls-ca-file string
    	NATS TLS CA file
  -nats-tls-cert-file string
    	NATS TLS cert file
  -nats-tls-insecure
    	NATS TLS insecure
  -nats-tls-key-file string
    	NATS TLS key file
  -nats-token string
    	NATS token
  -nats-url string
    	NATS URL
  -nats-username string
    	NATS username
  -nfs-clear-folder string
    	NFS clear folder, if clear op is mv
  -nfs-clear-key string
    	NFS clear key, if clear op is mv. default is origional key name.
  -nfs-clear-key-template string
    	NFS clear key template, if clear op is mv.
  -nfs-clear-op string
    	NFS clear operation. Valid values: mv, rm
  -nfs-fail-folder string
    	NFS fail folder, if fail op is mv
  -nfs-fail-key string
    	NFS fail key, if fail op is mv. default is original key name.
  -nfs-fail-key-template string
    	NFS fail key template, if fail op is mv.
  -nfs-fail-op string
    	NFS fail operation. Valid values: mv, rm
  -nfs-folder string
    	NFS folder
  -nfs-host string
    	NFS host
  -nfs-key string
    	NFS key
  -nfs-key-prefix string
    	NFS key prefix
  -nfs-key-regex string
    	NFS key regex
  -nfs-target string
    	NFS target
  -nsq-channel string
    	NSQ channel
  -nsq-enable-tls
    	Enable TLS
  -nsq-nsqd-address string
    	NSQ nsqd address
  -nsq-nsqlookupd-address string
    	NSQ nsqlookupd address
  -nsq-tls-ca-file string
    	NSQ TLS CA file
  -nsq-tls-cert-file string
    	NSQ TLS cert file
  -nsq-tls-key-file string
    	NSQ TLS key file
  -nsq-tls-skip-verify
    	NSQ TLS skip verify
  -nsq-topic string
    	NSQ topic
  -pass-work-as-arg
    	pass work as an argument
  -pass-work-as-stdin
    	pass work as stdin
  -payload-file string
    	file to write payload to
  -psql-clear-params string
    	PostgreSQL clear params
  -psql-clear-query string
    	PostgreSQL clear query
  -psql-database string
    	PostgreSQL database
  -psql-fail-params string
    	PostgreSQL fail params
  -psql-fail-query string
    	PostgreSQL fail query
  -psql-host string
    	PostgreSQL host
  -psql-password string
    	PostgreSQL password
  -psql-port string
    	PostgreSQL port (default "5432")
  -psql-query-key
    	PostgreSQL query returns key as first column and value as second column
  -psql-retrieve-field string
    	PostgreSQL retrieve field. If not set, all fields will be returned as a JSON object
  -psql-retrieve-params string
    	PostgreSQL retrieve params
  -psql-retrieve-query string
    	PostgreSQL retrieve query
  -psql-ssl-mode string
    	PostgreSQL SSL mode (default "disable")
  -psql-tls-cert string
    	PostgreSQL TLS cert
  -psql-tls-key string
    	PostgreSQL TLS key
  -psql-tls-root-cert string
    	PostgreSQL SSL root cert
  -psql-user string
    	PostgreSQL user
  -pulsar-address string
    	Pulsar address
  -pulsar-auth-cert-file string
    	Pulsar auth cert file
  -pulsar-auth-key-file string
    	Pulsar auth key file
  -pulsar-auth-oauth-params string
    	Pulsar auth oauth params
  -pulsar-auth-token string
    	Pulsar auth token
  -pulsar-auth-token-file string
    	Pulsar auth token file
  -pulsar-subscription string
    	Pulsar subscription name
  -pulsar-tls-allow-insecure-connection
    	Pulsar TLS allow insecure connection
  -pulsar-tls-trust-certs-file string
    	Pulsar TLS trust certs file path
  -pulsar-tls-validate-hostname
    	Pulsar TLS validate hostname
  -pulsar-topic string
    	Pulsar topic
  -pulsar-topics string
    	Pulsar topics, comma separated
  -pulsar-topics-pattern string
    	Pulsar topics pattern
  -rabbitmq-queue string
    	RabbitMQ queue
  -rabbitmq-url string
    	RabbitMQ URL
  -redis-enable-tls
    	Enable TLS
  -redis-host string
    	Redis host
  -redis-key string
    	Redis key
  -redis-password string
    	Redis password
  -redis-port string
    	Redis port (default "6379")
  -redis-steam-consumer-name string
    	Redis consumer name. Default is a random UUID
  -redis-stream-clear-op string
    	Redis clear operation. Valid values are 'ack' and 'del'.
  -redis-stream-consumer-group string
    	Redis consumer group
  -redis-stream-fail-op string
    	Redis fail operation. Valid values are 'ack' and 'del'.
  -redis-stream-value-keys string
    	Redis stream value keys to select. Comma separated, default all.
  -redis-tls-ca-file string
    	Redis TLS CA file
  -redis-tls-cert-file string
    	Redis TLS cert file
  -redis-tls-key-file string
    	Redis TLS key file
  -redis-tls-skip-verify
    	Redis TLS skip verify
  -scylla-clear-params string
    	Scylla clear params
  -scylla-clear-query string
    	Scylla clear query
  -scylla-consistency string
    	Scylla consistency (default "QUORUM")
  -scylla-fail-params string
    	Scylla fail params
  -scylla-fail-query string
    	Scylla fail query
  -scylla-hosts string
    	Scylla hosts
  -scylla-keyspace string
    	Scylla keyspace
  -scylla-local-dc string
    	Scylla local dc
  -scylla-password string
    	Scylla password
  -scylla-retrieve-field string
    	Scylla retrieve field. If not set, all fields will be returned as a JSON object
  -scylla-retrieve-params string
    	Scylla retrieve params
  -scylla-retrieve-query string
    	Scylla retrieve query
  -scylla-user string
    	Scylla user
  -smb-clear-key string
    	SMB clear key, if clear op is mv. default is origional key name.
  -smb-clear-key-template string
    	SMB clear key template, if clear op is mv.
  -smb-clear-op string
    	SMB clear operation. Valid values: mv, rm
  -smb-fail-key string
    	SMB fail key, if fail op is mv. default is original key name.
  -smb-fail-key-template string
    	SMB fail key template, if fail op is mv.
  -smb-fail-op string
    	SMB fail operation. Valid values: mv, rm
  -smb-host string
    	SMB host
  -smb-key string
    	SMB key
  -smb-key-glob string
    	SMB key glob
  -smb-pass string
    	SMB pass
  -smb-port int
    	SMB port (default 445)
  -smb-share string
    	SMB share
  -smb-user string
    	SMB user
```

### Environment Variables

- `AWS_REGION`
- `AWS_SDK_LOAD_CONFIG`
- `LOG_LEVEL`
- `NSQ_LOG_LEVEL`
- `PROCX_PAYLOAD`
- `PROCX_ACTIVEMQ_ADDRESS`
- `PROCX_ACTIVEMQ_ENABLE_TLS`
- `PROCX_ACTIVEMQ_NAME`
- `PROCX_ACTIVEMQ_TLS_CA_FILE`
- `PROCX_ACTIVEMQ_TLS_CERT_FILE`
- `PROCX_ACTIVEMQ_TLS_INSECURE`
- `PROCX_ACTIVEMQ_TLS_KEY_FILE`
- `PROCX_ACTIVEMQ_TYPE`
- `PROCX_AWS_DYNAMO_CLEAR_QUERY`
- `PROCX_AWS_DYNAMO_FAIL_QUERY`
- `PROCX_AWS_DYNAMO_INCLUDE_NEXT_TOKEN`
- `PROCX_AWS_DYNAMO_LIMIT`
- `PROCX_AWS_DYNAMO_NEXT_TOKEN`
- `PROCX_AWS_DYNAMO_RETRIEVE_FIELD`
- `PROCX_AWS_DYNAMO_RETRIEVE_QUERY`
- `PROCX_AWS_DYNAMO_UNMARSHAL_JSON`
- `PROCX_AWS_LOAD_CONFIG`
- `PROCX_AWS_REGION`
- `PROCX_AWS_ROLE_ARN`
- `PROCX_AWS_S3_BUCKET`
- `PROCX_AWS_S3_CLEAR_BUCKET`
- `PROCX_AWS_S3_CLEAR_KEY`
- `PROCX_AWS_S3_CLEAR_KEY_TEMPLATE`
- `PROCX_AWS_S3_CLEAR_OP`
- `PROCX_AWS_S3_FAIL_BUCKET`
- `PROCX_AWS_S3_FAIL_KEY`
- `PROCX_AWS_S3_FAIL_KEY_TEMPLATE`
- `PROCX_AWS_S3_FAIL_OP`
- `PROCX_AWS_S3_KEY`
- `PROCX_AWS_S3_KEY_PREFIX`
- `PROCX_AWS_S3_KEY_REGEX`
- `PROCX_AWS_SQS_QUEUE_URL`
- `PROCX_AWS_SQS_ROLE_ARN`
- `PROCX_CASSANDRA_CLEAR_PARAMS`
- `PROCX_CASSANDRA_CLEAR_QUERY`
- `PROCX_CASSANDRA_CONSISTENCY`
- `PROCX_CASSANDRA_FAIL_PARAMS`
- `PROCX_CASSANDRA_FAIL_QUERY`
- `PROCX_CASSANDRA_HOSTS`
- `PROCX_CASSANDRA_KEYSPACE`
- `PROCX_CASSANDRA_PASSWORD`
- `PROCX_CASSANDRA_RETRIEVE_FIELD`
- `PROCX_CASSANDRA_RETRIEVE_PARAMS`
- `PROCX_CASSANDRA_RETRIEVE_QUERY`
- `PROCX_CASSANDRA_USER`
- `PROCX_CENTAURI_CHANNEL`
- `PROCX_CENTAURI_KEY`
- `PROCX_CENTAURI_KEY_BASE64`
- `PROCX_CENTAURI_PEER_URL`
- `PROCX_COCKROACH_CLEAR_PARAMS`
- `PROCX_COCKROACH_CLEAR_QUERY`
- `PROCX_COCKROACH_DATABASE`
- `PROCX_COCKROACH_FAIL_PARAMS`
- `PROCX_COCKROACH_FAIL_QUERY`
- `PROCX_COCKROACH_HOST`
- `PROCX_COCKROACH_PASSWORD`
- `PROCX_COCKROACH_PORT`
- `PROCX_COCKROACH_RETRIEVE_FIELD`
- `PROCX_COCKROACH_RETRIEVE_PARAMS`
- `PROCX_COCKROACH_RETRIEVE_QUERY`
- `PROCX_COCKROACH_ROUTING_ID`
- `PROCX_COCKROACH_SSL_MODE`
- `PROCX_COCKROACH_TLS_CERT`
- `PROCX_COCKROACH_TLS_KEY`
- `PROCX_COCKROACH_TLS_ROOT_CERT`
- `PROCX_COCKROACH_USER`
- `PROCX_COUCHBASE_BUCKET_NAME`
- `PROCX_COUCHBASE_CLEAR_BUCKET`
- `PROCX_COUCHBASE_CLEAR_COLLECTION`
- `PROCX_COUCHBASE_CLEAR_DOC`
- `PROCX_COUCHBASE_CLEAR_ID`
- `PROCX_COUCHBASE_CLEAR_OP`
- `PROCX_COUCHBASE_CLEAR_SCOPE`
- `PROCX_COUCHBASE_COLLECTION`
- `PROCX_COUCHBASE_ENABLE_TLS`
- `PROCX_COUCHBASE_FAIL_BUCKET`
- `PROCX_COUCHBASE_FAIL_COLLECTION`
- `PROCX_COUCHBASE_FAIL_DOC`
- `PROCX_COUCHBASE_FAIL_ID`
- `PROCX_COUCHBASE_FAIL_OP`
- `PROCX_COUCHBASE_FAIL_SCOPE`
- `PROCX_COUCHBASE_ID`
- `PROCX_COUCHBASE_PASSWORD`
- `PROCX_COUCHBASE_RETRIEVE_PARAMS`
- `PROCX_COUCHBASE_RETRIEVE_QUERY`
- `PROCX_COUCHBASE_SCOPE`
- `PROCX_COUCHBASE_TLS_CA_FILE`
- `PROCX_COUCHBASE_TLS_CERT_FILE`
- `PROCX_COUCHBASE_TLS_INSECURE`
- `PROCX_COUCHBASE_TLS_KEY_FILE`
- `PROCX_COUCHBASE_USER`
- `PROCX_DAEMON`
- `PROCX_DAEMON_INTERVAL`
- `PROCX_DRIVER`
- `PROCX_ELASTICSEARCH_ADDRESS`
- `PROCX_ELASTICSEARCH_CLEAR_DOC`
- `PROCX_ELASTICSEARCH_CLEAR_INDEX`
- `PROCX_ELASTICSEARCH_CLEAR_OP`
- `PROCX_ELASTICSEARCH_ENABLE_TLS`
- `PROCX_ELASTICSEARCH_FAIL_DOC`
- `PROCX_ELASTICSEARCH_FAIL_INDEX`
- `PROCX_ELASTICSEARCH_FAIL_OP`
- `PROCX_ELASTICSEARCH_PASSWORD`
- `PROCX_ELASTICSEARCH_RETRIEVE_INDEX`
- `PROCX_ELASTICSEARCH_RETRIEVE_QUERY`
- `PROCX_ELASTICSEARCH_TLS_CA_FILE`
- `PROCX_ELASTICSEARCH_TLS_CERT_FILE`
- `PROCX_ELASTICSEARCH_TLS_KEY_FILE`
- `PROCX_ELASTICSEARCH_TLS_SKIP_VERIFY`
- `PROCX_ELASTICSEARCH_USERNAME`
- `PROCX_FS_CLEAR_FOLDER`
- `PROCX_FS_CLEAR_KEY`
- `PROCX_FS_CLEAR_KEY_TEMPLATE`
- `PROCX_FS_CLEAR_OP`
- `PROCX_FS_FAIL_FOLDER`
- `PROCX_FS_FAIL_KEY`
- `PROCX_FS_FAIL_KEY_TEMPLATE`
- `PROCX_FS_FAIL_OP`
- `PROCX_FS_FOLDER`
- `PROCX_FS_KEY`
- `PROCX_FS_KEY_PREFIX`
- `PROCX_FS_KEY_REGEX`
- `PROCX_GCP_BQ_CLEAR_QUERY`
- `PROCX_GCP_BQ_FAIL_QUERY`
- `PROCX_GCP_BQ_RETRIEVE_FIELD`
- `PROCX_GCP_BQ_RETRIEVE_QUERY`
- `PROCX_GCP_FIRESTORE_CLEAR_COLLECTION`
- `PROCX_GCP_FIRESTORE_CLEAR_OP`
- `PROCX_GCP_FIRESTORE_CLEAR_UPDATE`
- `PROCX_GCP_FIRESTORE_FAIL_COLLECTION`
- `PROCX_GCP_FIRESTORE_FAIL_OP`
- `PROCX_GCP_FIRESTORE_FAIL_UPDATE`
- `PROCX_GCP_FIRESTORE_RETRIEVE_COLLECTION`
- `PROCX_GCP_FIRESTORE_RETRIEVE_DOCUMENT`
- `PROCX_GCP_FIRESTORE_RETRIEVE_DOCUMENT_JSON_KEY`
- `PROCX_GCP_FIRESTORE_RETRIEVE_LIMIT`
- `PROCX_GCP_FIRESTORE_RETRIEVE_QUERY_OP`
- `PROCX_GCP_FIRESTORE_RETRIEVE_QUERY_ORDER`
- `PROCX_GCP_FIRESTORE_RETRIEVE_QUERY_ORDER_BY`
- `PROCX_GCP_FIRESTORE_RETRIEVE_QUERY_PATH`
- `PROCX_GCP_FIRESTORE_RETRIEVE_QUERY_VALUE`
- `PROCX_GCP_GCS_BUCKET`
- `PROCX_GCP_GCS_CLEAR_BUCKET`
- `PROCX_GCP_GCS_CLEAR_KEY`
- `PROCX_GCP_GCS_CLEAR_KEY_TEMPLATE`
- `PROCX_GCP_GCS_CLEAR_OP`
- `PROCX_GCP_GCS_FAIL_BUCKET`
- `PROCX_GCP_GCS_FAIL_KEY`
- `PROCX_GCP_GCS_FAIL_KEY_TEMPLATE`
- `PROCX_GCP_GCS_FAIL_OP`
- `PROCX_GCP_GCS_KEY`
- `PROCX_GCP_GCS_KEY_PREFIX`
- `PROCX_GCP_GCS_KEY_REGEX`
- `PROCX_GCP_PROJECT_ID`
- `PROCX_GCP_SUBSCRIPTION`
- `PROCX_GITHUB_BASE_BRANCH`
- `PROCX_GITHUB_BRANCH`
- `PROCX_GITHUB_CLEAR_OP`
- `PROCX_GITHUB_CLEAR_OP_LOCATION`
- `PROCX_GITHUB_COMMIT_EMAIL`
- `PROCX_GITHUB_COMMIT_MESSAGE`
- `PROCX_GITHUB_COMMIT_NAME`
- `PROCX_GITHUB_FAIL_OP`
- `PROCX_GITHUB_FAIL_OP_LOCATION`
- `PROCX_GITHUB_FILE`
- `PROCX_GITHUB_FILE_PREFIX`
- `PROCX_GITHUB_FILE_REGEX`
- `PROCX_GITHUB_OPEN_PR`
- `PROCX_GITHUB_OWNER`
- `PROCX_GITHUB_PR_BODY`
- `PROCX_GITHUB_PR_TITLE`
- `PROCX_GITHUB_REF`
- `PROCX_GITHUB_REPO`
- `PROCX_GITHUB_TOKEN`
- `PROCX_HOSTENV`
- `PROCX_HTTP_CLEAR_BODY`
- `PROCX_HTTP_CLEAR_BODY_FILE`
- `PROCX_HTTP_CLEAR_CONTENT_TYPE`
- `PROCX_HTTP_CLEAR_HEADERS`
- `PROCX_HTTP_CLEAR_METHOD`
- `PROCX_HTTP_CLEAR_SUCCESSFUL_STATUS_CODES`
- `PROCX_HTTP_CLEAR_URL`
- `PROCX_HTTP_ENABLE_TLS`
- `PROCX_HTTP_FAIL_BODY`
- `PROCX_HTTP_FAIL_BODY_FILE`
- `PROCX_HTTP_FAIL_CONTENT_TYPE`
- `PROCX_HTTP_FAIL_HEADERS`
- `PROCX_HTTP_FAIL_METHOD`
- `PROCX_HTTP_FAIL_SUCCESSFUL_STATUS_CODES`
- `PROCX_HTTP_FAIL_URL`
- `PROCX_HTTP_RETRIEVE_BODY`
- `PROCX_HTTP_RETRIEVE_BODY_FILE`
- `PROCX_HTTP_RETRIEVE_CONTENT_TYPE`
- `PROCX_HTTP_RETRIEVE_HEADERS`
- `PROCX_HTTP_RETRIEVE_KEY_JSON_SELECTOR`
- `PROCX_HTTP_RETRIEVE_METHOD`
- `PROCX_HTTP_RETRIEVE_SUCCESSFUL_STATUS_CODES`
- `PROCX_HTTP_RETRIEVE_URL`
- `PROCX_HTTP_RETRIEVE_WORK_JSON_SELECTOR`
- `PROCX_HTTP_TLS_CA_FILE`
- `PROCX_HTTP_TLS_CERT_FILE`
- `PROCX_HTTP_TLS_KEY_FILE`
- `PROCX_KAFKA_BROKERS`
- `PROCX_KAFKA_ENABLE_SASL`
- `PROCX_KAFKA_ENABLE_TLS`
- `PROCX_KAFKA_GROUP`
- `PROCX_KAFKA_SASL_PASSWORD`
- `PROCX_KAFKA_SASL_TYPE`
- `PROCX_KAFKA_SASL_USERNAME`
- `PROCX_KAFKA_TLS_CA_FILE`
- `PROCX_KAFKA_TLS_CERT_FILE`
- `PROCX_KAFKA_TLS_INSECURE`
- `PROCX_KAFKA_TLS_KEY_FILE`
- `PROCX_KAFKA_TOPIC`
- `PROCX_KEEP_PAYLOAD_FILE`
- `PROCX_MONGO_AUTH_SOURCE`
- `PROCX_MONGO_CLEAR_QUERY`
- `PROCX_MONGO_COLLECTION`
- `PROCX_MONGO_DATABASE`
- `PROCX_MONGO_ENABLE_TLS`
- `PROCX_MONGO_FAIL_QUERY`
- `PROCX_MONGO_HOST`
- `PROCX_MONGO_LIMIT`
- `PROCX_MONGO_PASSWORD`
- `PROCX_MONGO_PORT`
- `PROCX_MONGO_RETRIEVE_QUERY`
- `PROCX_MONGO_TLS_CA_FILE`
- `PROCX_MONGO_TLS_CERT_FILE`
- `PROCX_MONGO_TLS_INSECURE`
- `PROCX_MONGO_TLS_KEY_FILE`
- `PROCX_MONGO_USER`
- `PROCX_MSSQL_CLEAR_PARAMS`
- `PROCX_MSSQL_CLEAR_QUERY`
- `PROCX_MSSQL_DATABASE`
- `PROCX_MSSQL_FAIL_PARAMS`
- `PROCX_MSSQL_FAIL_QUERY`
- `PROCX_MSSQL_HOST`
- `PROCX_MSSQL_PASSWORD`
- `PROCX_MSSQL_PORT`
- `PROCX_MSSQL_RETRIEVE_FIELD`
- `PROCX_MSSQL_RETRIEVE_PARAMS`
- `PROCX_MSSQL_RETRIEVE_QUERY`
- `PROCX_MSSQL_USER`
- `PROCX_MYSQL_CLEAR_PARAMS`
- `PROCX_MYSQL_CLEAR_QUERY`
- `PROCX_MYSQL_DATABASE`
- `PROCX_MYSQL_FAIL_PARAMS`
- `PROCX_MYSQL_FAIL_QUERY`
- `PROCX_MYSQL_HOST`
- `PROCX_MYSQL_PASSWORD`
- `PROCX_MYSQL_PORT`
- `PROCX_MYSQL_RETRIEVE_FIELD`
- `PROCX_MYSQL_RETRIEVE_PARAMS`
- `PROCX_MYSQL_RETRIEVE_QUERY`
- `PROCX_MYSQL_USER`
- `PROCX_NATS_CLEAR_RESPONSE`
- `PROCX_NATS_CREDS_FILE`
- `PROCX_NATS_ENABLE_TLS`
- `PROCX_NATS_FAIL_RESPONSE`
- `PROCX_NATS_JWT_FILE`
- `PROCX_NATS_NKEY_FILE`
- `PROCX_NATS_PASSWORD`
- `PROCX_NATS_QUEUE_GROUP`
- `PROCX_NATS_SUBJECT`
- `PROCX_NATS_TLS_CA_FILE`
- `PROCX_NATS_TLS_CERT_FILE`
- `PROCX_NATS_TLS_INSECURE`
- `PROCX_NATS_TLS_KEY_FILE`
- `PROCX_NATS_TOKEN`
- `PROCX_NATS_URL`
- `PROCX_NATS_USERNAME`
- `PROCX_NFS_CLEAR_FOLDER`
- `PROCX_NFS_CLEAR_KEY`
- `PROCX_NFS_CLEAR_KEY_TEMPLATE`
- `PROCX_NFS_CLEAR_OP`
- `PROCX_NFS_FAIL_FOLDER`
- `PROCX_NFS_FAIL_KEY`
- `PROCX_NFS_FAIL_KEY_TEMPLATE`
- `PROCX_NFS_FAIL_OP`
- `PROCX_NFS_FOLDER`
- `PROCX_NFS_HOST`
- `PROCX_NFS_KEY`
- `PROCX_NFS_KEY_PREFIX`
- `PROCX_NFS_KEY_REGEX`
- `PROCX_NFS_TARGET`
- `PROCX_NSQ_CHANNEL`
- `PROCX_NSQ_ENABLE_TLS`
- `PROCX_NSQ_NSQD_ADDRESS`
- `PROCX_NSQ_NSQLOOKUPD_ADDRESS`
- `PROCX_NSQ_TLS_CA_FILE`
- `PROCX_NSQ_TLS_CERT_FILE`
- `PROCX_NSQ_TLS_INSECURE`
- `PROCX_NSQ_TLS_KEY_FILE`
- `PROCX_NSQ_TOPIC`
- `PROCX_PASS_WORK_AS_ARG`
- `PROCX_PASS_WORK_AS_STDIN`
- `PROCX_PAYLOAD_FILE`
- `PROCX_PSQL_CLEAR_PARAMS`
- `PROCX_PSQL_CLEAR_QUERY`
- `PROCX_PSQL_DATABASE`
- `PROCX_PSQL_FAIL_PARAMS`
- `PROCX_PSQL_FAIL_QUERY`
- `PROCX_PSQL_HOST`
- `PROCX_PSQL_PASSWORD`
- `PROCX_PSQL_PORT`
- `PROCX_PSQL_RETRIEVE_FIELD`
- `PROCX_PSQL_RETRIEVE_PARAMS`
- `PROCX_PSQL_RETRIEVE_QUERY`
- `PROCX_PSQL_SSL_MODE`
- `PROCX_PSQL_TLS_CERT`
- `PROCX_PSQL_TLS_KEY`
- `PROCX_PSQL_TLS_ROOT_CERT`
- `PROCX_PSQL_USER`
- `PROCX_PULSAR_ADDRESS`
- `PROCX_PULSAR_AUTH_CERT_FILE`
- `PROCX_PULSAR_AUTH_KEY_FILE`
- `PROCX_PULSAR_AUTH_OAUTH_PARAMS`
- `PROCX_PULSAR_AUTH_TOKEN`
- `PROCX_PULSAR_AUTH_TOKEN_FILE`
- `PROCX_PULSAR_SUBSCRIPTION`
- `PROCX_PULSAR_TLS_ALLOW_INSECURE_CONNECTION`
- `PROCX_PULSAR_TLS_TRUST_CERTS_FILE`
- `PROCX_PULSAR_TLS_VALIDATE_HOSTNAME`
- `PROCX_PULSAR_TOPIC`
- `PROCX_PULSAR_TOPICS`
- `PROCX_PULSAR_TOPICS_PATTERN`
- `PROCX_RABBITMQ_QUEUE`
- `PROCX_RABBITMQ_URL`
- `PROCX_REDIS_ENABLE_TLS`
- `PROCX_REDIS_HOST`
- `PROCX_REDIS_KEY`
- `PROCX_REDIS_PASSWORD`
- `PROCX_REDIS_PORT`
- `PROCX_REDIS_STREAM_CLEAR_OP`
- `PROCX_REDIS_STREAM_CONSUMER_GROUP`
- `PROCX_REDIS_STREAM_CONSUMER_NAME`
- `PROCX_REDIS_STREAM_FAIL_OP`
- `PROCX_REDIS_STREAM_VALUE_KEYS`
- `PROCX_REDIS_TLS_CA_FILE`
- `PROCX_REDIS_TLS_CERT_FILE`
- `PROCX_REDIS_TLS_INSECURE`
- `PROCX_REDIS_TLS_KEY_FILE`
- `PROCX_SCYLLA_CLEAR_PARAMS`
- `PROCX_SCYLLA_CLEAR_QUERY`
- `PROCX_SCYLLA_CONSISTENCY`
- `PROCX_SCYLLA_FAIL_PARAMS`
- `PROCX_SCYLLA_FAIL_QUERY`
- `PROCX_SCYLLA_HOSTS`
- `PROCX_SCYLLA_KEYSPACE`
- `PROCX_SCYLLA_LOCAL_DC`
- `PROCX_SCYLLA_PASSWORD`
- `PROCX_SCYLLA_RETRIEVE_FIELD`
- `PROCX_SCYLLA_RETRIEVE_PARAMS`
- `PROCX_SCYLLA_RETRIEVE_QUERY`
- `PROCX_SCYLLA_USER`
- `PROCX_SMB_CLEAR_KEY`
- `PROCX_SMB_CLEAR_KEY_TEMPLATE`
- `PROCX_SMB_CLEAR_OP`
- `PROCX_SMB_FAIL_KEY`
- `PROCX_SMB_FAIL_KEY_TEMPLATE`
- `PROCX_SMB_FAIL_OP`
- `PROCX_SMB_HOST`
- `PROCX_SMB_KEY`
- `PROCX_SMB_KEY_GLOB`
- `PROCX_SMB_PASS`
- `PROCX_SMB_PORT`
- `PROCX_SMB_SHARE`
- `PROCX_SMB_USER`

## Driver Examples

### ActiveMQ

The ActiveMQ driver will connect to the specified STOMP address and retrieve the next message from the specified queue or topic. If the message is successfully retrieved, the message will be deleted from the queue, otherwise it will be `nacked`. TLS is optional and shown below, if not used the flags are not required.

```bash
procx \
    -driver activemq \
    -activemq-address localhost:61613 \
    -activemq-type queue \
    -activemq-name my-queue \
    -activemq-enable-tls \
    -activemq-tls-ca-file /path/to/ca.pem \
    -activemq-tls-cert-file /path/to/cert.pem \
    -activemq-tls-key-file /path/to/key.pem \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### AWS DynamoDB

The AWS DynamoDB driver will execute the provided PartiQL query and return the matched results. An optional JSON path can be passed in the `-aws-dynamo-retrieve-field` flag, if this is provided it will be used to extract the value from the returned data before passing to the process, otherwise the full array of Dynamo JSON documents is passed. Similar to other SQL-based drivers, you can use `gjson` syntax to extract values from the data which can be used in subsequent clear and fail handling queries. The `-aws-dynamo-next-token` can be provided to continue querying from a previous result, and `-aws-dynamo-include-next-token` can be set to pass the `_nextToken` in the response payload. By default, the Dynamo JSON objects are unmarshalled, however this can be disabled with `-aws-dynamo-unmarshal-json=false`.

```bash
procx \
    -driver aws-dynamo \
    -aws-dynamo-retrieve-query "SELECT id,job,status FROM my-table WHERE status = 'pending'" \
    -aws-dynamo-limit 1 \
    -aws-dynamo-clear-query "UPDATE my-table SET status='complete' WHERE id = '{{0.id}}'" \
    -aws-dynamo-fail-query "UPDATE my-table SET status='failed' WHERE id = '{{0.id}}'" \
    -aws-region us-east-1 \
    -aws-role-arn arn:aws:iam::123456789012:role/my-role \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### AWS S3

The S3 driver will retrieve the first object which matches the specified input within the bucket. If `-aws-s3-key` is provided, this exact key will be retrieved. If `-aws-s3-key-prefix` is provided, this will be used as a prefix to select the first matching object. Finally, if `-aws-s3-key-regex` is provided, this will be used as a regular expression to select the first matching object. 

Upon completion of the work, the object can either be deleted from the bucket, or moved to a different bucket, with `-aws-s3-clear-op=[mv|rm]` and `-aws-s3-clear-bucket` flags. Similarly for failed executions, the object can either be deleted from the bucket, or moved to a different bucket, with `-aws-s3-fail-op=[mv|rm]` and `-aws-s3-fail-bucket` flags. 

By default, if the object is moved the key will be the same as the source key, this can be overridden with the `-aws-s3-clear-key` and `-aws-s3-fail-key` flags. You can also provide a `-aws-s3-clear-key-template` and/or `-aws-s3-fail-key-template` flag to use a templated key - this is useful if you have used the prefix or regex selector and want to retain the object key but rename the file.

```bash
procx \
    -driver aws-s3 \
    -payload-file my-payload.json \
    -aws-s3-bucket my-bucket \
    -aws-s3-key-regex 'jobs-.*?[a-z]' \
    -aws-s3-clear-op=mv \
    -aws-s3-clear-bucket my-bucket-completed \
    -aws-s3-clear-key-template 'success/{{key}}' \
    -aws-s3-fail-op=mv \
    -aws-s3-fail-bucket my-bucket-completed \
    -aws-s3-fail-key-template 'fail/{{key}}' \
    bash -c 'echo the payload is: $(cat my-payload.json)'
```

### AWS SQS

The SQS driver will retrieve the next message from the specified queue, and pass it to the process. Upon successful completion of the process, it will delete the message from the queue.

For cross-account access, you must provide the ARN of the role that has access to the queue, and the identity running procx must be able to assume the target identity.

If running on a developer workstation, you will most likely want to pass your `~/.aws/config` identity. To do so, pass the `-aws-load-config` flag.

```bash
procx \
    -aws-sqs-queue-url https://sqs.us-east-1.amazonaws.com/123456789012/my-queue \
    -aws-role-arn arn:aws:iam::123456789012:role/my-role \
    -aws-region us-east-1 \
    -driver aws-sqs \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Cassandra

The Cassandra driver will retrieve the specified rows from the specified keyspace table, and pass it to the process. Upon successful completion of the process, it will execute the specified query to update / remove the work from the table.

```bash
procx \
    -cassandra-keyspace mykeyspace \
    -cassandra-consistency QUORUM \
    -cassandra-clear-query "DELETE FROM mykeyspace.mytable WHERE id = ?" \
    -cassandra-clear-params "{{0.id}}" \
    -cassandra-hosts "localhost:9042,another:9042" \
    -cassandra-fail-query "UPDATE mykeyspace.mytable SET status = 'failed' WHERE id = ?" \
    -cassandra-fail-params "{{0.id}}" \
    -cassandra-retrieve-field work \
    -cassandra-retrieve-query "SELECT id, work FROM mykeyspace.mytable LIMIT 1" \
    -driver cassandra \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Centauri

The `centauri` driver integrates with a [Centauri](https://centauri.sh) network to retrieve the next message from the specified channel, and pass it to the process. Upon successful completion of the process, it will delete the message from the network. You can provide your private key either as `-centauri-key`, or as `-centauri-key-base64`.

```bash
procx \
    -centauri-channel my-channel \
    -centauri-key "$(</path/to/private.key)" \
    -centauri-peer-url https://api.test-peer1.centauri.sh \
    -driver centauri \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Cockroach

The CockroachDB driver will retrieve the data from the specified table, and pass it to the process.

```bash
procx \
    -cockroach-host localhost \
    -cockroach-port 26257 \
    -cockroach-database mydb \
    -cockroach-user myuser \
    -cockroach-password mypassword \
    -cockroach-retrieve-query "SELECT id, work from mytable where queue = $1 and status = $2 LIMIT 1" \
    -cockroach-retrieve-params "myqueue,pending" \
    -cockroach-clear-query "UPDATE mytable SET status = $1 where queue = $2 and id = $3" \
    -cockroach-clear-params "cleared,myqueue,{{0.id}}" \
    -cockroach-fail-query "UPDATE mytable SET failure_count = failure_count + 1 where queue = $1 and id = $2" \
    -cockroach-fail-params "myqueue,{{0.id}}" \
    -driver cockroach \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Couchbase

The Couchbase driver will retrieve the specified document from the bucket and pass it to the process. If `-couchbase-id` is specified, this document will be retrieved. Alternatively, you can use `-couchbase-retrieve-query` and `-couchbase-retrieve-params` to retrieve a document with a N1QL query. Upon completion of the work, the document can be moved (`mv`) to a different bucket/collection, deleted (`rm`) from the bucket, updated in place with a new document (`set`), or merged with a new document (`merge`). If moving the document, you can also provide a document which will be merged with the document before it is moved. You can use `gjson` selectors and mustache syntax to template the new document before it is merged.

```bash
procx \
    -couchbase-bucket my-bucket \
    -couchbase-collection my-collection \
    -couchbase-retrieve-query "SELECT id, jobName, work, status from my-collection where status = $1 LIMIT 1" \
    -couchbase-retrieve-params "pending" \
    -couchbase-clear-op=mv \
    -couchbase-clear-doc '{"status": "cleared"}' \
    -couchbase-clear-bucket my-bucket-cleared \
    -couchbase-clear-collection my-collection-cleared \
    -couchbase-fail-op=merge \
    -couchbase-fail-doc '{"status": "failed"}' \
    -driver couchbase \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Elasticsearch

The Elasticsearch driver will retrieve the next message from the specified index, and pass it to the process. Upon successful completion of the process, it will execute the specified query to either update, move, or delete the document from the index.

```bash
procx \
    -elasticsearch-address https://localhost:9200 \
    -elasticsearch-username elastic \
    -elasticsearch-password elastic \
    -elasticsearch-tls-skip-verify \
    -elasticsearch-retrieve-query '{
            "size": 1,
            "query": {
                "term": {
                    "hello": "world"
                }
            }
        }' \
    -elasticsearch-retrieve-index my-index \
    -elasticsearch-clear-op merge-put \
    -elasticsearch-index my-index \
    -elasticsearch-clear-doc '{"status": "completed"}' \
    -elasticsearch-fail-op move \
    -elasticsearch-fail-index my-index-failed \
    -driver elasticsearch \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### FS

The `fs` driver will traverse the specified locally mounted directory, and retrieve the first file which matches the specified key. Similar to the AWS S3 and NFS drivers, the FS driver supports `-fs-key`, `-fs-key-prefix`, and `-fs-key-regex` selection flags. 

Upon completion, the file can either be moved to a different folder, or it can be deleted, with the `-fs-clear-op` and `-fs-fail-op` flags (`mv` or `rm`). You can specify the target folder with the `-fs-clear-folder` and `-fs-fail-folder` flags, and the `-fs-clear-key` and `-fs-fail-key` flags let you rename the file on move. You can also use the `-fs-clear-key-template` and `-fs-fail-key-template` flags to specify a template for the key, which will be replaced with the key.

```bash
procx \
    -fs-folder /path/to/folder \
    -fs-key-prefix "my-prefix" \
    -fs-clear-op mv \
    -fs-clear-folder /path/to/cleared \
    -fs-clear-key-template "cleared_{{key}}" \
    -fs-fail-op mv \
    -fs-fail-folder /path/to/failed \
    -fs-fail-key-template "failed_{{key}}" \
    -driver fs \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### GCP BQ

The `gcp-bq` driver will retrieve the next message from the specified BigQuery table, and pass it to the process. Upon successful completion of the process, it will execute the specified query to update / remove the work from the table. By default, the row data will be returned as a JSON object, unless `-gcp-bq-retrieve-field` is specified, in which case only the specifed field will be returned.

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
procx \
    -gcp-project-id my-project \
    -gcp-bq-dataset my-dataset \
    -gcp-bq-table my-table \
    -gcp-bq-retrieve-query "SELECT id, work FROM mydatatest.mytable LIMIT 1" \
    -gcp-bq-clear-query "DELETE FROM my-table WHERE id = '{{0.id}}'" \
    -gcp-bq-fail-query "UPDATE my-table SET status = 'failed' WHERE id = '{{0.id}}'" \
    -driver gcp-bq \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### GCP GCS

The GCS driver will retrieve the first object which matches the specified input within the bucket. If `-gcp-gcs-key` is provided, this exact key will be retrieved. If `-gcp-gcs-key-prefix` is provided, this will be used as a prefix to select the first matching object. Finally, if `-gcp-gcs-key-regex` is provided, this will be used as a regular expression to select the first matching object.

Upon completion of the work, the object can either be deleted from the bucket, or moved to a different bucket, with `-gcp-gcs-clear-op=[mv|rm]` and `-gcp-gcs-clear-bucket` flags. Similarly for failed executions, the object can either be deleted from the bucket, or moved to a different bucket, with `-gcp-gcs-fail-op=[mv|rm]` and `-gcp-gcs-fail-bucket` flags. 

By default, if the object is moved the key will be the same as the source key, this can be overridden with the `-gcp-gcs-clear-key` and `-gcp-gcs-fail-key` flags. You can also provide a `-gcp-gcs-clear-key-template` and/or `-gcp-gcs-fail-key-template` flag to use a templated key - this is useful if you have used the prefix or regex selector and want to retain the object key but rename the file.

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
procx \
    -driver gcp-gcs \
    -gcp-project-id my-project \
    -payload-file my-payload.json \
    -gcp-gcs-bucket my-bucket \
    -gcp-gcs-key-regex 'jobs-.*?[a-z]' \
    -gcp-gcs-clear-op=rm \
    -gcp-gcs-fail-op=mv \
    -gcp-gcs-fail-bucket my-bucket-completed \
    -gcp-gcs-fail-key-template 'fail/{{key}}' \
    bash -c 'echo the payload is: $(cat my-payload.json)'
```

### GCP Firestore

The GCP Firestore driver will retrieve the documents in the collection `-gcp-firestore-retrieve-collection` which matches the specified input query. If `-gcp-firestore-retrieve-document` is specified, this exact document ID will be retrieved. If `-gcp-firestore-retrieve-query-path` is specified, it will be used with `-gcp-firestore-retrieve-query-op` and `-gcp-firestore-retrieve-query-value` to construct a select query. To order the documents before selecting the first response, `-gcp-firestore-retrieve-query-order-by` and `-gcp-firestore-retrieve-query-order` can be used. The documents will be returned to the process as a `[]map[string]interface{}` JSON array. If neither document nor query is specified, the first document in the collection will be retrieved. If `-gcp-firestore-retrieve-document-json-key` is provided, it will be used to select a single field in the JSON repsonse to pass to the process. Upon completion, the document can either be moved to a new collection, updated, or deleted. If updating, the new fields can be provided as a JSON string which will be merged with the object.

```bash
procx \
    -driver gcp-firestore \
    -gcp-project-id my-project \
    -gcp-firestore-retrieve-collection my-collection \
    -gcp-firestore-retrieve-query-path 'status' \
    -gcp-firestore-retrieve-query-op '==' \
    -gcp-firestore-retrieve-query-value 'pending' \
    -gcp-firestore-retrieve-query-order-by 'created_at' \
    -gcp-firestore-retrieve-query-order 'desc' \
    -gcp-firestore-retrieve-document-json-key 'work' \
    -gcp-firestore-clear-op=mv \
    -gcp-firestore-clear-update '{"status": "completed"}' \
    -gcp-firestore-clear-collection my-collection-completed \
    -gcp-firestore-fail-op=mv \
    -gcp-firestore-fail-update '{"status": "failed"}' \
    -gcp-firestore-fail-collection my-collection-failed \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```    

### GCP Pub/Sub

The GCP Pub/Sub driver will retrieve the next message from the specified subscription, and pass it to the process. Upon successful completion of the process, it will acknowledge the message.

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
procx \
    -gcp-project-id my-project \
    -gcp-pubsub-subscription my-subscription \
    -driver gcp-pubsub \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### GitHub

The GitHub driver will retrieve the specified object (`-github-file`, `-github-file-prefix`, or `-github-file-regex`) from a GitHub repository (`-github-repo`) and pass it to the process. This mirrors other file-system drivers and supports both regex and prefix selection in addition to explicit key selection. Upon completion or failure, the object can either be deleted (`rm`) or moved (`mv`) in the repository. This can be done either on the same branch (`-github-ref` or `-github-base-branch`), or a new branch (`-github-branch`). If on a new branch, a pull request can be opened with `-github-open-pr`. If opening a new PR without a branch specified, a new branch name will be generated.

```bash
procx \
    -driver github \
    -github-repo my-repo \
    -github-owner my-owner \
    -github-branch my-new-branch \
    -github-base-branch main \
    -github-file-regex 'pending/jobs-.*?[a-z]' \
    -github-token my-token \
    -github-open-pr \
    -github-commit-name my-commit-name \
    -github-commit-email my-commit-email \
    -github-commit-message my-commit-message \
    -github-pr-title my-pr-title \
    -github-pr-body my-pr-body \
    -github-clear-op rm \
    -github-fail-op mv \
    -github-fail-location 'failed/{{key}}' \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### HTTP

The HTTP driver will connect to any HTTP(s) endpoint, retrieve the content from the endpoint, and return it to the process. If the upstream returns a JSON payload, the `-http-retrieve-work-json-selector` flag can be used to select a specific value from the JSON response, and similarly the `-http-retrieve-key-json-selector` flag can be used to select a value from the JSON response which can then be used to replace any instances of `{{key}}` in either the clear or failure URL or payload. If using internal PKI, mTLS, or disabling TLS validation, pass the `-http-enable-tls` flag and the corresponding TLS flags.

```bash
procx \
    -http-retrieve-url https://example.com/jobs \
    -http-retrieve-work-json-selector '0.work' \
    -http-retrieve-key-json-selector '0.id' \
    -http-retrieve-headers 'ExampleToken:foobar,ExampleHeader:barfoo' \
    -http-clear-url 'https://example.com/jobs/{{key}}' \
    -http-clear-method DELETE \
    -http-clear-headers 'ExampleToken:foobar,ExampleHeader:barfoo' \
    -http-fail-url 'https://example.com/jobs/{{key}}' \
    -http-fail-method POST \
    -http-fail-headers 'ExampleToken:foobar,ExampleHeader:barfoo' \
    -http-fail-body '{"id": "{{key}}", "status": "failed"}' \
    -driver http \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Kafka

The Kafka driver will retrieve the next message from the specified topic, and pass it to the process. If a group is passed, this will be used to ensure that the message is only retrieved from the topic once across distributed workers and subsequent executions. Similar to the Pub/Sub drivers, if there are no messages in the topic when the process starts, it will wait for the first message. TLS and SASL authentication are optional.

```bash
procx \
    -kafka-brokers localhost:9092 \
    -kafka-topic my-topic \
    -kafka-group my-group \
    -kafka-enable-tls \
    -kafka-tls-ca-file /path/to/ca.pem \
    -kafka-tls-cert-file /path/to/cert.pem \
    -kafka-tls-key-file /path/to/key.pem \
    -kafka-enable-sasl \
    -kafka-sasl-type plain \
    -kafka-sasl-username my-username \
    -kafka-sasl-password my-password \
    -driver kafka \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### MongoDB

The MongoDB driver will retrieve the documents from the specified collection, and pass it to the process. Upon successful completion of the process, it will run the specified mongo command.

```bash
procx \
    -mongo-collection my-collection \
    -mongo-database my-database \
    -mongo-host localhost \
    -mongo-port 27017 \
    -mongo-user my-user \
    -mongo-password my-password \
    -mongo-retrieve-query '{"status": "pending"}' \
    -mongo-limit 1 \
    -mongo-clear-query '{"delete": "my-collection", "deletes": [{"q": {"_id": {"$oid": "{{0._id}}"}}, "limit": 1}]}' \
    -mongo-fail-query '{"update":"my-collection","updates":[{"q":{"_id":{"$oid":"{{0._id}}"}},"u":{"$set": {"failed":true}}}]}' \
    -driver mongodb \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### MSSQL

The MSSQL driver will retrieve the messages from the specified Microsoft SQL Server, and pass it to the process.

```bash
procx \
    -mssql-host localhost \
    -mssql-port 1433 \
    -mssql-database mydb \
    -mssql-user sa \
    -mssql-password 'mypassword!' \
    -mssql-retrieve-query "SET ROWCOUNT 1 SELECT id, work from mytable where queue = ? and status = ?" \
    -mssql-retrieve-params "myqueue,pending" \
    -mssql-clear-query "UPDATE mytable SET status = ? where queue = ? and id = ?" \
    -mssql-clear-params "cleared,myqueue,{{0.id}}" \
    -mssql-fail-query "UPDATE mytable SET failure_count = failure_count + 1 where queue = ? and id = ?" \
    -mssql-fail-params "myqueue,{{0.id}}" \
    -driver mssql \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### MySQL

The MySQL driver will retrieve the messages from the specified database, and pass it to the process.

```bash
procx \
    -mysql-host localhost \
    -mysql-port 3306 \
    -mysql-database mydb \
    -mysql-user myuser \
    -mysql-password mypassword \
    -mysql-retrieve-query "SELECT id, work from mytable where queue = ? and status = ? LIMIT 1" \
    -mysql-retrieve-params "myqueue,pending" \
    -mysql-clear-query "UPDATE mytable SET status = ? where queue = ? and id = ?" \
    -mysql-clear-params "cleared,myqueue,{{0.id}}" \
    -mysql-fail-query "UPDATE mytable SET failure_count = failure_count + 1 where queue = ? and id = ?" \
    -mysql-fail-params "myqueue,{{0.id}}" \
    -driver mysql \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### NATS

The NATS driver will retrieve the next message from the specified subject, and pass it to the process. If the received message type is a request, the message will be replied to with the result of the process (`0` or `1` for success or failure, respectively). You can optionally provide a `-nats-clear-response` and `-nats-fail-response` flag to send a custom response on completion. You can optionally provide `-nats-queue-group` to namespace the queue group worker subscriptions.

```bash
procx \
    -nats-subject my-subject \
    -nats-url localhost:4222 \
    -nats-username my-user \
    -nats-password my-password \
    -nats-clear-response "OK" \
    -nats-fail-response "FAIL" \
    -nats-queue-group my-group \
    -nats-enable-tls \
    -nats-tls-ca-file /path/to/ca.pem \
    -nats-tls-cert-file /path/to/cert.pem \
    -nats-tls-key-file /path/to/key.pem \
    -driver nats \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### NFS

The `nfs` driver will mount the specified NFS directory, and retrieve the first file which matches the specified key. Similar to the AWS S3 driver, the NFS driver supports `-nfs-key`, `-nfs-key-prefix`, and `-nfs-key-regex` selection flags. 

Upon completion, the file can either be moved to a different folder in the NFS, or it can be deleted, with the `-nfs-clear-op` and `-nfs-fail-op` flags (`mv` or `rm`). You can specify the target folder with the `-nfs-clear-folder` and `-nfs-fail-folder` flags, and the `-nfs-clear-key` and `-nfs-fail-key` flags let you rename the file on move. You can also use the `-nfs-clear-key-template` and `-nfs-fail-key-template` flags to specify a template for the key, which will be replaced with the key.

```bash
procx \
    -nfs-host nfs.example.com \
    -nfs-target /path/to/nfs \
    -nfs-key-prefix "my-prefix" \
    -nfs-clear-op mv \
    -nfs-clear-folder cleared \
    -nfs-clear-key-template "cleared_{{key}}" \
    -nfs-fail-op mv \
    -nfs-fail-folder failed \
    -nfs-fail-key-template "failed_{{key}}" \
    -driver nfs \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### NSQ

The NSQ driver will connect to the specified `nsqlookupd` or `nsqd` endpoint, retrieve the next message from the specified topic, and pass it to the process.

```bash
procx \
    -nsq-nsqlookupd-address localhost:4161 \
    -nsq-topic my-topic \
    -nsq-channel my-channel \
    -driver nsq \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### PostgreSQL

The PostgreSQL driver will retrieve the messages from the specified queue, and pass it to the process.

```bash
procx \
    -psql-host localhost \
    -psql-port 5432 \
    -psql-database mydb \
    -psql-user myuser \
    -psql-password mypassword \
    -psql-retrieve-query "SELECT id, work from mytable where queue = $1 and status = $2" \
    -psql-retrieve-params "myqueue,pending" \
    -psql-clear-query "UPDATE mytable SET status = $1 where queue = $2 and id = $3" \
    -psql-clear-params "cleared,myqueue,{{0.id}}" \
    -psql-fail-query "UPDATE mytable SET failure_count = failure_count + 1 where queue = $1 and id = $2" \
    -psql-fail-params "myqueue,{{0.id}}" \
    -driver postgres \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Pulsar

The Pulsar driver will connect to the specified comma-separated Pulsar endpoint(s) and retrieve the next message from the specified topic, and pass it to the process. An `ack` will be sent on success, and a `nack` will be sent on failure. Clients can subscribe to either a specific topic, a set of topics (comma separated), or a regex pattern. Token and TLS auth methods are supported.

```bash
procx \
    -pulsar-address localhost:6650,localhost:6651 \
    -pulsar-topic my-topic \
    -pulsar-subscription my-subscription \
    -pulsar-auth-cert-file /path/to/cert.pem \
    -pulsar-auth-key-file /path/to/key.pem \
    -pulsar-tls-trust-certs-file /path/to/trusted.pem \
    -driver pulsar \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### RabbitMQ

The RabbitMQ driver will connect to the specified queue AMQP endpoint and retrieve the next message from the specified queue.

```bash
procx \
    -rabbitmq-url amqp://guest:guest@localhost:5672 \
    -rabbitmq-queue my-queue \
    -driver rabbitmq \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Redis List

The Redis List driver will connect to the specified Redis server and retrieve the next message from the specified list.

```bash
procx \
    -redis-host localhost \
    -redis-port 6379 \
    -redis-key my-list \
    -driver redis-list \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Redis Pub/Sub

The Redis Pub/Sub driver will connect to the specified Redis server and retrieve the next message from the specified subscription.

```bash
procx \
    -redis-host localhost \
    -redis-port 6379 \
    -redis-key my-subscription \
    -driver redis-pubsub \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Redis Stream

The Redis Stream driver will connect to the specified Redis server and retrieve the next message from the specified stream. The message contents will be returned as a JSON object. You can optionally select a subset of keys with the `-redis-stream-value-keys` flag, providing a comma-separated list of keys to include in the payload. By default, the newest message will be read from the stream. You can control worker namespacing with the `-redis-stream-consumer-group` flag. You can either `ack` or `del` the message with the `-redis-stream-clear-op` and `-redis-stream-fail-op` flags.

```bash
procx \
    -redis-host localhost \
    -redis-port 6379 \
    -redis-key my-stream \
    -redis-stream-consumer-group my-group \
    -redis-stream-value-keys "key1,key2" \
    -redis-stream-clear-op del \
    -redis-stream-fail-op ack \
    -driver redis-stream \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Scylla

The Scylla driver will retrieve the specified rows from the specified keyspace table, and pass it to the process. Upon successful completion of the process, it will execute the specified query to update / remove the work from the table.

```bash
procx \
    -scylla-keyspace mykeyspace \
    -scylla-consistency QUORUM \
    -scylla-clear-query "DELETE FROM mykeyspace.mytable WHERE id = ?" \
    -scylla-clear-params "{{0.id}}" \
    -scylla-hosts "localhost:9042,another:9042" \
    -scylla-fail-query "UPDATE mykeyspace.mytable SET status = 'failed' WHERE id = ?" \
    -scylla-fail-params "{{0.id}}" \
    -scylla-retrieve-field work \
    -scylla-retrieve-query "SELECT id, work FROM mykeyspace.mytable LIMIT 1" \
    -driver scylla \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### SMB

The SMB driver will connect to the specified SMB endpoint and retrieve the specified file `-smb-key`, however the first file which matches a glob `-smb-key-glob` can also be used. Similar to the NFS driver, on clear / faiure, the file can be moved (`mv`) or deleted (`rm`) with the `-smb-clear-op` and `-smb-fail-op` flags. The `-smb-clear-key` and `-smb-fail-key` flags can be used to specify the new path, and `-smb-clear-key-template` and `-smb-fail-key-template` can be used to specify a template for the new path, where `{{key}}` is replaced with the original key base name.

```bash
procx \
    -smb-host localhost \
    -smb-port 445 \
    -smb-share myshare \
    -smb-key-glob "jobs/job-*" \
    -smb-clear-op mv \
    -smb-clear-key-template "cleared/cleared_{{key}}" \
    -smb-fail-op mv \
    -smb-fail-key-template "failed/failed_{{key}}" \
    -driver smb \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Local

The local driver is a simple wrapper around the process to execute, primarily for local testing. It does not communicate with any queue, and expects the job payload to be manually defined by the operator as a `PROCX_PAYLOAD` environment variable.

This can also be used to read in a file, or to shim in a local pipe for testing.

```bash
PROCX_PAYLOAD="$(</path/to/payload.txt)" \
procx \
    -driver local \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

## Orchestration

procx is solely focused on the worker-side consumption and clearing of work, and intentionally has no scope to the scheduling or management of work.

This allows you to plug in any scheduling or management system you want, and have procx consume the work from that system.

If you are running in Kubernetes, the [`procx-operator`](https://github.com/robertlestak/procx-operator) is a simple operator that will manage ProcX workloads on top of Kubernetes and KEDA.

## Deployment

You will need to install procx in the container which will be used to run your job. You can either compile procx from source, or use the latest precompiled binaries available.

### Example Dockerfile

```dockerfile
FROM node:17

RUN apt-get update && apt-get install -y \
    curl

RUN curl -LO https://github.com/robertlestak/procx/releases/latest/download/procx_linux && \
    chmod +x procx_linux && \
    mv procx_linux /usr/local/bin/procx

RUN echo "console.log('the payload is:', process.env.PROCX_PAYLOAD)" > app.js

CMD ["node", "app.js"]
ENTRYPOINT ["/usr/local/bin/procx"]
```

```bash
docker build -t procx .
```

```bash
docker run --rm -it \
    -v ~/.aws:/root/.aws \
    -e PROCX_AWS_REGION=us-east-1 \
    -e PROCX_AWS_SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789012/my-queue \
    -e PROCX_AWS_ROLE_ARN=arn:aws:iam::123456789012:role/my-role \
    -e PROCX_DRIVER=aws-sqs \
    -e AWS_SDK_LOAD_CONFIG=1 \
    procx
```
