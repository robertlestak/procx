package flags

var (
	ElasticsearchAddress       = FlagSet.String("elasticsearch-address", "", "Elasticsearch address")
	ElasticsearchUsername      = FlagSet.String("elasticsearch-username", "", "Elasticsearch username")
	ElasticsearchPassword      = FlagSet.String("elasticsearch-password", "", "Elasticsearch password")
	ElasticsearchTLSSkipVerify = FlagSet.Bool("elasticsearch-tls-skip-verify", false, "Elasticsearch TLS skip verify")
	ElasticsearchEnableTLS     = FlagSet.Bool("elasticsearch-enable-tls", false, "Elasticsearch enable TLS")
	ElasticsearchCAFile        = FlagSet.String("elasticsearch-tls-ca-file", "", "Elasticsearch TLS CA file")
	ElasticsearchCertFile      = FlagSet.String("elasticsearch-tls-cert-file", "", "Elasticsearch TLS cert file")
	ElasticsearchKeyFile       = FlagSet.String("elasticsearch-tls-key-file", "", "Elasticsearch TLS key file")
	ElasticsearchRetrieveQuery = FlagSet.String("elasticsearch-retrieve-query", "", "Elasticsearch retrieve query")
	ElasticsearchRetrieveIndex = FlagSet.String("elasticsearch-retrieve-index", "", "Elasticsearch retrieve index")
	ElasticsearchClearDoc      = FlagSet.String("elasticsearch-clear-doc", "", "Elasticsearch clear doc")
	ElasticsearchClearIndex    = FlagSet.String("elasticsearch-clear-index", "", "Elasticsearch clear index")
	ElasticsearchClearOp       = FlagSet.String("elasticsearch-clear-op", "", "Elasticsearch clear op. Valid values are: delete, put, merge-put, move")
	ElasticsearchFailDoc       = FlagSet.String("elasticsearch-fail-doc", "", "Elasticsearch fail doc")
	ElasticsearchFailIndex     = FlagSet.String("elasticsearch-fail-index", "", "Elasticsearch fail index")
	ElasticsearchFailOp        = FlagSet.String("elasticsearch-fail-op", "", "Elasticsearch fail op. Valid values are: delete, put, merge-put, move")
)
