package flags

var (
	NatsURL           = FlagSet.String("nats-url", "", "Nats URL")
	NatsSubject       = FlagSet.String("nats-subject", "", "Nats subject")
	NatsCredsFile     = FlagSet.String("nats-creds-file", "", "Nats creds file")
	NatsJWTFile       = FlagSet.String("nats-jwt-file", "", "Nats JWT file")
	NatsNKeyFile      = FlagSet.String("nats-nkey-file", "", "Nats NKey file")
	NatsUsername      = FlagSet.String("nats-username", "", "Nats username")
	NatsPassword      = FlagSet.String("nats-password", "", "Nats password")
	NatsToken         = FlagSet.String("nats-token", "", "Nats token")
	NatsEnableTLS     = FlagSet.Bool("nats-enable-tls", false, "Nats enable TLS")
	NatsTLSInsecure   = FlagSet.Bool("nats-tls-insecure", false, "Nats TLS insecure")
	NatsTLSCAFile     = FlagSet.String("nats-tls-ca-file", "", "Nats TLS CA file")
	NatsTLSCertFile   = FlagSet.String("nats-tls-cert-file", "", "Nats TLS cert file")
	NatsTLSKeyFile    = FlagSet.String("nats-tls-key-file", "", "Nats TLS key file")
	NatsClearResponse = FlagSet.String("nats-clear-response", "", "Nats clear response")
	NatsFailResponse  = FlagSet.String("nats-fail-response", "", "Nats fail response")
)
