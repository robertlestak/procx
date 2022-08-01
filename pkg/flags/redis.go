package flags

var (
	RedisHost     = FlagSet.String("redis-host", "", "Redis host")
	RedisPort     = FlagSet.String("redis-port", "6379", "Redis port")
	RedisPassword = FlagSet.String("redis-password", "", "Redis password")
	RedisKey      = FlagSet.String("redis-key", "", "Redis key")
)
