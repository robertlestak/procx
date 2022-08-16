package flags

var (
	FSKey              = FlagSet.String("fs-key", "", "FS key")
	FSFolder           = FlagSet.String("fs-folder", "", "FS folder")
	FSKeyRegex         = FlagSet.String("fs-key-regex", "", "FS key regex")
	FSKeyPrefix        = FlagSet.String("fs-key-prefix", "", "FS key prefix")
	FSClearOp          = FlagSet.String("fs-clear-op", "", "FS clear operation. Valid values: mv, rm")
	FSFailOp           = FlagSet.String("fs-fail-op", "", "FS fail operation. Valid values: mv, rm")
	FSClearFolder      = FlagSet.String("fs-clear-folder", "", "FS clear folder, if clear op is mv")
	FSClearKey         = FlagSet.String("fs-clear-key", "", "FS clear key, if clear op is mv. default is origional key name.")
	FSClearKeyTemplate = FlagSet.String("fs-clear-key-template", "", "FS clear key template, if clear op is mv.")
	FSFailFolder       = FlagSet.String("fs-fail-folder", "", "FS fail folder, if fail op is mv")
	FSFailKey          = FlagSet.String("fs-fail-key", "", "FS fail key, if fail op is mv. default is original key name.")
	FSFailKeyTemplate  = FlagSet.String("fs-fail-key-template", "", "FS fail key template, if fail op is mv.")
)
