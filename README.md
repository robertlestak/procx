# procx - simple job queue worker

procx is a small process manager that can wrap around any existing application / script / process, and integrate with a job queue system to enable autoscaling job executions with no native code integration.

procx is a single compiled binary that can be packaged in your existing job code container. procx is configured with environment variables or command line flags, and is started with the path to the process to execute.

procx will retrieve the next job from the queue, and pass it to the process. Upon success (exit code 0), procx will mark the job as complete. Upon failure (exit code != 0), procx will mark the job as failed to be requeued.

procx will make the job payload available in the `PROCX_PAYLOAD` environment variable. If `-pass-work-as-arg` is set, the job payload string will be appended to the process arguments.

By default, the subprocess spawned by procx will not have access to the host environment variables. This can be changed by setting the `-hostenv` flag.

By default, procx will connect to the data source, consume a single message, and then exit when the spawned process exits. If the `-daemon` flag is set, procx will connect to the data source and consume messages until the process is killed, or until a job fails.

## Drivers

Currently, the following drivers are supported:

- AWS SQS (`aws-sqs`)
- Cassandra (`cassandra`)
- Centauri (`centauri`)
- GCP Pub/Sub (`gcp-pubsub`)
- PostgreSQL (`postgres`)
- MongoDB (`mongodb`)
- MySQL (`mysql`)
- RabbitMQ (`rabbitmq`)
- Redis List (`redis-list`)
- Redis Pub/Sub (`redis-pubsub`)
- Local (`local`)

Plans to add more drivers in the future, and PRs are welcome.

See [Driver Examples](#driver-examples) for more information.

## Install

```bash
curl -SsL https://raw.githubusercontent.com/robertlestak/procx/main/scripts/install.sh | bash -e
```

### A note on permissions

Depending on the path of `INSTALL_DIR` and the permissions of the user running the installation script, you may get a Permission Denied error if you are trying to move the binary into a location which your current user does not have access to. This is most often the case when running the script as a non-root user yet trying to install into `/usr/local/bin`. To fix this, you can either:

Create a `$HOME/bin` directory in your current user home directory. This will be the default installation directory. Be sure to add this to your `$PATH` environment variable.
Use `sudo` to run the installation script, to install into `/usr/local/bin` (`curl -SsL https://raw.githubusercontent.com/robertlestak/procx/main/scripts/install.sh | sudo bash -e`).

### Build From Source

```bash
mkdir -p bin
go build -o bin/procx cmd/procx/*.go
```

#### Building for a Specific Driver

By default, the `procx` binary is compiled for all drivers. This is to enable a truly build-once-run-anywhere experience. However some users may want a smaller binary for embedded workloads. To enable this, you can edit `pkg/drivers/drivers.go` and remove the drivers you do not want to include, and recompile.

While building for a specific driver may seem contrary to the ethos of procx, the decoupling between the job queue and work still enables a write-once-run-anywhere experience, and simply requires DevOps to rebuild the image with your new drivers if you are shifting upstream data sources.

## Usage

```bash
procx [flags] <process path>
  -aws-load-config
    	load AWS config from ~/.aws/config
  -aws-region string
    	AWS region
  -aws-sqs-queue-url string
    	AWS SQS queue URL
  -aws-sqs-role-arn string
    	AWS SQS role ARN
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
  -centauri-peer-url string
    	Centauri peer URL
  -daemon
    	run as daemon
  -driver string
    	driver to use. (aws-sqs, cassandra, centauri, gcp-pubsub, postgres, mongodb, mysql, rabbitmq, redis-list, redis-pubsub, local)
  -gcp-project-id string
    	GCP project ID
  -gcp-pubsub-subscription string
    	GCP Pub/Sub subscription name
  -hostenv
    	use host environment
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
  -pass-work-as-arg
    	pass work as an argument
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
  -redis-host string
    	Redis host
  -redis-key string
    	Redis key
  -redis-password string
    	Redis password
  -redis-port string
    	Redis port (default "6379")
```

### Environment Variables

- `PROCX_AWS_REGION`
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
- `PROCX_CASSANDRA_QUERY_KEY`
- `PROCX_CASSANDRA_RETRIEVE_PARAMS`
- `PROCX_CASSANDRA_RETRIEVE_QUERY`
- `PROCX_CASSANDRA_USER`
- `PROCX_CENTAURI_CHANNEL`
- `PROCX_CENTAURI_KEY`
- `PROCX_CENTAURI_PEER_URL`
- `PROCX_GCP_PROJECT_ID`
- `PROCX_GCP_PUBSUB_SUBSCRIPTION`
- `PROCX_DRIVER`
- `PROCX_HOSTENV`
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
- `PROCX_PASS_WORK_AS_ARG`
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
- `PROCX_DAEMON`

## Driver Examples

### AWS SQS

The SQS driver will retrieve the next message from the specified queue, and pass it to the process. Upon successful completion of the process, it will delete the message from the queue.

For cross-account access, you must provide the ARN of the role that has access to the queue, and the identity running procx must be able to assume the target identity.

If running on a developer workstation, you will most likely want to pass your `~/.aws/config` identity. To do so, pass the `-aws-load-config` flag.

```bash
procx \
    -aws-sqs-queue-url https://sqs.us-east-1.amazonaws.com/123456789012/my-queue \
    -aws-sqs-role-arn arn:aws:iam::123456789012:role/my-role \
    -aws-region us-east-1 \
    -driver aws-sqs \
    bash -c 'echo the payload is: $PROCX_PAYLOAD'
```

### Cassandra

The Cassandra driver will retrieve the next message from the specified keyspace table, and pass it to the process. Upon successful completion of the process, it will execute the specified query to update / remove the work from the table.

```bash
procx \
    -cassandra-keyspace mykeyspace \
    -cassandra-table mytable \
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

The `centauri` driver integrates with a [Centauri](https://centauri.sh) network to retrieve the next message from the specified channel, and pass it to the process. Upon successful completion of the process, it will delete the message from the network.

```bash
procx \
    -centauri-channel my-channel \
    -centauri-key "$(</path/to/private.key)" \
    -centauri-peer-url https://api.test-peer1.centauri.sh \
    -driver centauri \
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
    -e PROCX_AWS_SQS_ROLE_ARN=arn:aws:iam::123456789012:role/my-role \
    -e PROCX_DRIVER=aws-sqs \
    -e AWS_SDK_LOAD_CONFIG=1 \
    procx
```
