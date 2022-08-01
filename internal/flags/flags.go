package flags

import "flag"

var (
	FlagSet       = flag.NewFlagSet("procx", flag.ContinueOnError)
	Driver        = FlagSet.String("driver", "", "driver to use. (aws-sqs, aws-dynamo, cassandra, centauri, gcp-pubsub, local, mongodb, mysql, postgres, rabbitmq, redis-list, redis-pubsub)")
	HostEnv       = FlagSet.Bool("hostenv", false, "use host environment")
	PassWorkAsArg = FlagSet.Bool("pass-work-as-arg", false, "pass work as an argument")
	Daemon        = FlagSet.Bool("daemon", false, "run as daemon")
)
