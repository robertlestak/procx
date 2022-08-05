package flags

import "flag"

var (
	FlagSet         = flag.NewFlagSet("procx", flag.ContinueOnError)
	Driver          = FlagSet.String("driver", "", "driver to use. (aws-dynamo, aws-s3, aws-sqs, cassandra, centauri, gcp-gcs, gcp-pubsub, local, mongodb, mysql, postgres, rabbitmq, redis-list, redis-pubsub)")
	HostEnv         = FlagSet.Bool("hostenv", false, "use host environment")
	PassWorkAsArg   = FlagSet.Bool("pass-work-as-arg", false, "pass work as an argument")
	PayloadFile     = FlagSet.String("payload-file", "", "file to write payload to")
	KeepPayloadFile = FlagSet.Bool("keep-payload-file", false, "keep payload file after processing")
	Daemon          = FlagSet.Bool("daemon", false, "run as daemon")
)
