package flags

var (
	CentauriPeerURL   = FlagSet.String("centauri-peer-url", "", "Centauri peer URL")
	CentauriChannel   = FlagSet.String("centauri-channel", "default", "Centauri channel")
	CentauriKey       = FlagSet.String("centauri-key", "", "Centauri key")
	CentauriKeyBase64 = FlagSet.String("centauri-key-base64", "", "Centauri key base64")
)
