package flags

var (
	SMBHost             = FlagSet.String("smb-host", "", "SMB host")
	SMBPort             = FlagSet.Int("smb-port", 445, "SMB port")
	SMBUser             = FlagSet.String("smb-user", "", "SMB user")
	SMBPass             = FlagSet.String("smb-pass", "", "SMB pass")
	SMBKey              = FlagSet.String("smb-key", "", "SMB key")
	SMBShare            = FlagSet.String("smb-share", "", "SMB share")
	SMBKeyGlob          = FlagSet.String("smb-key-glob", "", "SMB key glob")
	SMBClearOp          = FlagSet.String("smb-clear-op", "", "SMB clear operation. Valid values: mv, rm")
	SMBFailOp           = FlagSet.String("smb-fail-op", "", "SMB fail operation. Valid values: mv, rm")
	SMBClearKey         = FlagSet.String("smb-clear-key", "", "SMB clear key, if clear op is mv. default is origional key name.")
	SMBClearKeyTemplate = FlagSet.String("smb-clear-key-template", "", "SMB clear key template, if clear op is mv.")
	SMBFailKey          = FlagSet.String("smb-fail-key", "", "SMB fail key, if fail op is mv. default is original key name.")
	SMBFailKeyTemplate  = FlagSet.String("smb-fail-key-template", "", "SMB fail key template, if fail op is mv.")
)
