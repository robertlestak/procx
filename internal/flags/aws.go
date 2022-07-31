package flags

var (
	AWSRegion     = FlagSet.String("aws-region", "", "AWS region")
	AWSLoadConfig = FlagSet.Bool("aws-load-config", false, "load AWS config from ~/.aws/config")
	SQSRoleARN    = FlagSet.String("aws-sqs-role-arn", "", "AWS SQS role ARN")
	SQSQueueURL   = FlagSet.String("aws-sqs-queue-url", "", "AWS SQS queue URL")
)
