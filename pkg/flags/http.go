package flags

var (
	HTTPRetreieveMethod               = FlagSet.String("http-retrieve-method", "GET", "HTTP retrieve method")
	HTTPRetrieveURL                   = FlagSet.String("http-retrieve-url", "", "HTTP retrieve url")
	HTTPRetrieveContentType           = FlagSet.String("http-retrieve-content-type", "", "HTTP retrieve content type")
	HTTPRetrieveSuccessfulStatusCodes = FlagSet.String("http-retrieve-successful-status-codes", "", "HTTP retrieve successful status codes")
	HTTPRetrieveHeaders               = FlagSet.String("http-retrieve-headers", "", "HTTP retrieve headers")
	HTTPRetrieveKeyJSONSelector       = FlagSet.String("http-retrieve-key-json-selector", "", "HTTP retrieve key json selector")
	HTTPRetrieveWorkJSONSelector      = FlagSet.String("http-retrieve-work-json-selector", "", "HTTP retrieve work json selector")
	HTTPRetrieveBodyFile              = FlagSet.String("http-retrieve-body-file", "", "HTTP retrieve body file")
	HTTPRetrieveBody                  = FlagSet.String("http-retrieve-body", "", "HTTP retrieve body")
	HTTPClearMethod                   = FlagSet.String("http-clear-method", "GET", "HTTP clear method")
	HTTPClearURL                      = FlagSet.String("http-clear-url", "", "HTTP clear url")
	HTTPClearContentType              = FlagSet.String("http-clear-content-type", "", "HTTP clear content type")
	HTTPClearSuccessfulStatusCodes    = FlagSet.String("http-clear-successful-status-codes", "", "HTTP clear successful status codes")
	HTTPClearHeaders                  = FlagSet.String("http-clear-headers", "", "HTTP clear headers")
	HTTPClearBodyFile                 = FlagSet.String("http-clear-body-file", "", "HTTP clear body file")
	HTTPClearBody                     = FlagSet.String("http-clear-body", "", "HTTP clear body")
	HTTPFailMethod                    = FlagSet.String("http-fail-method", "GET", "HTTP fail method")
	HTTPFailURL                       = FlagSet.String("http-fail-url", "", "HTTP fail url")
	HTTPFailContentType               = FlagSet.String("http-fail-content-type", "", "HTTP fail content type")
	HTTPFailSuccessfulStatusCodes     = FlagSet.String("http-fail-successful-status-codes", "", "HTTP fail successful status codes")
	HTTPFailHeaders                   = FlagSet.String("http-fail-headers", "", "HTTP fail headers")
	HTTPFailBodyFile                  = FlagSet.String("http-fail-body-file", "", "HTTP fail body file")
	HTTPFailBody                      = FlagSet.String("http-fail-body", "", "HTTP fail body")
	HTTPEnableTLS                     = FlagSet.Bool("http-enable-tls", false, "HTTP enable tls")
	HTTPTLSInsecure                   = FlagSet.Bool("http-tls-insecure", false, "HTTP tls insecure")
	HTTPTLSCertFile                   = FlagSet.String("http-tls-cert-file", "", "HTTP tls cert file")
	HTTPTLSKeyFile                    = FlagSet.String("http-tls-key-file", "", "HTTP tls key file")
	HTTPTLSCAFile                     = FlagSet.String("http-tls-ca-file", "", "HTTP tls ca file")
)
