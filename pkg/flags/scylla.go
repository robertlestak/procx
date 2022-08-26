package flags

var (
	ScyllaHosts          = FlagSet.String("scylla-hosts", "", "Scylla hosts")
	ScyllaUser           = FlagSet.String("scylla-user", "", "Scylla user")
	ScyllaPassword       = FlagSet.String("scylla-password", "", "Scylla password")
	ScyllaKeyspace       = FlagSet.String("scylla-keyspace", "", "Scylla keyspace")
	ScyllaConsistency    = FlagSet.String("scylla-consistency", "QUORUM", "Scylla consistency")
	ScyllaRetrieveQuery  = FlagSet.String("scylla-retrieve-query", "", "Scylla retrieve query")
	ScyllaRetrieveParams = FlagSet.String("scylla-retrieve-params", "", "Scylla retrieve params")
	ScyllaLocalDC        = FlagSet.String("scylla-local-dc", "", "Scylla local dc")
	ScyllaClearQuery     = FlagSet.String("scylla-clear-query", "", "Scylla clear query")
	ScyllaClearParams    = FlagSet.String("scylla-clear-params", "", "Scylla clear params")
	ScyllaFailQuery      = FlagSet.String("scylla-fail-query", "", "Scylla fail query")
	ScyllaFailParams     = FlagSet.String("scylla-fail-params", "", "Scylla fail params")
	ScyllaRetrieveField  = FlagSet.String("scylla-retrieve-field", "", "Scylla retrieve field. If not set, all fields will be returned as a JSON object")
)
