package flags

var (
	MysqlHost           = FlagSet.String("mysql-host", "", "MySQL host")
	MysqlPort           = FlagSet.String("mysql-port", "3306", "MySQL port")
	MysqlUser           = FlagSet.String("mysql-user", "", "MySQL user")
	MysqlPassword       = FlagSet.String("mysql-password", "", "MySQL password")
	MysqlDatabase       = FlagSet.String("mysql-database", "", "MySQL database")
	MysqlQueryKey       = FlagSet.Bool("mysql-query-key", false, "MySQL query returns key as first column and value as second column")
	MysqlRetrieveQuery  = FlagSet.String("mysql-retrieve-query", "", "MySQL retrieve query")
	MysqlRetrieveParams = FlagSet.String("mysql-retrieve-params", "", "MySQL retrieve params")
	MysqlClearQuery     = FlagSet.String("mysql-clear-query", "", "MySQL clear query")
	MysqlClearParams    = FlagSet.String("mysql-clear-params", "", "MySQL clear params")
	MysqlFailQuery      = FlagSet.String("mysql-fail-query", "", "MySQL fail query")
	MysqlFailParams     = FlagSet.String("mysql-fail-params", "", "MySQL fail params")
)
