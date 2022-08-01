package flags

var (
	GCPProjectID    = FlagSet.String("gcp-project-id", "", "GCP project ID")
	GCPSubscription = FlagSet.String("gcp-pubsub-subscription", "", "GCP Pub/Sub subscription name")
)
