package flags

import "flag"

var (
	FlagSet         = flag.NewFlagSet("procx", flag.ContinueOnError)
	Driver          = FlagSet.String("driver", "", "driver to use. (activemq, aws-dynamo, aws-s3, aws-sqs, cassandra, centauri, cockroach, couchbase, elasticsearch, etcd, fs, gcp-bq, gcp-firestore, gcp-gcs, gcp-pubsub, github, http, kafka, local, mongodb, mssql, mysql, nats, nfs, nsq, postgres, pulsar, rabbitmq, redis-list, redis-pubsub, redis-stream, scylla, smb)")
	HostEnv         = FlagSet.Bool("hostenv", false, "use host environment")
	PassWorkAsArg   = FlagSet.Bool("pass-work-as-arg", false, "pass work as an argument")
	PassWorkAsStdin = FlagSet.Bool("pass-work-as-stdin", false, "pass work as stdin")
	PayloadFile     = FlagSet.String("payload-file", "", "file to write payload to")
	KeepPayloadFile = FlagSet.Bool("keep-payload-file", false, "keep payload file after processing")
	Daemon          = FlagSet.Bool("daemon", false, "run as daemon")
	DaemonInterval  = FlagSet.Int("daemon-interval", 0, "daemon interval in milliseconds")
)
