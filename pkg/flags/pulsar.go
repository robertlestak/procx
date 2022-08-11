package flags

var (
	PulsarAddress                    = FlagSet.String("pulsar-address", "", "Pulsar address")
	PulsarSubscription               = FlagSet.String("pulsar-subscription", "", "Pulsar subscription name")
	PulsarTopic                      = FlagSet.String("pulsar-topic", "", "Pulsar topic")
	PulsarTopicsPattern              = FlagSet.String("pulsar-topics-pattern", "", "Pulsar topics pattern")
	PulsarTopics                     = FlagSet.String("pulsar-topics", "", "Pulsar topics, comma separated")
	PulsarTLSTrustCertsFilePath      = FlagSet.String("pulsar-tls-trust-certs-file", "", "Pulsar TLS trust certs file path")
	PulsarTLSAllowInsecureConnection = FlagSet.Bool("pulsar-tls-allow-insecure-connection", false, "Pulsar TLS allow insecure connection")
	PulsarTLSValidateHostname        = FlagSet.Bool("pulsar-tls-validate-hostname", false, "Pulsar TLS validate hostname")
	PulsarAuthToken                  = FlagSet.String("pulsar-auth-token", "", "Pulsar auth token")
	PulsarAuthTokenFile              = FlagSet.String("pulsar-auth-token-file", "", "Pulsar auth token file")
	PulsarAuthCertFile               = FlagSet.String("pulsar-auth-cert-file", "", "Pulsar auth cert file")
	PulsarAuthKeyFile                = FlagSet.String("pulsar-auth-key-file", "", "Pulsar auth key file")
	PulsarAuthOAuthParams            = FlagSet.String("pulsar-auth-oauth-params", "", "Pulsar auth oauth params")
)
