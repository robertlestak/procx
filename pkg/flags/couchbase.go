package flags

var (
	CouchbaseAddress         = FlagSet.String("couchbase-address", "", "Couchbase address")
	CouchbaseUser            = FlagSet.String("couchbase-user", "", "Couchbase user")
	CouchbasePassword        = FlagSet.String("couchbase-password", "", "Couchbase password")
	CouchbaseBucketName      = FlagSet.String("couchbase-bucket", "", "Couchbase bucket name")
	CouchbaseScope           = FlagSet.String("couchbase-scope", "_default", "Couchbase scope")
	CouchbaseCollection      = FlagSet.String("couchbase-collection", "_default", "Couchbase collection")
	CouchbaseRetrieveQuery   = FlagSet.String("couchbase-retrieve-query", "", "Couchbase retrieve query")
	CouchbaseRetrieveParams  = FlagSet.String("couchbase-retrieve-params", "", "Couchbase retrieve params")
	CouchbaseID              = FlagSet.String("couchbase-id", "", "Couchbase id")
	CouchbaseClearOp         = FlagSet.String("couchbase-clear-op", "", "Couchbase clear op. one of: mv, rm, set, merge")
	CouchbaseClearDoc        = FlagSet.String("couchbase-clear-doc", "", "Couchbase clear doc, if op is set or merge")
	CouchbaseClearBucket     = FlagSet.String("couchbase-clear-bucket", "", "Couchbase clear bucket, if op is set or merge. Default to the current bucket.")
	CouchbaseClearScope      = FlagSet.String("couchbase-clear-scope", "_default", "Couchbase clear scope, default to the current scope.")
	CouchbaseClearCollection = FlagSet.String("couchbase-clear-collection", "_default", "Couchbase clear collection, default to the current collection.")
	CouchbaseClearID         = FlagSet.String("couchbase-clear-id", "", "Couchbase clear id, default to the current id.")

	CouchbaseFailOp         = FlagSet.String("couchbase-fail-op", "", "Couchbase fail op. one of: mv, rm, set, merge")
	CouchbaseFailDoc        = FlagSet.String("couchbase-fail-doc", "", "Couchbase fail doc, if op is set or merge")
	CouchbaseFailBucket     = FlagSet.String("couchbase-fail-bucket", "", "Couchbase fail bucket, if op is set or merge. Default to the current bucket.")
	CouchbaseFailScope      = FlagSet.String("couchbase-fail-scope", "_default", "Couchbase fail scope, default to the current scope.")
	CouchbaseFailCollection = FlagSet.String("couchbase-fail-collection", "_default", "Couchbase fail collection, default to the current collection.")
	CouchbaseFailID         = FlagSet.String("couchbase-fail-id", "", "Couchbase fail id, default to the current id.")

	// TLS
	CouchbaseEnableTLS   = FlagSet.Bool("couchbase-enable-tls", false, "Enable TLS")
	CouchbaseTLSInsecure = FlagSet.Bool("couchbase-tls-insecure", false, "Enable TLS insecure")
	CouchbaseCAFile      = FlagSet.String("couchbase-tls-ca-file", "", "Couchbase TLS CA file")
	CouchbaseCertFile    = FlagSet.String("couchbase-tls-cert-file", "", "Couchbase TLS cert file")
	CouchbaseKeyFile     = FlagSet.String("couchbase-tls-key-file", "", "Couchbase TLS key file")
)
