package flags

var (
	KafkaBrokers      = FlagSet.String("kafka-brokers", "", "Kafka brokers, comma separated")
	KafkaGroup        = FlagSet.String("kafka-group", "", "Kafka group")
	KafkaTopic        = FlagSet.String("kafka-topic", "", "Kafka topic")
	KafkaEnableTLS    = FlagSet.Bool("kafka-enable-tls", false, "Enable TLS")
	KafkaTLSInsecure  = FlagSet.Bool("kafka-tls-insecure", false, "Enable TLS insecure")
	KafkaCAFile       = FlagSet.String("kafka-ca-file", "", "Kafka CA file")
	KafkaCertFile     = FlagSet.String("kafka-cert-file", "", "Kafka cert file")
	KafkaKeyFile      = FlagSet.String("kafka-key-file", "", "Kafka key file")
	KafkaEnableSasl   = FlagSet.Bool("kafka-enable-sasl", false, "Enable SASL")
	KafkaSaslType     = FlagSet.String("kafka-sasl-type", "", "Kafka SASL type. Can be either 'scram' or 'plain'")
	KafkaSaslUsername = FlagSet.String("kafka-sasl-username", "", "Kafka SASL user")
	KafkaSaslPassword = FlagSet.String("kafka-sasl-password", "", "Kafka SASL password")
)
