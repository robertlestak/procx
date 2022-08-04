package flags

var (
	MongoHost          = FlagSet.String("mongo-host", "", "MongoDB host")
	MongoPort          = FlagSet.String("mongo-port", "27017", "MongoDB port")
	MongoUser          = FlagSet.String("mongo-user", "", "MongoDB user")
	MongoPassword      = FlagSet.String("mongo-password", "", "MongoDB password")
	MongoDatabase      = FlagSet.String("mongo-database", "", "MongoDB database")
	MongoCollection    = FlagSet.String("mongo-collection", "", "MongoDB collection")
	MongoRetrieveQuery = FlagSet.String("mongo-retrieve-query", "", "MongoDB retrieve query")
	MongoClearQuery    = FlagSet.String("mongo-clear-query", "", "MongoDB clear query")
	MongoFailQuery     = FlagSet.String("mongo-fail-query", "", "MongoDB fail query")
)
