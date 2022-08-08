# procx - simple job queue worker

procx is a small process manager that can wrap around any existing application / script / process, and integrate with a job queue system to enable autoscaling job executions with no native code integration.

procx is a single compiled binary that can be packaged in your existing job code container. procx is configured with environment variables or command line flags, and is started with the path to the process to execute.

procx will retrieve the next job from the queue, and pass it to the process. Upon success (exit code 0), procx will mark the job as complete. Upon failure (exit code != 0), procx will mark the job as failed to be requeued.

By default, the subprocess spawned by procx will not have access to the host environment variables. This can be changed by setting the `-hostenv` flag.

By default, procx will connect to the data source, consume a single message, and then exit when the spawned process exits. If the `-daemon` flag is set, procx will connect to the data source and consume messages until the process is killed, or until a job fails.

## Payload

By default, procx will export the payload as an environment variable `PROCX_PAYLOAD`. If `-pass-work-as-arg` is set, the job payload string will be appended to the process arguments, and if the `-payload-file` flag is set, the payload will be written to the specified file path. procx will clean up the file at the end of the job, unless you pass `-keep-payload-file`. Finally, if `-pass-work-as-stdin` is set, the job payload will be piped to stdin of the process.

## Drivers

Currently, the following drivers are supported:

- [AWS DynamoDB](#aws-dynamodb) (`aws-dynamo`)
- [AWS S3](#aws-s3) (`aws-s3`)
- [AWS SQS](#aws-sqs) (`aws-sqs`)
- [Cassandra](#cassandra) (`cassandra`)
- [Centauri](#centauri) (`centauri`)
- [Elasticsearch](#elasticsearch) (`elasticsearch`)
- [GCP Big Query](#gcp-bq) (`gcp-bq`)
- [GCP Cloud Storage](#gcp-gcs) (`gcp-gcs`)
- [GCP Pub/Sub](#gcp-pubsub) (`gcp-pubsub`)
- [Kafka](#kafka) (`kafka`)
- [PostgreSQL](#postgresql) (`postgres`)
- [MongoDB](#mongodb) (`mongodb`)
- [MySQL](#mysql) (`mysql`)
- [NATS](#nats) (`nats`)
- [NFS](#nfs) (`nfs`)
- [RabbitMQ](#rabbitmq) (`rabbitmq`)
- [Redis List](#redis-list) (`redis-list`)
- [Redis Pub/Sub](#redis-pubsub) (`redis-pubsub`)
- [Redis Stream](#redis-stream) (`redis-stream`)
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
  -aws-dynamo-clear-query string
    	AWS DynamoDB clear query
  -aws-dynamo-data-path string
    	AWS DynamoDB data JSON path
  -aws-dynamo-fail-query string
    	AWS DynamoDB fail query
  -aws-dynamo-key-path string
    	AWS DynamoDB query key JSON path
  -aws-dynamo-retrieve-query string
    	AWS DynamoDB retrieve query
  -aws-dynamo-table string
    	AWS DynamoDB table name
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
  -cassandra-query-key
    	Cassandra query returns key as first column and value as second column
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
  -daemon
    	run as daemon
  -driver string
    	driver to use. (aws-dynamo, aws-s3, aws-sqs, cassandra, centauri, elasticsearch, gcp-bq, gcp-gcs, gcp-pubsub, kafka, local, mongodb, mysql, nats, nfs, postgres, rabbitmq, redis-list, redis-pubsub, redis-stream)
  -elasticsearch-address string
    	Elasticsearch address
  -elasticsearch-clear-index string
    	Elasticsearch clear index
  -elasticsearch-clear-op string
    	Elasticsearch clear op. Valid values are: delete, put, merge-put, move
  -elasticsearch-clear-query string
    	Elasticsearch clear query
  -elasticsearch-enable-tls
    	Elasticsearch enable TLS
  -elasticsearch-fail-index string
    	Elasticsearch fail index
  -elasticsearch-fail-op string
    	Elasticsearch fail op. Valid values are: delete, put, merge-put, move
  -elasticsearch-fail-query string
    	Elasticsearch fail query
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
  -gcp-bq-clear-query string
    	GCP BQ clear query
  -gcp-bq-fail-query string
    	GCP BQ fail query
  -gcp-bq-query-key
    	GCP BQ query returns key as first column and value as second column
  -gcp-bq-retrieve-query string
    	GCP BQ retrieve query
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
  -hostenv
    	use host environment
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
  -mongo-clear-query string
    	MongoDB clear query
  -mongo-collection string
    	MongoDB collection
  -mongo-database string
    	MongoDB database
  -mongo-fail-query string
    	MongoDB fail query
  -mongo-host string
    	MongoDB host
  -mongo-password string
    	MongoDB password
  -mongo-port string
    	MongoDB port (default "27017")
  -mongo-retrieve-query string
    	MongoDB retrieve query
  -mongo-user string
    	MongoDB user
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
  -mysql-query-key
    	MySQL query returns key as first column and value as second column
  -mysql-retrieve-params string
    	MySQL retrieve params
  -mysql-retrieve-query string
    	MySQL retrieve query
  -mysql-user string
    	MySQL user
  -nats-clear-response string
    	Nats clear response
  -nats-creds-file string
    	Nats creds file
  -nats-enable-tls
    	Nats enable TLS
  -nats-fail-response string
    	Nats fail response
  -nats-jwt-file string
    	Nats JWT file
  -nats-nkey-file string
    	Nats NKey file
  -nats-password string
    	Nats password
  -nats-queue-group string
    	Nats queue group
  -nats-subject string
    	Nats subject
  -nats-tls-ca-file string
    	Nats TLS CA file
  -nats-tls-cert-file string
    	Nats TLS cert file
  -nats-tls-insecure
    	Nats TLS insecure
  -nats-tls-key-file string
    	Nats TLS key file
  -nats-token string
    	Nats token
  -nats-url string
    	Nats URL
  -nats-username string
    	Nats username
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
  -nfs-mount-path string
    	NFS mount path
  -nfs-target string
    	NFS target
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
  -psql-retrieve-params string
    	PostgreSQL retrieve params
  -psql-retrieve-query string
    	PostgreSQL retrieve query
  -psql-ssl-mode string
    	PostgreSQL SSL mode (default "disable")
  -psql-user string
    	PostgreSQL user
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
```

### Environment Variables

- `PROCX_AWS_REGION`
- `PROCX_AWS_ROLE_ARN`
- `PROCX_AWS_DYNAMO_DATA_PATH`
- `PROCX_AWS_DYNAMO_TABLE`
- `PROCX_AWS_DYNAMO_KEY_PATH`
- `PROCX_AWS_DYNAMO_RETRIEVE_QUERY`
- `PROCX_AWS_DYNAMO_CLEAR_QUERY`
- `PROCX_AWS_DYNAMO_FAIL_QUERY`
- `PROCX_AWS_S3_BUCKET`
- `PROCX_AWS_S3_KEY`
- `PROCX_AWS_S3_KEY_PREFIX`
- `PROCX_AWS_S3_KEY_REGEX`
- `PROCX_AWS_S3_CLEAR_BUCKET`
- `PROCX_AWS_S3_CLEAR_KEY`
- `PROCX_AWS_S3_CLEAR_KEY_TEMPLATE`
- `PROCX_AWS_S3_CLEAR_OP`
- `PROCX_AWS_S3_FAIL_BUCKET`
- `PROCX_AWS_S3_FAIL_KEY`
- `PROCX_AWS_S3_FAIL_KEY_TEMPLATE`
- `PROCX_AWS_S3_FAIL_OP`
- `PROCX_AWS_SQS_QUEUE_URL`
- `PROCX_CASSANDRA_CLEAR_PARAMS`
- `PROCX_CASSANDRA_CLEAR_QUERY`
- `PROCX_CASSANDRA_CONSISTENCY`
- `PROCX_CASSANDRA_FAIL_PARAMS`
- `PROCX_CASSANDRA_FAIL_QUERY`
- `PROCX_CASSANDRA_HOSTS`
- `PROCX_CASSANDRA_KEYSPACE`
- `PROCX_CASSANDRA_PASSWORD`
- `PROCX_CASSANDRA_QUERY_KEY`
- `PROCX_CASSANDRA_RETRIEVE_PARAMS`
- `PROCX_CASSANDRA_RETRIEVE_QUERY`
- `PROCX_CASSANDRA_USER`
- `PROCX_CENTAURI_CHANNEL`
- `PROCX_CENTAURI_KEY`
- `PROCX_CENTAURI_KEY_BASE64`
- `PROCX_CENTAURI_PEER_URL`
- `PROCX_ELASTICSEARCH_ADDRESS`
- `PROCX_ELASTICSEARCH_USERNAME`
- `PROCX_ELASTICSEARCH_PASSWORD`
- `PROCX_ELASTICSEARCH_TLS_SKIP_VERIFY`
- `PROCX_ELASTICSEARCH_RETRIEVE_QUERY`
- `PROCX_ELASTICSEARCH_RETRIEVE_INDEX`
- `PROCX_ELASTICSEARCH_CLEAR_QUERY`
- `PROCX_ELASTICSEARCH_CLEAR_INDEX`
- `PROCX_ELASTICSEARCH_CLEAR_OP`
- `PROCX_ELASTICSEARCH_FAIL_QUERY`
- `PROCX_ELASTICSEARCH_FAIL_INDEX`
- `PROCX_ELASTICSEARCH_FAIL_OP`
- `PROCX_ELASTICSEARCH_ENABLE_TLS`
- `PROCX_ELASTICSEARCH_TLS_CA_FILE`
- `PROCX_ELASTICSEARCH_TLS_CERT_FILE`
- `PROCX_ELASTICSEARCH_TLS_KEY_FILE`
- `PROCX_GCP_PROJECT_ID`
- `PROCX_GCP_BQ_CLEAR_QUERY`
- `PROCX_GCP_BQ_FAIL_QUERY`
- `PROCX_GCP_BQ_QUERY_KEY`
- `PROCX_GCP_BQ_RETRIEVE_QUERY`
- `PROCX_GCP_GCS_BUCKET`
- `PROCX_GCP_GCS_KEY`
- `PROCX_GCP_GCS_KEY_PREFIX`
- `PROCX_GCP_GCS_KEY_REGEX`
- `PROCX_GCP_GCS_CLEAR_BUCKET`
- `PROCX_GCP_GCS_CLEAR_KEY`
- `PROCX_GCP_GCS_CLEAR_KEY_TEMPLATE`
- `PROCX_GCP_GCS_CLEAR_OP`
- `PROCX_GCP_GCS_FAIL_BUCKET`
- `PROCX_GCP_GCS_FAIL_KEY`
- `PROCX_GCP_GCS_FAIL_KEY_TEMPLATE`
- `PROCX_GCP_GCS_FAIL_OP`
- `PROCX_GCP_PUBSUB_SUBSCRIPTION`
- `PROCX_KAFKA_BROKERS`
- `PROCX_KAFKA_GROUP`
- `PROCX_KAFKA_TOPIC`
- `PROCX_KAFKA_TLS_CA_FILE`
- `PROCX_KAFKA_TLS_CERT_FILE`
- `PROCX_KAFKA_TLS_KEY_FILE`
- `PROCX_KAFKA_ENABLE_TLS`
- `PROCX_KAFKA_ENABLE_SASL`
- `PROCX_KAFKA_SASL_USERNAME`
- `PROCX_KAFKA_SASL_PASSWORD`
- `PROCX_KAFKA_SASL_TYPE`
- `PROCX_KAFKA_TLS_INSECURE`
- `PROCX_DRIVER`
- `PROCX_HOSTENV`
- `PROCX_KEEP_PAYLOAD_FILE`
- `PROCX_MONGO_CLEAR_QUERY`
- `PROCX_MONGO_COLLECTION`
- `PROCX_MONGO_DATABASE`
- `PROCX_MONGO_FAIL_QUERY`
- `PROCX_MONGO_HOST`
- `PROCX_MONGO_PASSWORD`
- `PROCX_MONGO_PORT`
- `PROCX_MONGO_RETRIEVE_QUERY`
- `PROCX_MONGO_USER`
- `PROCX_MYSQL_CLEAR_PARAMS`
- `PROCX_MYSQL_CLEAR_QUERY`
- `PROCX_MYSQL_DATABASE`
- `PROCX_MYSQL_FAIL_PARAMS`
- `PROCX_MYSQL_FAIL_QUERY`
- `PROCX_MYSQL_HOST`
- `PROCX_MYSQL_PASSWORD`
- `PROCX_MYSQL_PORT`
- `PROCX_MYSQL_QUERY_KEY`
- `PROCX_MYSQL_RETRIEVE_PARAMS`
- `PROCX_MYSQL_RETRIEVE_QUERY`
- `PROCX_MYSQL_USER`
- `PROCX_NATS_URL`
- `PROCX_NATS_QUEUE_GROUP`
- `PROCX_NATS_SUBJECT`
- `PROCX_NATS_CREDS_FILE`
- `PROCX_NATS_JWT_FILE`
- `PROCX_NATS_TLS_CA_FILE`
- `PROCX_NATS_TLS_CERT_FILE`
- `PROCX_NATS_TLS_KEY_FILE`
- `PROCX_NATS_ENABLE_TLS`
- `PROCX_NATS_TLS_INSECURE`
- `PROCX_NATS_NKEY_FILE`
- `PROCX_NATS_USERNAME`
- `PROCX_NATS_PASSWORD`
- `PROCX_NATS_CLEAR_RESPONSE`
- `PROCX_NATS_FAIL_RESPONSE`
- `PROCX_NFS_FAIL_OP`
- `PROCX_NFS_FOLDER`
- `PROCX_NFS_HOST`
- `PROCX_NFS_KEY`
- `PROCX_NFS_KEY_PREFIX`
- `PROCX_NFS_KEY_REGEX`
- `PROCX_NFS_MOUNT_PATH`
- `PROCX_NFS_TARGET`
- `PROCX_NFS_CLEAR_OP`
- `PROCX_NFS_FAIL_OP`
- `PROCX_NFS_CLEAR_FOLDER`
- `PROCX_NFS_FAIL_FOLDER`
- `PROCX_NFS_CLEAR_KEY`
- `PROCX_NFS_FAIL_KEY`
- `PROCX_NFS_CLEAR_KEY_TEMPLATE`
- `PROCX_NFS_FAIL_KEY_TEMPLATE`
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
- `PROCX_PSQL_QUERY_KEY`
- `PROCX_PSQL_RETRIEVE_PARAMS`
- `PROCX_PSQL_RETRIEVE_QUERY`
- `PROCX_PSQL_SSL_MODE`
- `PROCX_PSQL_USER`
- `PROCX_RABBITMQ_URL`
- `PROCX_RABBITMQ_QUEUE`
- `PROCX_REDIS_HOST`
- `PROCX_REDIS_PORT`
- `PROCX_REDIS_PASSWORD`
- `PROCX_REDIS_KEY`
- `PROCX_REDIS_ENABLE_TLS`
- `PROCX_REDIS_TLS_CA_FILE`
- `PROCX_REDIS_TLS_CERT_FILE`
- `PROCX_REDIS_TLS_KEY_FILE`
- `PROCX_REDIS_TLS_INSECURE`
- `PROCX_REDIS_STREAM_CLEAR_OP`
- `PROCX_REDIS_STREAM_FAIL_OP`
- `PROCX_REDIS_STREAM_VALUE_KEYS`
- `PROCX_REDIS_STREAM_CONSUMER_GROUP`
- `PROCX_REDIS_STREAM_CONSUMER_NAME`
- `PROCX_DAEMON`

## Driver Examples

### AWS DynamoDB

The AWS DynamoDB driver will execute the provided PartiQL query and return the first result. An optional JSON path can be passed in the `-aws-dynamo-key-path` flag, if this is provided it will be used to extract the value from the returned data, and this will replace `{{key}}` in subsequent clear and fail handling queries. Additionally, an optional `-aws-dynamo-data-path` flag can be passed in, if this is provided it will be used to extract the data from the returned JSON.

```bash
procx \
    -driver aws-dynamo \
    -aws-dynamo-table my-table \
    -aws-dynamo-key-path 'id.S' \
    -aws-dynamo-retrieve-query "SELECT id,job,status FROM my-table WHERE status = 'pending'" \
    -aws-dynamo-data-path 'job.S' \
    -aws-dynamo-clear-query "UPDATE my-table SET status='complete' WHERE id = '{{key}}'" \
    -aws-dynamo-fail-query "UPDATE my-table SET status='failed' WHERE id = '{{key}}'" \
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

The Cassandra driver will retrieve the next message from the specified keyspace table, and pass it to the process. Upon successful completion of the process, it will execute the specified query to update / remove the work from the table.

```bash
procx \
    -cassandra-keyspace mykeyspace \
    -cassandra-consistency QUORUM \
    -cassandra-clear-query "DELETE FROM mykeyspace.mytable WHERE id = ?" \
    -cassandra-clear-params "{{key}}" \
    -cassandra-hosts "localhost:9042,another:9042" \
    -cassandra-fail-query "UPDATE mykeyspace.mytable SET status = 'failed' WHERE id = ?" \
    -cassandra-fail-params "{{key}}" \
    -cassandra-query-key \
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

### Elasticsearch

The Elasticsearch driver will retrieve the next message from the specified index, and pass it to the process. Upon successful completion of the process, it will execute the specified query to either update, move, or delete the document from the index.

```bash
procx \
    -elasticsearch-address https://localhost:9200 \
    -elasticsearch-username elastic \
    -elasticsearch-password elastic \
    -elasticsearch-tls-skip-verify \
    -elasticsearch-retrieve-query '{"status": "pending"}' \
    -elasticsearch-retrieve-index my-index \
    -elasticsearch-clear-op merge-put \
    -elasticsearch-clear-index my-index \
    -elasticsearch-clear-query '{"status": "completed"}' \
    -elasticsearch-fail-op move \
    -elasticsearch-fail-index my-index-failed \
    -driver elasticsearch \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### GCP BQ

The `gcp-bq` driver will retrieve the next message from the specified BigQuery table, and pass it to the process. Upon successful completion of the process, it will execute the specified query to update / remove the work from the table. By default, it is assumed that a single value (single column and row) is returned by `-gcp-bq-retrieve-query`. You can optionally set `-gcp-bq-query-key` and return a unique key column and work value as the second column. This then enables you to template your `-gcp-bq-clear-query` and `-gcp-bq-fail-query` queries with the `{{key}}` placeholder.

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
procx \
    -gcp-bq-project my-project \
    -gcp-bq-dataset my-dataset \
    -gcp-bq-table my-table \
    -gcp-bq-retrieve-query "SELECT id, work FROM mydatatest.mytable LIMIT 1" \
    -gcp-bq-query-key \
    -gcp-bq-clear-query "DELETE FROM my-table WHERE id = '{{key}}'" \
    -gcp-bq-fail-query "UPDATE my-table SET status = 'failed' WHERE id = '{{key}}'" \
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
    -payload-file my-payload.json \
    -gcp-gcs-bucket my-bucket \
    -gcp-gcs-key-regex 'jobs-.*?[a-z]' \
    -gcp-gcs-clear-op=rm \
    -gcp-gcs-fail-op=mv \
    -gcp-gcs-fail-bucket my-bucket-completed \
    -gcp-gcs-fail-key-template 'fail/{{key}}' \
    bash -c 'echo the payload is: $(cat my-payload.json)'
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
    -kafka-sasl-mechanism plain \
    -kafka-sasl-username my-username \
    -kafka-sasl-password my-password \
    -driver kafka \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### MongoDB

The MongoDB driver will retrieve the next message from the specified collection, and pass it to the process. Upon successful completion of the process, it will run the specified mongo command. The Mongo ObjectID `_id` will be passed in for the placeholder `{{key}}`.

```bash
procx \
    -mongo-collection my-collection \
    -mongo-database my-database \
    -mongo-host localhost \
    -mongo-port 27017 \
    -mongo-user my-user \
    -mongo-password my-password \
    -mongo-retrieve-query '{"status": "pending"}' \
    -mongo-clear-query '{"delete": "my-collection", "deletes": [{"q": {"_id": {"$oid": "{{key}}"}}, "limit": 1}]}' \
    -mongo-fail-query '{"update":"my-collection","updates":[{"q":{"_id":{"$oid":"{{key}}"}},"u":{"$set": {"failed":true}}}]}' \
    -driver mongodb \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### MySQL

The MySQL driver will retrieve the next message from the specified queue, and pass it to the process. By default, the query used to retrieve the message (`-mysql-retrieve-query`) will assume to return a single column, however if you pass `-mysql-query-key` it will assume to return a two-column result, with the first column being the key and the second column being the value. This then allows you to provide a placeholder `{{key}}` param for clearing / failure queries, and this will be replaced with the respective key.

```bash
procx \
    -mysql-host localhost \
    -mysql-port 3306 \
    -mysql-database mydb \
    -mysql-user myuser \
    -mysql-password mypassword \
    -mysql-retrieve-query "SELECT id, work from mytable where queue = ? and status = ?" \
    -mysql-query-key \
    -mysql-retrieve-params "myqueue,pending" \
    -mysql-clear-query "UPDATE mytable SET status = ? where queue = ? and id = ?" \
    -mysql-clear-params "cleared,myqueue,{{key}}" \
    -mysql-fail-query "UPDATE mytable SET failure_count = failure_count + 1 where queue = ? and id = ?" \
    -mysql-fail-params "myqueue,{{key}}" \
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

The NFS driver will mount the specified NFS directory, and retrieve the first file which matches the specified key. Similar to the AWS S3 driver, the NFS driver supports `-nfs-key`, `-nfs-key-prefix`, and `-nfs-key-regex` selection flags. 

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

### PostgreSQL

The PostgreSQL driver will retrieve the next message from the specified queue, and pass it to the process. By default, the query used to retrieve the message (`-psql-retrieve-query`) will assume to return a single column, however if you pass `-psql-query-key` it will assume to return a two-column result, with the first column being the key and the second column being the value. This then allows you to provide a placeholder `{{key}}` param for clearing / failure queries, and this will be replaced with the respective key.

```bash
procx \
    -psql-host localhost \
    -psql-port 5432 \
    -psql-database mydb \
    -psql-user myuser \
    -psql-password mypassword \
    -psql-retrieve-query "SELECT id, work from mytable where queue = $1 and status = $2" \
    -psql-query-key \
    -psql-retrieve-params "myqueue,pending" \
    -psql-clear-query "UPDATE mytable SET status = $1 where queue = $2 and id = $3" \
    -psql-clear-params "cleared,myqueue,{{key}}" \
    -psql-fail-query "UPDATE mytable SET failure_count = failure_count + 1 where queue = $1 and id = $2" \
    -psql-fail-params "myqueue,{{key}}" \
    -driver postgres \
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
