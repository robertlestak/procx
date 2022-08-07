package flags

var (
	ElasticsearchAddress       = FlagSet.String("elasticsearch-address", "", "Elasticsearch address")
	ElasticsearchUsername      = FlagSet.String("elasticsearch-username", "", "Elasticsearch username")
	ElasticsearchPassword      = FlagSet.String("elasticsearch-password", "", "Elasticsearch password")
	ElasticsearchTLSSkipVerify = FlagSet.Bool("elasticsearch-tls-skip-verify", false, "Elasticsearch TLS skip verify")
	ElasticsearchRetrieveQuery = FlagSet.String("elasticsearch-retrieve-query", "", "Elasticsearch retrieve query")
	ElasticsearchRetrieveIndex = FlagSet.String("elasticsearch-retrieve-index", "", "Elasticsearch retrieve index")
	ElasticsearchClearQuery    = FlagSet.String("elasticsearch-clear-query", "", "Elasticsearch clear query")
	ElasticsearchClearIndex    = FlagSet.String("elasticsearch-clear-index", "", "Elasticsearch clear index")
	ElasticsearchClearOp       = FlagSet.String("elasticsearch-clear-op", "", "Elasticsearch clear op. Valid values are: delete, put, merge-put, move")
	ElasticsearchFailQuery     = FlagSet.String("elasticsearch-fail-query", "", "Elasticsearch fail query")
	ElasticsearchFailIndex     = FlagSet.String("elasticsearch-fail-index", "", "Elasticsearch fail index")
	ElasticsearchFailOp        = FlagSet.String("elasticsearch-fail-op", "", "Elasticsearch fail op. Valid values are: delete, put, merge-put, move")
)
