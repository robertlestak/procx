package flags

var (
	AWSRegion     = FlagSet.String("aws-region", "", "AWS region")
	AWSLoadConfig = FlagSet.Bool("aws-load-config", false, "load AWS config from ~/.aws/config")
	AWSRoleARN    = FlagSet.String("aws-role-arn", "", "AWS role ARN")
	SQSQueueURL   = FlagSet.String("aws-sqs-queue-url", "", "AWS SQS queue URL")

	AWSDynamoTable         = FlagSet.String("aws-dynamo-table", "", "AWS DynamoDB table name")
	AWSDynamoQueryKeyPath  = FlagSet.String("aws-dynamo-key-path", "", "AWS DynamoDB query key JSON path")
	AWSDynamoDataPath      = FlagSet.String("aws-dynamo-data-path", "", "AWS DynamoDB data JSON path")
	AWSDynamoRetrieveQuery = FlagSet.String("aws-dynamo-retrieve-query", "", "AWS DynamoDB retrieve query")
	AWSDynamoClearQuery    = FlagSet.String("aws-dynamo-clear-query", "", "AWS DynamoDB clear query")
	AWSDynamoFailQuery     = FlagSet.String("aws-dynamo-fail-query", "", "AWS DynamoDB fail query")
)
