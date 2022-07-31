package main

import "os"

func parseAWSFlags() {
	if os.Getenv("QJOB_AWS_REGION") != "" {
		r := os.Getenv("QJOB_AWS_REGION")
		flagAWSRegion = &r
	}
	if os.Getenv("QJOB_AWS_SQS_ROLE_ARN") != "" {
		r := os.Getenv("QJOB_AWS_SQS_ROLE_ARN")
		flagSQSRoleARN = &r
	}
	if os.Getenv("QJOB_AWS_SQS_QUEUE_URL") != "" {
		r := os.Getenv("QJOB_AWS_SQS_QUEUE_URL")
		flagSQSQueueURL = &r
	}
	if os.Getenv("QJOB_AWS_LOAD_CONFIG") != "" || os.Getenv("AWS_SDK_LOAD_CONFIG") != "" {
		r := os.Getenv("QJOB_AWS_LOAD_CONFIG")
		t := r == "true"
		flagAWSLoadConfig = &t
	}
	if *flagAWSLoadConfig {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
}

func parseBaseFlags() {
	if os.Getenv("QJOB_DRIVER") != "" {
		d := os.Getenv("QJOB_DRIVER")
		flagDriver = &d
	}
	if os.Getenv("QJOB_HOSTENV") != "" {
		h := os.Getenv("QJOB_HOSTENV")
		t := h == "true"
		flagHostEnv = &t
	}
	if os.Getenv("QJOB_PASS_WORK_AS_ARG") != "" {
		r := os.Getenv("QJOB_PASS_WORK_AS_ARG")
		t := r == "true"
		flagPassWorkAsArg = &t
	}
	if os.Getenv("QJOB_DAEMON") != "" {
		r := os.Getenv("QJOB_DAEMON")
		t := r == "true"
		flagDaemon = &t
	}
}

func parseRabbitFlags() {
	if os.Getenv("QJOB_RABBITMQ_URL") != "" {
		r := os.Getenv("QJOB_RABBITMQ_URL")
		flagRabbitMQURL = &r
	}
	if os.Getenv("QJOB_RABBITMQ_QUEUE") != "" {
		r := os.Getenv("QJOB_RABBITMQ_QUEUE")
		flagRabbitMQQueue = &r
	}
}

func parseCentauriFlags() {
	if os.Getenv("QJOB_CENTAURI_PEER_URL") != "" {
		r := os.Getenv("QJOB_CENTAURI_PEER_URL")
		flagCentauriPeerURL = &r
	}
	if os.Getenv("QJOB_CENTAURI_CHANNEL") != "" {
		r := os.Getenv("QJOB_CENTAURI_CHANNEL")
		flagCentauriChannel = &r
	}
	if os.Getenv("QJOB_CENTAURI_KEY") != "" {
		r := os.Getenv("QJOB_CENTAURI_KEY")
		flagCentauriKey = &r
	}
}

func parseRedisFlags() {
	if os.Getenv("QJOB_REDIS_HOST") != "" {
		r := os.Getenv("QJOB_REDIS_HOST")
		flagRedisHost = &r
	}
	if os.Getenv("QJOB_REDIS_PORT") != "" {
		r := os.Getenv("QJOB_REDIS_PORT")
		flagRedisPort = &r
	}
	if os.Getenv("QJOB_REDIS_PASSWORD") != "" {
		r := os.Getenv("QJOB_REDIS_PASSWORD")
		flagRedisPassword = &r
	}
	if os.Getenv("QJOB_REDIS_KEY") != "" {
		r := os.Getenv("QJOB_REDIS_KEY")
		flagRedisKey = &r
	}
}

func parseGCPFlags() {
	if os.Getenv("QJOB_GCP_PROJECT_ID") != "" {
		r := os.Getenv("QJOB_GCP_PROJECT_ID")
		flagGCPProjectID = &r
	}
	if os.Getenv("QJOB_GCP_SUBSCRIPTION") != "" {
		r := os.Getenv("QJOB_GCP_SUBSCRIPTION")
		flagGCPSubscription = &r
	}
}

func parsePsqlFlags() {

	if os.Getenv("QJOB_PSQL_HOST") != "" {
		r := os.Getenv("QJOB_PSQL_HOST")
		flagPsqlHost = &r
	}
	if os.Getenv("QJOB_PSQL_PORT") != "" {
		r := os.Getenv("QJOB_PSQL_PORT")
		flagPsqlPort = &r
	}
	if os.Getenv("QJOB_PSQL_USER") != "" {
		r := os.Getenv("QJOB_PSQL_USER")
		flagPsqlUser = &r
	}
	if os.Getenv("QJOB_PSQL_PASSWORD") != "" {
		r := os.Getenv("QJOB_PSQL_PASSWORD")
		flagPsqlPassword = &r
	}
	if os.Getenv("QJOB_PSQL_DATABASE") != "" {
		r := os.Getenv("QJOB_PSQL_DATABASE")
		flagPsqlDatabase = &r
	}
	if os.Getenv("QJOB_PSQL_SSL_MODE") != "" {
		r := os.Getenv("QJOB_PSQL_SSL_MODE")
		flagPsqlSSLMode = &r
	}
	if os.Getenv("QJOB_PSQL_RETRIEVE_QUERY") != "" {
		r := os.Getenv("QJOB_PSQL_RETRIEVE_QUERY")
		flagPsqlRetrieveQuery = &r
	}
	if os.Getenv("QJOB_PSQL_CLEAR_QUERY") != "" {
		r := os.Getenv("QJOB_PSQL_CLEAR_QUERY")
		flagPsqlClearQuery = &r
	}
	if os.Getenv("QJOB_PSQL_FAIL_QUERY") != "" {
		r := os.Getenv("QJOB_PSQL_FAIL_QUERY")
		flagPsqlFailQuery = &r
	}
	if os.Getenv("QJOB_PSQL_RETRIEVE_PARAMS") != "" {
		r := os.Getenv("QJOB_PSQL_RETRIEVE_PARAMS")
		flagPsqlRetrieveParams = &r
	}
	if os.Getenv("QJOB_PSQL_CLEAR_PARAMS") != "" {
		r := os.Getenv("QJOB_PSQL_CLEAR_PARAMS")
		flagPsqlClearParams = &r
	}
	if os.Getenv("QJOB_PSQL_FAIL_PARAMS") != "" {
		r := os.Getenv("QJOB_PSQL_FAIL_PARAMS")
		flagPsqlFailParams = &r
	}
	if os.Getenv("QJOB_PSQL_QUERY_KEY") != "" {
		r := os.Getenv("QJOB_PSQL_QUERY_KEY")
		t := r == "true"
		flagPsqlQueryKey = &t
	}
}

func parseMongoFlags() {
	if os.Getenv("QJOB_MONGO_HOST") != "" {
		r := os.Getenv("QJOB_MONGO_HOST")
		flagMongoHost = &r
	}
	if os.Getenv("QJOB_MONGO_PORT") != "" {
		r := os.Getenv("QJOB_MONGO_PORT")
		flagMongoPort = &r
	}
	if os.Getenv("QJOB_MONGO_USER") != "" {
		r := os.Getenv("QJOB_MONGO_USER")
		flagMongoUser = &r
	}
	if os.Getenv("QJOB_MONGO_PASSWORD") != "" {
		r := os.Getenv("QJOB_MONGO_PASSWORD")
		flagMongoPassword = &r
	}
	if os.Getenv("QJOB_MONGO_DATABASE") != "" {
		r := os.Getenv("QJOB_MONGO_DATABASE")
		flagMongoDatabase = &r
	}
	if os.Getenv("QJOB_MONGO_RETRIEVE_QUERY") != "" {
		r := os.Getenv("QJOB_MONGO_RETRIEVE_QUERY")
		flagMongoRetrieveQuery = &r
	}
	if os.Getenv("QJOB_MONGO_CLEAR_QUERY") != "" {
		r := os.Getenv("QJOB_MONGO_CLEAR_QUERY")
		flagMongoClearQuery = &r
	}
	if os.Getenv("QJOB_MONGO_FAIL_QUERY") != "" {
		r := os.Getenv("QJOB_MONGO_FAIL_QUERY")
		flagMongoFailQuery = &r
	}
	if os.Getenv("QJOB_MONGO_COLLECTION") != "" {
		r := os.Getenv("QJOB_MONGO_COLLECTION")
		flagMongoCollection = &r
	}
}

func parseMysqlFlags() {

	if os.Getenv("QJOB_MYSQL_HOST") != "" {
		r := os.Getenv("QJOB_MYSQL_HOST")
		flagPsqlHost = &r
	}
	if os.Getenv("QJOB_MYSQL_PORT") != "" {
		r := os.Getenv("QJOB_MYSQL_PORT")
		flagPsqlPort = &r
	}
	if os.Getenv("QJOB_MYSQL_USER") != "" {
		r := os.Getenv("QJOB_MYSQL_USER")
		flagPsqlUser = &r
	}
	if os.Getenv("QJOB_MYSQL_PASSWORD") != "" {
		r := os.Getenv("QJOB_MYSQL_PASSWORD")
		flagPsqlPassword = &r
	}
	if os.Getenv("QJOB_MYSQL_DATABASE") != "" {
		r := os.Getenv("QJOB_MYSQL_DATABASE")
		flagPsqlDatabase = &r
	}
	if os.Getenv("QJOB_MYSQL_SSL_MODE") != "" {
		r := os.Getenv("QJOB_MYSQL_SSL_MODE")
		flagPsqlSSLMode = &r
	}
	if os.Getenv("QJOB_MYSQL_RETRIEVE_QUERY") != "" {
		r := os.Getenv("QJOB_MYSQL_RETRIEVE_QUERY")
		flagPsqlRetrieveQuery = &r
	}
	if os.Getenv("QJOB_MYSQL_CLEAR_QUERY") != "" {
		r := os.Getenv("QJOB_MYSQL_CLEAR_QUERY")
		flagPsqlClearQuery = &r
	}
	if os.Getenv("QJOB_MYSQL_FAIL_QUERY") != "" {
		r := os.Getenv("QJOB_MYSQL_FAIL_QUERY")
		flagPsqlFailQuery = &r
	}
	if os.Getenv("QJOB_MYSQL_RETRIEVE_PARAMS") != "" {
		r := os.Getenv("QJOB_MYSQL_RETRIEVE_PARAMS")
		flagPsqlRetrieveParams = &r
	}
	if os.Getenv("QJOB_MYSQL_CLEAR_PARAMS") != "" {
		r := os.Getenv("QJOB_MYSQL_CLEAR_PARAMS")
		flagPsqlClearParams = &r
	}
	if os.Getenv("QJOB_MYSQL_FAIL_PARAMS") != "" {
		r := os.Getenv("QJOB_MYSQL_FAIL_PARAMS")
		flagPsqlFailParams = &r
	}
	if os.Getenv("QJOB_MYSQL_QUERY_KEY") != "" {
		r := os.Getenv("QJOB_MYSQL_QUERY_KEY")
		t := r == "true"
		flagPsqlQueryKey = &t
	}
}

func parseEnvToFlags() {
	parseBaseFlags()
	parseAWSFlags()
	parseCentauriFlags()
	parseRabbitFlags()
	parseRedisFlags()
	parseGCPFlags()
	parsePsqlFlags()
	parseMongoFlags()
	parseMysqlFlags()
}
