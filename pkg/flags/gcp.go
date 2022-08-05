package flags

var (
	GCPProjectID    = FlagSet.String("gcp-project-id", "", "GCP project ID")
	GCPSubscription = FlagSet.String("gcp-pubsub-subscription", "", "GCP Pub/Sub subscription name")

	GCPGCSBucket           = FlagSet.String("gcp-gcs-bucket", "", "GCP GCS bucket")
	GCPGCSKey              = FlagSet.String("gcp-gcs-key", "", "GCP GCS key")
	GCPGCSKeyRegex         = FlagSet.String("gcp-gcs-key-regex", "", "GCP GCS key regex")
	GCPGCSKeyPrefix        = FlagSet.String("gcp-gcs-key-prefix", "", "GCP GCS key prefix")
	GCPGCSClearOp          = FlagSet.String("gcp-gcs-clear-op", "", "GCP GCS clear operation. Valid values: mv, rm")
	GCPGCSFailOp           = FlagSet.String("gcp-gcs-fail-op", "", "GCP GCS fail operation. Valid values: mv, rm")
	GCPGCSClearBucket      = FlagSet.String("gcp-gcs-clear-bucket", "", "GCP GCS clear bucket, if clear op is mv")
	GCPGCSClearKey         = FlagSet.String("gcp-gcs-clear-key", "", "GCP GCS clear key, if clear op is mv. default is origional key name.")
	GCPGCSClearKeyTemplate = FlagSet.String("gcp-gcs-clear-key-template", "", "GCP GCS clear key template, if clear op is mv.")
	GCPGCSFailBucket       = FlagSet.String("gcp-gcs-fail-bucket", "", "GCP GCS fail bucket, if fail op is mv")
	GCPGCSFailKey          = FlagSet.String("gcp-gcs-fail-key", "", "GCP GCS fail key, if fail op is mv. default is original key name.")
	GCPGCSFailKeyTemplate  = FlagSet.String("gcp-gcs-fail-key-template", "", "GCP GCS fail key template, if fail op is mv.")
)
