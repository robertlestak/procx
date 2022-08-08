package flags

var (
	RedisHost     = FlagSet.String("redis-host", "", "Redis host")
	RedisPort     = FlagSet.String("redis-port", "6379", "Redis port")
	RedisPassword = FlagSet.String("redis-password", "", "Redis password")
	RedisKey      = FlagSet.String("redis-key", "", "Redis key")

	RedisEnableTLS     = FlagSet.Bool("redis-enable-tls", false, "Enable TLS")
	RedisTLSSkipVerify = FlagSet.Bool("redis-tls-skip-verify", false, "Redis TLS skip verify")
	RedisCAFile        = FlagSet.String("redis-tls-ca-file", "", "Redis TLS CA file")
	RedisCertFile      = FlagSet.String("redis-tls-cert-file", "", "Redis TLS cert file")
	RedisKeyFile       = FlagSet.String("redis-tls-key-file", "", "Redis TLS key file")

	RedisConsumerGroup = FlagSet.String("redis-stream-consumer-group", "", "Redis consumer group")
	RedisConsumerName  = FlagSet.String("redis-steam-consumer-name", "", "Redis consumer name. Default is a random UUID")
	RedisValueKeys     = FlagSet.String("redis-stream-value-keys", "", "Redis stream value keys to select. Comma separated, default all.")
	RedisClearOp       = FlagSet.String("redis-stream-clear-op", "", "Redis clear operation. Valid values are 'ack' and 'del'.")
	RedisFailOp        = FlagSet.String("redis-stream-fail-op", "", "Redis fail operation. Valid values are 'ack' and 'del'.")
)
