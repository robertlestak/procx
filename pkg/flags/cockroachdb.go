package flags

var (
	CockroachDBRoutingID      = FlagSet.String("cockroach-routing-id", "", "CockroachDB routing id")
	CockroachDBHost           = FlagSet.String("cockroach-host", "", "CockroachDB host")
	CockroachDBPort           = FlagSet.String("cockroach-port", "26257", "CockroachDB port")
	CockroachDBUser           = FlagSet.String("cockroach-user", "", "CockroachDB user")
	CockroachDBPassword       = FlagSet.String("cockroach-password", "", "CockroachDB password")
	CockroachDBDatabase       = FlagSet.String("cockroach-database", "", "CockroachDB database")
	CockroachDBSSLMode        = FlagSet.String("cockroach-ssl-mode", "disable", "CockroachDB SSL mode")
	CockroachDBTLSRootCert    = FlagSet.String("cockroach-tls-root-cert", "", "CockroachDB SSL root cert")
	CockroachDBTLSCert        = FlagSet.String("cockroach-tls-cert", "", "CockroachDB TLS cert")
	CockroachDBTLSKey         = FlagSet.String("cockroach-tls-key", "", "CockroachDB TLS key")
	CockroachDBQueryKey       = FlagSet.Bool("cockroach-query-key", false, "CockroachDB query returns key as first column and value as second column")
	CockroachDBRetrieveField  = FlagSet.String("cockroach-retrieve-field", "", "CockroachDB retrieve field. If not set, all fields will be returned as a JSON object")
	CockroachDBRetrieveQuery  = FlagSet.String("cockroach-retrieve-query", "", "CockroachDB retrieve query")
	CockroachDBRetrieveParams = FlagSet.String("cockroach-retrieve-params", "", "CockroachDB retrieve params")
	CockroachDBClearQuery     = FlagSet.String("cockroach-clear-query", "", "CockroachDB clear query")
	CockroachDBClearParams    = FlagSet.String("cockroach-clear-params", "", "CockroachDB clear params")
	CockroachDBFailQuery      = FlagSet.String("cockroach-fail-query", "", "CockroachDB fail query")
	CockroachDBFailParams     = FlagSet.String("cockroach-fail-params", "", "CockroachDB fail params")
)