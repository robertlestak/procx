package flags

var (
	MSSqlHost           = FlagSet.String("mssql-host", "", "MSSQL host")
	MSSqlPort           = FlagSet.String("mssql-port", "1433", "MSSQL port")
	MSSqlUser           = FlagSet.String("mssql-user", "", "MSSQL user")
	MSSqlPassword       = FlagSet.String("mssql-password", "", "MSSQL password")
	MSSqlDatabase       = FlagSet.String("mssql-database", "", "MSSQL database")
	MSSqlRetrieveField  = FlagSet.String("mssql-retrieve-field", "", "MSSQL retrieve field. If not set, all fields will be returned as a JSON object")
	MSSqlRetrieveQuery  = FlagSet.String("mssql-retrieve-query", "", "MSSQL retrieve query")
	MSSqlRetrieveParams = FlagSet.String("mssql-retrieve-params", "", "MSSQL retrieve params")
	MSSqlClearQuery     = FlagSet.String("mssql-clear-query", "", "MSSQL clear query")
	MSSqlClearParams    = FlagSet.String("mssql-clear-params", "", "MSSQL clear params")
	MSSqlFailQuery      = FlagSet.String("mssql-fail-query", "", "MSSQL fail query")
	MSSqlFailParams     = FlagSet.String("mssql-fail-params", "", "MSSQL fail params")
)
