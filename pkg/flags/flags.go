package flags

import "flag"

var (
	FlagSet         = flag.NewFlagSet("procx", flag.ContinueOnError)
	Driver          = FlagSet.String("driver", "", "driver to use. (aws-dynamo, aws-s3, aws-sqs, cassandra, centauri, elasticsearch, gcp-bq, gcp-gcs, gcp-pubsub, kafka, local, mongodb, mysql, nats, nfs, postgres, rabbitmq, redis-list, redis-pubsub)")
	HostEnv         = FlagSet.Bool("hostenv", false, "use host environment")
	PassWorkAsArg   = FlagSet.Bool("pass-work-as-arg", false, "pass work as an argument")
	PassWorkAsStdin = FlagSet.Bool("pass-work-as-stdin", false, "pass work as stdin")
	PayloadFile     = FlagSet.String("payload-file", "", "file to write payload to")
	KeepPayloadFile = FlagSet.Bool("keep-payload-file", false, "keep payload file after processing")
	Daemon          = FlagSet.Bool("daemon", false, "run as daemon")
)
