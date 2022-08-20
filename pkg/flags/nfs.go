package flags

var (
	NFSHost             = FlagSet.String("nfs-host", "", "NFS host")
	NFSKey              = FlagSet.String("nfs-key", "", "NFS key")
	NFSFolder           = FlagSet.String("nfs-folder", "", "NFS folder")
	NFSTarget           = FlagSet.String("nfs-target", "", "NFS target")
	NFSKeyRegex         = FlagSet.String("nfs-key-regex", "", "NFS key regex")
	NFSKeyPrefix        = FlagSet.String("nfs-key-prefix", "", "NFS key prefix")
	NFSClearOp          = FlagSet.String("nfs-clear-op", "", "NFS clear operation. Valid values: mv, rm")
	NFSFailOp           = FlagSet.String("nfs-fail-op", "", "NFS fail operation. Valid values: mv, rm")
	NFSClearFolder      = FlagSet.String("nfs-clear-folder", "", "NFS clear folder, if clear op is mv")
	NFSClearKey         = FlagSet.String("nfs-clear-key", "", "NFS clear key, if clear op is mv. default is origional key name.")
	NFSClearKeyTemplate = FlagSet.String("nfs-clear-key-template", "", "NFS clear key template, if clear op is mv.")
	NFSFailFolder       = FlagSet.String("nfs-fail-folder", "", "NFS fail folder, if fail op is mv")
	NFSFailKey          = FlagSet.String("nfs-fail-key", "", "NFS fail key, if fail op is mv. default is original key name.")
	NFSFailKeyTemplate  = FlagSet.String("nfs-fail-key-template", "", "NFS fail key template, if fail op is mv.")
)
