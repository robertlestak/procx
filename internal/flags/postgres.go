package flags

var (
	PsqlHost           = FlagSet.String("psql-host", "", "PostgreSQL host")
	PsqlPort           = FlagSet.String("psql-port", "5432", "PostgreSQL port")
	PsqlUser           = FlagSet.String("psql-user", "", "PostgreSQL user")
	PsqlPassword       = FlagSet.String("psql-password", "", "PostgreSQL password")
	PsqlDatabase       = FlagSet.String("psql-database", "", "PostgreSQL database")
	PsqlSSLMode        = FlagSet.String("psql-ssl-mode", "disable", "PostgreSQL SSL mode")
	PsqlQueryKey       = FlagSet.Bool("psql-query-key", false, "PostgreSQL query returns key as first column and value as second column")
	PsqlRetrieveQuery  = FlagSet.String("psql-retrieve-query", "", "PostgreSQL retrieve query")
	PsqlRetrieveParams = FlagSet.String("psql-retrieve-params", "", "PostgreSQL retrieve params")
	PsqlClearQuery     = FlagSet.String("psql-clear-query", "", "PostgreSQL clear query")
	PsqlClearParams    = FlagSet.String("psql-clear-params", "", "PostgreSQL clear params")
	PsqlFailQuery      = FlagSet.String("psql-fail-query", "", "PostgreSQL fail query")
	PsqlFailParams     = FlagSet.String("psql-fail-params", "", "PostgreSQL fail params")
)
