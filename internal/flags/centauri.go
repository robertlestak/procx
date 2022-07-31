package flags

var (
	CentauriPeerURL = FlagSet.String("centauri-peer-url", "", "Centauri peer URL")
	CentauriChannel = FlagSet.String("centauri-channel", "default", "Centauri channel")
	CentauriKey     = FlagSet.String("centauri-key", "", "Centauri key")
)
