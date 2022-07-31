package main

import "os"

func parseAWSFlags() {
	if os.Getenv("PROCX_AWS_REGION") != "" {
		r := os.Getenv("PROCX_AWS_REGION")
		flagAWSRegion = &r
	}
	if os.Getenv("PROCX_AWS_SQS_ROLE_ARN") != "" {
		r := os.Getenv("PROCX_AWS_SQS_ROLE_ARN")
		flagSQSRoleARN = &r
	}
	if os.Getenv("PROCX_AWS_SQS_QUEUE_URL") != "" {
		r := os.Getenv("PROCX_AWS_SQS_QUEUE_URL")
		flagSQSQueueURL = &r
	}
	if os.Getenv("PROCX_AWS_LOAD_CONFIG") != "" || os.Getenv("AWS_SDK_LOAD_CONFIG") != "" {
		r := os.Getenv("PROCX_AWS_LOAD_CONFIG")
		t := r == "true"
		flagAWSLoadConfig = &t
	}
	if *flagAWSLoadConfig {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
}

func parseBaseFlags() {
	if os.Getenv("PROCX_DRIVER") != "" {
		d := os.Getenv("PROCX_DRIVER")
		flagDriver = &d
	}
	if os.Getenv("PROCX_HOSTENV") != "" {
		h := os.Getenv("PROCX_HOSTENV")
		t := h == "true"
		flagHostEnv = &t
	}
	if os.Getenv("PROCX_PASS_WORK_AS_ARG") != "" {
		r := os.Getenv("PROCX_PASS_WORK_AS_ARG")
		t := r == "true"
		flagPassWorkAsArg = &t
	}
	if os.Getenv("PROCX_DAEMON") != "" {
		r := os.Getenv("PROCX_DAEMON")
		t := r == "true"
		flagDaemon = &t
	}
}

func parseRabbitFlags() {
	if os.Getenv("PROCX_RABBITMQ_URL") != "" {
		r := os.Getenv("PROCX_RABBITMQ_URL")
		flagRabbitMQURL = &r
	}
	if os.Getenv("PROCX_RABBITMQ_QUEUE") != "" {
		r := os.Getenv("PROCX_RABBITMQ_QUEUE")
		flagRabbitMQQueue = &r
	}
}

func parseCassandraFlags() {
	if os.Getenv("PROCX_CASSANDRA_HOSTS") != "" {
		r := os.Getenv("PROCX_CASSANDRA_HOSTS")
		flagCassandraHosts = &r
	}
	if os.Getenv("PROCX_CASSANDRA_KEYSPACE") != "" {
		r := os.Getenv("PROCX_CASSANDRA_KEYSPACE")
		flagCassandraKeyspace = &r
	}
	if os.Getenv("PROCX_CASSANDRA_USER") != "" {
		r := os.Getenv("PROCX_CASSANDRA_USER")
		flagCassandraUser = &r
	}
	if os.Getenv("PROCX_CASSANDRA_PASSWORD") != "" {
		r := os.Getenv("PROCX_CASSANDRA_PASSWORD")
		flagCassandraPassword = &r
	}
	if os.Getenv("PROCX_CASSANDRA_CONSISTENCY") != "" {
		r := os.Getenv("PROCX_CASSANDRA_CONSISTENCY")
		flagCassandraConsistency = &r
	}
	if os.Getenv("PROCX_CASSANDRA_RETRIEVE_QUERY") != "" {
		r := os.Getenv("PROCX_CASSANDRA_RETRIEVE_QUERY")
		flagCassandraRetrieveQuery = &r
	}
	if os.Getenv("PROCX_CASSANDRA_RETRIEVE_PARAMS") != "" {
		r := os.Getenv("PROCX_CASSANDRA_RETRIEVE_PARAMS")
		flagCassandraRetrieveParams = &r
	}
	if os.Getenv("PROCX_CASSANDRA_CLEAR_QUERY") != "" {
		r := os.Getenv("PROCX_CASSANDRA_CLEAR_QUERY")
		flagCassandraClearQuery = &r
	}
	if os.Getenv("PROCX_CASSANDRA_CLEAR_PARAMS") != "" {
		r := os.Getenv("PROCX_CASSANDRA_CLEAR_PARAMS")
		flagCassandraClearParams = &r
	}
	if os.Getenv("PROCX_CASSANDRA_FAIL_QUERY") != "" {
		r := os.Getenv("PROCX_CASSANDRA_FAIL_QUERY")
		flagCassandraFailQuery = &r
	}
	if os.Getenv("PROCX_CASSANDRA_FAIL_PARAMS") != "" {
		r := os.Getenv("PROCX_CASSANDRA_FAIL_PARAMS")
		flagCassandraFailParams = &r
	}
	if os.Getenv("PROCX_CASSANDRA_QUERY_KEY") != "" {
		r := os.Getenv("PROCX_CASSANDRA_QUERY_KEY")
		t := r == "true"
		flagCassandraQueryKey = &t
	}
}

func parseCentauriFlags() {
	if os.Getenv("PROCX_CENTAURI_PEER_URL") != "" {
		r := os.Getenv("PROCX_CENTAURI_PEER_URL")
		flagCentauriPeerURL = &r
	}
	if os.Getenv("PROCX_CENTAURI_CHANNEL") != "" {
		r := os.Getenv("PROCX_CENTAURI_CHANNEL")
		flagCentauriChannel = &r
	}
	if os.Getenv("PROCX_CENTAURI_KEY") != "" {
		r := os.Getenv("PROCX_CENTAURI_KEY")
		flagCentauriKey = &r
	}
}

func parseRedisFlags() {
	if os.Getenv("PROCX_REDIS_HOST") != "" {
		r := os.Getenv("PROCX_REDIS_HOST")
		flagRedisHost = &r
	}
	if os.Getenv("PROCX_REDIS_PORT") != "" {
		r := os.Getenv("PROCX_REDIS_PORT")
		flagRedisPort = &r
	}
	if os.Getenv("PROCX_REDIS_PASSWORD") != "" {
		r := os.Getenv("PROCX_REDIS_PASSWORD")
		flagRedisPassword = &r
	}
	if os.Getenv("PROCX_REDIS_KEY") != "" {
		r := os.Getenv("PROCX_REDIS_KEY")
		flagRedisKey = &r
	}
}

func parseGCPFlags() {
	if os.Getenv("PROCX_GCP_PROJECT_ID") != "" {
		r := os.Getenv("PROCX_GCP_PROJECT_ID")
		flagGCPProjectID = &r
	}
	if os.Getenv("PROCX_GCP_SUBSCRIPTION") != "" {
		r := os.Getenv("PROCX_GCP_SUBSCRIPTION")
		flagGCPSubscription = &r
	}
}

func parsePsqlFlags() {

	if os.Getenv("PROCX_PSQL_HOST") != "" {
		r := os.Getenv("PROCX_PSQL_HOST")
		flagPsqlHost = &r
	}
	if os.Getenv("PROCX_PSQL_PORT") != "" {
		r := os.Getenv("PROCX_PSQL_PORT")
		flagPsqlPort = &r
	}
	if os.Getenv("PROCX_PSQL_USER") != "" {
		r := os.Getenv("PROCX_PSQL_USER")
		flagPsqlUser = &r
	}
	if os.Getenv("PROCX_PSQL_PASSWORD") != "" {
		r := os.Getenv("PROCX_PSQL_PASSWORD")
		flagPsqlPassword = &r
	}
	if os.Getenv("PROCX_PSQL_DATABASE") != "" {
		r := os.Getenv("PROCX_PSQL_DATABASE")
		flagPsqlDatabase = &r
	}
	if os.Getenv("PROCX_PSQL_SSL_MODE") != "" {
		r := os.Getenv("PROCX_PSQL_SSL_MODE")
		flagPsqlSSLMode = &r
	}
	if os.Getenv("PROCX_PSQL_RETRIEVE_QUERY") != "" {
		r := os.Getenv("PROCX_PSQL_RETRIEVE_QUERY")
		flagPsqlRetrieveQuery = &r
	}
	if os.Getenv("PROCX_PSQL_CLEAR_QUERY") != "" {
		r := os.Getenv("PROCX_PSQL_CLEAR_QUERY")
		flagPsqlClearQuery = &r
	}
	if os.Getenv("PROCX_PSQL_FAIL_QUERY") != "" {
		r := os.Getenv("PROCX_PSQL_FAIL_QUERY")
		flagPsqlFailQuery = &r
	}
	if os.Getenv("PROCX_PSQL_RETRIEVE_PARAMS") != "" {
		r := os.Getenv("PROCX_PSQL_RETRIEVE_PARAMS")
		flagPsqlRetrieveParams = &r
	}
	if os.Getenv("PROCX_PSQL_CLEAR_PARAMS") != "" {
		r := os.Getenv("PROCX_PSQL_CLEAR_PARAMS")
		flagPsqlClearParams = &r
	}
	if os.Getenv("PROCX_PSQL_FAIL_PARAMS") != "" {
		r := os.Getenv("PROCX_PSQL_FAIL_PARAMS")
		flagPsqlFailParams = &r
	}
	if os.Getenv("PROCX_PSQL_QUERY_KEY") != "" {
		r := os.Getenv("PROCX_PSQL_QUERY_KEY")
		t := r == "true"
		flagPsqlQueryKey = &t
	}
}

func parseMongoFlags() {
	if os.Getenv("PROCX_MONGO_HOST") != "" {
		r := os.Getenv("PROCX_MONGO_HOST")
		flagMongoHost = &r
	}
	if os.Getenv("PROCX_MONGO_PORT") != "" {
		r := os.Getenv("PROCX_MONGO_PORT")
		flagMongoPort = &r
	}
	if os.Getenv("PROCX_MONGO_USER") != "" {
		r := os.Getenv("PROCX_MONGO_USER")
		flagMongoUser = &r
	}
	if os.Getenv("PROCX_MONGO_PASSWORD") != "" {
		r := os.Getenv("PROCX_MONGO_PASSWORD")
		flagMongoPassword = &r
	}
	if os.Getenv("PROCX_MONGO_DATABASE") != "" {
		r := os.Getenv("PROCX_MONGO_DATABASE")
		flagMongoDatabase = &r
	}
	if os.Getenv("PROCX_MONGO_RETRIEVE_QUERY") != "" {
		r := os.Getenv("PROCX_MONGO_RETRIEVE_QUERY")
		flagMongoRetrieveQuery = &r
	}
	if os.Getenv("PROCX_MONGO_CLEAR_QUERY") != "" {
		r := os.Getenv("PROCX_MONGO_CLEAR_QUERY")
		flagMongoClearQuery = &r
	}
	if os.Getenv("PROCX_MONGO_FAIL_QUERY") != "" {
		r := os.Getenv("PROCX_MONGO_FAIL_QUERY")
		flagMongoFailQuery = &r
	}
	if os.Getenv("PROCX_MONGO_COLLECTION") != "" {
		r := os.Getenv("PROCX_MONGO_COLLECTION")
		flagMongoCollection = &r
	}
}

func parseMysqlFlags() {

	if os.Getenv("PROCX_MYSQL_HOST") != "" {
		r := os.Getenv("PROCX_MYSQL_HOST")
		flagPsqlHost = &r
	}
	if os.Getenv("PROCX_MYSQL_PORT") != "" {
		r := os.Getenv("PROCX_MYSQL_PORT")
		flagPsqlPort = &r
	}
	if os.Getenv("PROCX_MYSQL_USER") != "" {
		r := os.Getenv("PROCX_MYSQL_USER")
		flagPsqlUser = &r
	}
	if os.Getenv("PROCX_MYSQL_PASSWORD") != "" {
		r := os.Getenv("PROCX_MYSQL_PASSWORD")
		flagPsqlPassword = &r
	}
	if os.Getenv("PROCX_MYSQL_DATABASE") != "" {
		r := os.Getenv("PROCX_MYSQL_DATABASE")
		flagPsqlDatabase = &r
	}
	if os.Getenv("PROCX_MYSQL_SSL_MODE") != "" {
		r := os.Getenv("PROCX_MYSQL_SSL_MODE")
		flagPsqlSSLMode = &r
	}
	if os.Getenv("PROCX_MYSQL_RETRIEVE_QUERY") != "" {
		r := os.Getenv("PROCX_MYSQL_RETRIEVE_QUERY")
		flagPsqlRetrieveQuery = &r
	}
	if os.Getenv("PROCX_MYSQL_CLEAR_QUERY") != "" {
		r := os.Getenv("PROCX_MYSQL_CLEAR_QUERY")
		flagPsqlClearQuery = &r
	}
	if os.Getenv("PROCX_MYSQL_FAIL_QUERY") != "" {
		r := os.Getenv("PROCX_MYSQL_FAIL_QUERY")
		flagPsqlFailQuery = &r
	}
	if os.Getenv("PROCX_MYSQL_RETRIEVE_PARAMS") != "" {
		r := os.Getenv("PROCX_MYSQL_RETRIEVE_PARAMS")
		flagPsqlRetrieveParams = &r
	}
	if os.Getenv("PROCX_MYSQL_CLEAR_PARAMS") != "" {
		r := os.Getenv("PROCX_MYSQL_CLEAR_PARAMS")
		flagPsqlClearParams = &r
	}
	if os.Getenv("PROCX_MYSQL_FAIL_PARAMS") != "" {
		r := os.Getenv("PROCX_MYSQL_FAIL_PARAMS")
		flagPsqlFailParams = &r
	}
	if os.Getenv("PROCX_MYSQL_QUERY_KEY") != "" {
		r := os.Getenv("PROCX_MYSQL_QUERY_KEY")
		t := r == "true"
		flagPsqlQueryKey = &t
	}
}

func parseEnvToFlags() {
	parseBaseFlags()
	parseAWSFlags()
	parseCassandraFlags()
	parseCentauriFlags()
	parseRabbitFlags()
	parseRedisFlags()
	parseGCPFlags()
	parsePsqlFlags()
	parseMongoFlags()
	parseMysqlFlags()
}
