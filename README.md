# qjob - simple job queue worker

qjob is a small process manager that can wrap around any existing application / script / process, and integrate with a job queue system to enable autoscaling job executions with no native code integration.

qjob is a single compiled binary that can be packaged in your existing job code container. qjob is configured with environment variables or command line flags, and is started with the path to the process to execute.

qjob will retrieve the next job from the queue, and pass it to the process. Upon success (exit code 0), qjob will mark the job as complete. Upon failure (exit code != 0), qjob will mark the job as failed to be requeued.

qjob will make the job payload available in the `QJOB_PAYLOAD` environment variable. If `-pass-work-as-arg` is set, the job payload string will be appended to the process arguments.

By default, the subprocess spawned by qjob will not have access to the host environment variables. This can be changed by setting the `-hostenv` flag.

By default, qjob will connect to the data source, consume a single message, and then exit when the spawned process exits. If the `-daemon` flag is set, qjob will connect to the data source and consume messages until the process is killed, or until a job fails.

## Install

```bash
curl -SsL https://raw.githubusercontent.com/robertlestak/qjob/main/scripts/install.sh | bash -e
```

### A note on permissions

Depending on the path of `INSTALL_DIR` and the permissions of the user running the installation script, you may get a Permission Denied error if you are trying to move the binary into a location which your current user does not have access to. This is most often the case when running the script as a non-root user yet trying to install into `/usr/local/bin`. To fix this, you can either:

Create a `$HOME/bin` directory in your current user home directory. This will be the default installation directory. Be sure to add this to your `$PATH` environment variable.
Use `sudo` to run the installation script, to install into `/usr/local/bin` (`curl -SsL https://raw.githubusercontent.com/robertlestak/qjob/main/scripts/install.sh | sudo bash -e`).

## Usage

```bash
qjob [flags] <process path>
  -aws-load-config
    	load AWS config from ~/.aws/config
  -aws-region string
    	AWS region
  -aws-sqs-queue-url string
    	AWS SQS queue URL
  -aws-sqs-role-arn string
    	AWS SQS role ARN
  -daemon
    	run as daemon
  -driver string
    	driver to use. (aws-sqs, gcp-pubsub, postgres, mysql, rabbitmq, redis-list, redis-pubsub, local)
  -gcp-project-id string
    	GCP project ID
  -gcp-pubsub-subscription string
    	GCP Pub/Sub subscription name
  -hostenv
    	use host environment
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

- `QJOB_AWS_REGION`
- `QJOB_AWS_SQS_QUEUE_URL`
- `QJOB_AWS_SQS_ROLE_ARN`
- `QJOB_GCP_PROJECT_ID`
- `QJOB_GCP_PUBSUB_SUBSCRIPTION`
- `QJOB_DRIVER`
- `QJOB_HOSTENV`
- `QJOB_MYSQL_CLEAR_PARAMS`
- `QJOB_MYSQL_CLEAR_QUERY`
- `QJOB_MYSQL_DATABASE`
- `QJOB_MYSQL_FAIL_PARAMS`
- `QJOB_MYSQL_FAIL_QUERY`
- `QJOB_MYSQL_HOST`
- `QJOB_MYSQL_PASSWORD`
- `QJOB_MYSQL_PORT`
- `QJOB_MYSQL_QUERY_KEY`
- `QJOB_MYSQL_RETRIEVE_PARAMS`
- `QJOB_MYSQL_RETRIEVE_QUERY`
- `QJOB_MYSQL_USER`
- `QJOB_PASS_WORK_AS_ARG`
- `QJOB_PSQL_CLEAR_PARAMS`
- `QJOB_PSQL_CLEAR_QUERY`
- `QJOB_PSQL_DATABASE`
- `QJOB_PSQL_FAIL_PARAMS`
- `QJOB_PSQL_FAIL_QUERY`
- `QJOB_PSQL_HOST`
- `QJOB_PSQL_PASSWORD`
- `QJOB_PSQL_PORT`
- `QJOB_PSQL_QUERY_KEY`
- `QJOB_PSQL_RETRIEVE_PARAMS`
- `QJOB_PSQL_RETRIEVE_QUERY`
- `QJOB_PSQL_SSL_MODE`
- `QJOB_PSQL_USER`
- `QJOB_RABBITMQ_URL`
- `QJOB_RABBITMQ_QUEUE`
- `QJOB_REDIS_HOST`
- `QJOB_REDIS_PORT`
- `QJOB_REDIS_PASSWORD`
- `QJOB_REDIS_KEY`
- `QJOB_DAEMON`

## Drivers

Currently, the following drivers are supported:

- AWS SQS (`aws-sqs`)
- GCP Pub/Sub (`gcp-pubsub`)
- PostgreSQL (`postgres`)
- MySQL (`mysql`)
- RabbitMQ (`rabbitmq`)
- Redis List (`redis-list`)
- Redis Pub/Sub (`redis-pubsub`)
- Local (`local`)

Plans to add more drivers in the future, and PRs are welcome.

### AWS SQS

The SQS driver will retrieve the next message from the specified queue, and pass it to the process. Upon successful completion of the process, it will delete the message from the queue.

For cross-account access, you must provide the ARN of the role that has access to the queue, and the identity running qjob must be able to assume the target identity.

If running on a developer workstation, you will most likely want to pass your `~/.aws/config` identity. To do so, pass the `-aws-load-config` flag.

```bash
qjob \
    -aws-sqs-queue-url https://sqs.us-east-1.amazonaws.com/123456789012/my-queue \
    -aws-sqs-role-arn arn:aws:iam::123456789012:role/my-role \
    -aws-region us-east-1 \
    -driver aws-sqs \
    bash -c 'echo the payload is: $QJOB_PAYLOAD'
```

### GCP Pub/Sub

The GCP Pub/Sub driver will retrieve the next message from the specified subscription, and pass it to the process. Upon successful completion of the process, it will acknowledge the message.

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
qjob \
    -gcp-project-id my-project \
    -gcp-pubsub-subscription my-subscription \
    -driver gcp-pubsub \
    bash -c 'echo the payload is: $QJOB_PAYLOAD'
```

### MySQL

The MySQL driver will retrieve the next message from the specified queue, and pass it to the process. By default, the query used to retrieve the message (`-mysql-retrieve-query`) will assume to return a single column, however if you pass `-mysql-query-key` it will assume to return a two-column result, with the first column being the key and the second column being the value. This then allows you to provide a placeholder `{{key}}` param for clearing / failure queries, and this will be replaced with the respective key.

```bash
qjob \
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
    bash -c 'echo the payload is: $QJOB_PAYLOAD'
```

### PostgreSQL

The PostgreSQL driver will retrieve the next message from the specified queue, and pass it to the process. By default, the query used to retrieve the message (`-psql-retrieve-query`) will assume to return a single column, however if you pass `-psql-query-key` it will assume to return a two-column result, with the first column being the key and the second column being the value. This then allows you to provide a placeholder `{{key}}` param for clearing / failure queries, and this will be replaced with the respective key.

```bash
qjob \
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
    bash -c 'echo the payload is: $QJOB_PAYLOAD'
```

### RabbitMQ

The RabbitMQ driver will connect to the specified queue AMQP endpoint and retrieve the next message from the specified queue.

```bash
qjob \
    -rabbitmq-url amqp://guest:guest@localhost:5672 \
    -rabbitmq-queue my-queue \
    -driver rabbitmq \
    bash -c 'echo the payload is: $QJOB_PAYLOAD'
```

### Redis List

The Redis List driver will connect to the specified Redis server and retrieve the next message from the specified list.

```bash
qjob \
    -redis-host localhost \
    -redis-port 6379 \
    -redis-key my-list \
    -driver redis-list \
    bash -c 'echo the payload is: $QJOB_PAYLOAD'
```

### Redis Pub/Sub

The Redis Pub/Sub driver will connect to the specified Redis server and retrieve the next message from the specified subscription.

```bash
qjob \
    -redis-host localhost \
    -redis-port 6379 \
    -redis-key my-subscription \
    -driver redis-pubsub \
    bash -c 'echo the payload is: $QJOB_PAYLOAD'
```

### Local

The local driver is a simple wrapper around the process to execute, primarily for local testing. It does not communicate with any queue, and expects the job payload to be manually defined by the operator as a `QJOB_PAYLOAD` environment variable.

```bash
QJOB_PAYLOAD="my payload" \
qjob \
    -driver local \
    bash -c 'echo the payload is: $QJOB_PAYLOAD'
```

## Orchestration

qjob is solely focused on the worker-side consumption and clearing of work, and intentionally has no scope to the scheduling or management of work.

This allows you to plug in any scheduling or management system you want, and have qjob consume the work from that system.

For example, you can use [keda](https://keda.sh) to monitor your queue and scale qjob worker pods based on the messages in the queue, and when started, qjob will consume and complete the work from the queue.

If you are running in Kubernetes, the [`qjob-operator`](https://github.com/robertlestak/qjob-operator) is a simple operator that will manage QJob workloads on top of Kubernetes and KEDA.

## Deployment

You will need to install qjob in the container which will be used to run your job. You can either compile qjob from source, or use the latest precompiled binaries available.

### Example Dockerfile

```dockerfile
FROM node:17

RUN apt-get update && apt-get install -y \
    curl

RUN curl -LO https://github.com/robertlestak/qjob/releases/latest/download/qjob_linux && \
    chmod +x qjob_linux && \
    mv qjob_linux /usr/local/bin/qjob

RUN echo "console.log('the payload is:', process.env.QJOB_PAYLOAD)" > app.js

CMD ["node", "app.js"]
ENTRYPOINT ["/usr/local/bin/qjob"]
```

```bash
docker build -t qjob .
```

```bash
docker run --rm -it \
    -v ~/.aws:/root/.aws \
    -e QJOB_AWS_REGION=us-east-1 \
    -e QJOB_AWS_SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789012/my-queue \
    -e QJOB_AWS_SQS_ROLE_ARN=arn:aws:iam::123456789012:role/my-role \
    -e QJOB_DRIVER=aws-sqs \
    -e AWS_SDK_LOAD_CONFIG=1 \
    qjob
```