package flags

var (
	AWSRegion     = FlagSet.String("aws-region", "", "AWS region")
	AWSLoadConfig = FlagSet.Bool("aws-load-config", false, "load AWS config from ~/.aws/config")
	AWSRoleARN    = FlagSet.String("aws-role-arn", "", "AWS role ARN")
	SQSQueueURL   = FlagSet.String("aws-sqs-queue-url", "", "AWS SQS queue URL")

	AWSDynamoTable         = FlagSet.String("aws-dynamo-table", "", "AWS DynamoDB table name")
	AWSDynamoRetrieveField = FlagSet.String("aws-dynamo-retrieve-field", "", "AWS DynamoDB retrieve field")
	AWSDynamoRetrieveQuery = FlagSet.String("aws-dynamo-retrieve-query", "", "AWS DynamoDB retrieve query")
	AWSDynamoClearQuery    = FlagSet.String("aws-dynamo-clear-query", "", "AWS DynamoDB clear query")
	AWSDynamoFailQuery     = FlagSet.String("aws-dynamo-fail-query", "", "AWS DynamoDB fail query")

	AWSS3Bucket           = FlagSet.String("aws-s3-bucket", "", "AWS S3 bucket")
	AWSS3Key              = FlagSet.String("aws-s3-key", "", "AWS S3 key")
	AWSS3KeyRegex         = FlagSet.String("aws-s3-key-regex", "", "AWS S3 key regex")
	AWSS3KeyPrefix        = FlagSet.String("aws-s3-key-prefix", "", "AWS S3 key prefix")
	AWSS3ClearOp          = FlagSet.String("aws-s3-clear-op", "", "AWS S3 clear operation. Valid values: mv, rm")
	AWSS3FailOp           = FlagSet.String("aws-s3-fail-op", "", "AWS S3 fail operation. Valid values: mv, rm")
	AWSS3ClearBucket      = FlagSet.String("aws-s3-clear-bucket", "", "AWS S3 clear bucket, if clear op is mv")
	AWSS3ClearKey         = FlagSet.String("aws-s3-clear-key", "", "AWS S3 clear key, if clear op is mv. default is origional key name.")
	AWSS3ClearKeyTemplate = FlagSet.String("aws-s3-clear-key-template", "", "AWS S3 clear key template, if clear op is mv.")
	AWSS3FailBucket       = FlagSet.String("aws-s3-fail-bucket", "", "AWS S3 fail bucket, if fail op is mv")
	AWSS3FailKey          = FlagSet.String("aws-s3-fail-key", "", "AWS S3 fail key, if fail op is mv. default is original key name.")
	AWSS3FailKeyTemplate  = FlagSet.String("aws-s3-fail-key-template", "", "AWS S3 fail key template, if fail op is mv.")
)
