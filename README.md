# qjob - simple job queue worker

qjob is a small process manager that can wrap around any existing application / script / process, and integrate with a job queue system to enable autoscaling job executions with no native code integration.

qjob is a single compiled binary that can be packaged in your existing job code container. qjob is configured with environment variables or command line flags, and is started with the path to the process to execute.

qjob will retrieve the next job from the queue, and pass it to the process. Upon success (exit code 0), qjob will mark the job as complete. Upon failure (exit code != 0), qjob will mark the job as failed to be requeued.

qjob will make the job payload available in the `QJOB_PAYLOAD` environment variable. If `-pass-work-as-arg` is set, the job payload string will be appended to the process arguments.

By default, the subprocess spawned by qjob will not have access to the host environment variables. This can be changed by setting the `-hostenv` flag.

## Usage

```bash
qjob [flags] <process path>
  -aws-region string
        AWS region
  -aws-sqs-queue-url string
        AWS SQS queue URL
  -aws-sqs-role-arn string
        AWS SQS role ARN
  -driver string
        driver to use. (aws-sqs, rabbitmq, local)
  -hostenv
        use host environment
  -pass-work-as-arg
        pass work as an argument
  -rabbitmq-queue string
        RabbitMQ queue
  -rabbitmq-url string
        RabbitMQ URL
```

### Environment Variables

- QJOB_AWS_REGION
        AWS region
- QJOB_AWS_SQS_QUEUE_URL
        AWS SQS queue URL
- QJOB_AWS_SQS_ROLE_ARN
        AWS SQS role ARN
- QJOB_DRIVER
        driver to use
- QJOB_HOSTENV
        use host environment
- QJOB_PASS_WORK_AS_ARG
        pass work as an argument
- QJOB_RABBITMQ_URL
        RabbitMQ URL
- QJOB_RABBITMQ_QUEUE
        RabbitMQ queue

## Drivers

Currently, the following drivers are supported:

- AWS SQS (`aws-sqs`)
- RabbitMQ (`rabbitmq`)
- Local (`local`)

Plans to add more drivers in the future, and PRs are welcome.

### AWS SQS

The SQS driver will retrieve the next message from the specified queue, and pass it to the process. Upon successful completion of the process, it will delete the message from the queue.

For cross-account access, you must provide the ARN of the role that has access to the queue, and the identity running qjob must be able to assume the target identity.

```bash
qjob \
    -aws-sqs-queue-url https://sqs.us-east-1.amazonaws.com/123456789012/my-queue \
    -aws-sqs-role-arn arn:aws:iam::123456789012:role/my-role \
    -aws-region us-east-1 \
    -driver aws-sqs \
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

### Local

The local driver is a simple wrapper around the process to execute, primarily for local testing. It does not communicate with any queue, and expects the job payload to be manually defined by the operator.

## Orchestration

qjob is solely focused on the worker-side consumption and clearing of work, and intentionally has no scope to the scheduling or management of work.

This allows you to plug in any scheduling or management system you want, and have qjob consume the work from that system.

For example, you can use [keda](https://keda.sh) to monitor your queue and scale qjob worker pods based on the messages in the queue, and when started, qjob will consume and complete the work from the queue.

## Deployment

You will need to install qjob in the container which will be used to run your job. You can either compile qjob from source, or use the latest precompiled binaries available.

### Example Dockerfile

```dockerfile
FROM node:17

RUN apt-get update && apt-get install -y \
    curl

RUN curl -LO https://github.com/robertlestak/qjob/releases/download/v0.0.1/qjob_linux_amd64 && \
    chmod +x qjob_linux_amd64 && \
    mv qjob_linux_amd64 /usr/local/bin/qjob

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