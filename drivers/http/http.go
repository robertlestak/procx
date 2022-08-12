package http

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type HTTPRequest struct {
	Method                string
	URL                   string
	ContentType           string
	SuccessfulStatusCodes []int
	Headers               map[string]string
	Body                  io.Reader
}

type RetrieveRequest struct {
	HTTPRequest
	KeyJSONSelector  string
	WorkJSONSelector string
}

type HTTP struct {
	Client          *http.Client
	EnableTLS       *bool
	TLSCA           *string
	TLSCert         *string
	TLSKey          *string
	TLSInsecure     *bool
	RetrieveRequest *RetrieveRequest
	ClearRequest    *HTTPRequest
	FailRequest     *HTTPRequest
	Key             *string
}

func (d *HTTP) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "http",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	if d.RetrieveRequest == nil {
		d.RetrieveRequest = &RetrieveRequest{}
	}
	if d.ClearRequest == nil {
		d.ClearRequest = &HTTPRequest{}
	}
	if d.FailRequest == nil {
		d.FailRequest = &HTTPRequest{}
	}
	if os.Getenv(prefix+"HTTP_RETRIEVE_METHOD") != "" {
		d.RetrieveRequest.Method = os.Getenv(prefix + "HTTP_RETRIEVE_METHOD")
	}
	if os.Getenv(prefix+"HTTP_RETRIEVE_URL") != "" {
		d.RetrieveRequest.URL = os.Getenv(prefix + "HTTP_RETRIEVE_URL")
	}
	if os.Getenv(prefix+"HTTP_RETRIEVE_CONTENT_TYPE") != "" {
		d.RetrieveRequest.ContentType = os.Getenv(prefix + "HTTP_RETRIEVE_CONTENT_TYPE")
	}
	if os.Getenv(prefix+"HTTP_RETRIEVE_SUCCESSFUL_STATUS_CODES") != "" {
		d.RetrieveRequest.SuccessfulStatusCodes = parseIntSlice(os.Getenv(prefix + "HTTP_RETRIEVE_SUCCESSFUL_STATUS_CODES"))
	}
	if os.Getenv(prefix+"HTTP_RETRIEVE_HEADERS") != "" {
		d.RetrieveRequest.Headers = parseHeaderMap(os.Getenv(prefix + "HTTP_RETRIEVE_HEADERS"))
	}
	if os.Getenv(prefix+"HTTP_RETRIEVE_KEY_JSON_SELECTOR") != "" {
		d.RetrieveRequest.KeyJSONSelector = os.Getenv(prefix + "HTTP_RETRIEVE_KEY_JSON_SELECTOR")
	}
	if os.Getenv(prefix+"HTTP_RETRIEVE_WORK_JSON_SELECTOR") != "" {
		d.RetrieveRequest.WorkJSONSelector = os.Getenv(prefix + "HTTP_RETRIEVE_WORK_JSON_SELECTOR")
	}
	if os.Getenv(prefix+"HTTP_RETRIEVE_BODY_FILE") != "" {
		var err error
		d.RetrieveRequest.Body, err = os.Open(os.Getenv(prefix + "HTTP_RETRIEVE_BODY_FILE"))
		if err != nil {
			return err
		}
	}
	if os.Getenv(prefix+"HTTP_RETRIEVE_BODY") != "" {
		d.RetrieveRequest.Body = bytes.NewBufferString(os.Getenv(prefix + "HTTP_RETRIEVE_BODY"))
	}
	if os.Getenv(prefix+"HTTP_CLEAR_METHOD") != "" {
		d.ClearRequest.Method = os.Getenv(prefix + "HTTP_CLEAR_METHOD")
	}
	if os.Getenv(prefix+"HTTP_CLEAR_URL") != "" {
		d.ClearRequest.URL = os.Getenv(prefix + "HTTP_CLEAR_URL")
	}
	if os.Getenv(prefix+"HTTP_CLEAR_CONTENT_TYPE") != "" {
		d.ClearRequest.ContentType = os.Getenv(prefix + "HTTP_CLEAR_CONTENT_TYPE")
	}
	if os.Getenv(prefix+"HTTP_CLEAR_SUCCESSFUL_STATUS_CODES") != "" {
		d.ClearRequest.SuccessfulStatusCodes = parseIntSlice(os.Getenv(prefix + "HTTP_CLEAR_SUCCESSFUL_STATUS_CODES"))
	}
	if os.Getenv(prefix+"HTTP_CLEAR_HEADERS") != "" {
		d.ClearRequest.Headers = parseHeaderMap(os.Getenv(prefix + "HTTP_CLEAR_HEADERS"))
	}
	if os.Getenv(prefix+"HTTP_CLEAR_BODY_FILE") != "" {
		var err error
		d.ClearRequest.Body, err = os.Open(os.Getenv(prefix + "HTTP_CLEAR_BODY_FILE"))
		if err != nil {
			return err
		}
	}
	if os.Getenv(prefix+"HTTP_CLEAR_BODY") != "" {
		d.ClearRequest.Body = bytes.NewBufferString(os.Getenv(prefix + "HTTP_CLEAR_BODY"))
	}
	if os.Getenv(prefix+"HTTP_FAIL_METHOD") != "" {
		d.FailRequest.Method = os.Getenv(prefix + "HTTP_FAIL_METHOD")
	}
	if os.Getenv(prefix+"HTTP_FAIL_URL") != "" {
		d.FailRequest.URL = os.Getenv(prefix + "HTTP_FAIL_URL")
	}
	if os.Getenv(prefix+"HTTP_FAIL_CONTENT_TYPE") != "" {
		d.FailRequest.ContentType = os.Getenv(prefix + "HTTP_FAIL_CONTENT_TYPE")
	}
	if os.Getenv(prefix+"HTTP_FAIL_SUCCESSFUL_STATUS_CODES") != "" {
		d.FailRequest.SuccessfulStatusCodes = parseIntSlice(os.Getenv(prefix + "HTTP_FAIL_SUCCESSFUL_STATUS_CODES"))
	}
	if os.Getenv(prefix+"HTTP_FAIL_HEADERS") != "" {
		d.FailRequest.Headers = parseHeaderMap(os.Getenv(prefix + "HTTP_FAIL_HEADERS"))
	}
	if os.Getenv(prefix+"HTTP_FAIL_BODY_FILE") != "" {
		var err error
		d.FailRequest.Body, err = os.Open(os.Getenv(prefix + "HTTP_FAIL_BODY_FILE"))
		if err != nil {
			return err
		}
	}
	if os.Getenv(prefix+"HTTP_FAIL_BODY") != "" {
		d.FailRequest.Body = bytes.NewBufferString(os.Getenv(prefix + "HTTP_FAIL_BODY"))
	}
	if os.Getenv(prefix+"HTTP_ENABLE_TLS") != "" {
		v := os.Getenv(prefix+"HTTP_ENABLE_TLS") == "true"
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"HTTP_TLS_CERT_FILE") != "" {
		v := os.Getenv(prefix + "HTTP_TLS_CERT_FILE")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"HTTP_TLS_KEY_FILE") != "" {
		v := os.Getenv(prefix + "HTTP_TLS_KEY_FILE")
		d.TLSKey = &v
	}
	if os.Getenv(prefix+"HTTP_TLS_CA_FILE") != "" {
		v := os.Getenv(prefix + "HTTP_TLS_CA_FILE")
		d.TLSCA = &v
	}
	return nil
}

func parseIntSlice(s string) []int {
	var r []int
	for _, v := range strings.Split(s, ",") {
		i, e := strconv.Atoi(v)
		if e != nil {
			continue
		}
		r = append(r, i)
	}
	return r
}

func parseHeaderMap(s string) map[string]string {
	r := make(map[string]string)
	for _, v := range strings.Split(s, ",") {
		kv := strings.Split(v, ":")
		if len(kv) != 2 {
			continue
		}
		r[kv[0]] = kv[1]
	}
	return r
}

func (d *HTTP) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "http",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	rr := &RetrieveRequest{
		HTTPRequest: HTTPRequest{
			Method:                *flags.HTTPRetreieveMethod,
			URL:                   *flags.HTTPRetrieveURL,
			ContentType:           *flags.HTTPRetrieveContentType,
			SuccessfulStatusCodes: parseIntSlice(*flags.HTTPRetrieveSuccessfulStatusCodes),
			Headers:               parseHeaderMap(*flags.HTTPRetrieveHeaders),
		},
		KeyJSONSelector:  *flags.HTTPRetrieveKeyJSONSelector,
		WorkJSONSelector: *flags.HTTPRetrieveWorkJSONSelector,
	}
	if *flags.HTTPRetrieveBodyFile != "" {
		var err error
		rr.Body, err = os.Open(*flags.HTTPRetrieveBodyFile)
		if err != nil {
			return err
		}
	}
	if *flags.HTTPRetrieveBody != "" {
		rr.Body = bytes.NewBufferString(*flags.HTTPRetrieveBody)
	}
	if rr.HTTPRequest.Method == "" {
		rr.HTTPRequest.Method = "GET"
	}
	d.RetrieveRequest = rr
	cr := &HTTPRequest{
		Method:                *flags.HTTPClearMethod,
		URL:                   *flags.HTTPClearURL,
		ContentType:           *flags.HTTPClearContentType,
		SuccessfulStatusCodes: parseIntSlice(*flags.HTTPClearSuccessfulStatusCodes),
		Headers:               parseHeaderMap(*flags.HTTPClearHeaders),
	}
	if cr.Method == "" {
		cr.Method = "GET"
	}
	if *flags.HTTPClearBodyFile != "" {
		var err error
		cr.Body, err = os.Open(*flags.HTTPClearBodyFile)
		if err != nil {
			return err
		}
	}
	if *flags.HTTPClearBody != "" {
		cr.Body = bytes.NewBufferString(*flags.HTTPClearBody)
	}
	d.ClearRequest = cr
	fr := &HTTPRequest{
		Method:                *flags.HTTPFailMethod,
		URL:                   *flags.HTTPFailURL,
		ContentType:           *flags.HTTPFailContentType,
		SuccessfulStatusCodes: parseIntSlice(*flags.HTTPFailSuccessfulStatusCodes),
		Headers:               parseHeaderMap(*flags.HTTPFailHeaders),
	}
	if fr.Method == "" {
		fr.Method = "GET"
	}
	if *flags.HTTPFailBodyFile != "" {
		var err error
		fr.Body, err = os.Open(*flags.HTTPFailBodyFile)
		if err != nil {
			return err
		}
	}
	if *flags.HTTPFailBody != "" {
		fr.Body = bytes.NewBufferString(*flags.HTTPFailBody)
	}
	d.FailRequest = fr
	d.EnableTLS = flags.HTTPEnableTLS
	d.TLSCert = flags.HTTPTLSCertFile
	d.TLSKey = flags.HTTPTLSKeyFile
	d.TLSCA = flags.HTTPTLSCAFile
	return nil
}

func (d *HTTP) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "http",
		"fn":  "Init",
	})
	l.Debug("Initializing http driver")
	d.Client = &http.Client{}
	tc, err := utils.TlsConfig(d.EnableTLS, d.TLSInsecure, d.TLSCA, d.TLSCert, d.TLSKey)
	if err != nil {
		return err
	}
	d.Client.Transport = &http.Transport{
		TLSClientConfig: tc,
	}
	return nil
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (d *HTTP) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "http",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from http")
	if d.RetrieveRequest == nil {
		return nil, errors.New("RetrieveRequest is nil")
	}
	if d.RetrieveRequest.HTTPRequest.Method == "" {
		d.RetrieveRequest.HTTPRequest.Method = "GET"
	}
	if d.RetrieveRequest.HTTPRequest.URL == "" {
		return nil, errors.New("URL is nil")
	}
	req, err := http.NewRequest(d.RetrieveRequest.Method, d.RetrieveRequest.URL, d.RetrieveRequest.Body)
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	for k, v := range d.RetrieveRequest.Headers {
		req.Header.Add(k, v)
	}
	if d.RetrieveRequest.ContentType != "" {
		req.Header.Add("Content-Type", d.RetrieveRequest.ContentType)
	}
	resp, err := d.Client.Do(req)
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	if len(d.RetrieveRequest.SuccessfulStatusCodes) > 0 {
		if !contains(d.RetrieveRequest.SuccessfulStatusCodes, resp.StatusCode) {
			l.Errorf("Status code %d not in successful status codes", resp.StatusCode)
			return nil, err
		}
	}
	if d.RetrieveRequest.KeyJSONSelector != "" || d.RetrieveRequest.WorkJSONSelector != "" {
		bd, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Errorf("%+v", err)
			return nil, err
		}
		if d.RetrieveRequest.KeyJSONSelector != "" {
			key := gjson.ParseBytes(bd).Get(d.RetrieveRequest.KeyJSONSelector)
			if !key.Exists() {
				l.Errorf("Key not found in json")
				return nil, errors.New("key not found in json")
			}
			s := key.String()
			d.Key = &s
		}
		if d.RetrieveRequest.WorkJSONSelector != "" {
			work := gjson.ParseBytes(bd).Get(d.RetrieveRequest.WorkJSONSelector)
			if !work.Exists() {
				l.Errorf("Work not found in json")
				return nil, errors.New("work not found in json")
			}
			return strings.NewReader(work.String()), nil
		}
	}
	return resp.Body, nil
}

func (d *HTTP) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "http",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from http")
	if d.ClearRequest == nil {
		return nil
	}
	if d.ClearRequest.URL == "" {
		return nil
	}
	if d.ClearRequest.Method == "" {
		d.ClearRequest.Method = "GET"
	}
	if d.ClearRequest.URL == "" {
		return errors.New("URL is nil")
	}
	if d.Key != nil && *d.Key != "" {
		d.ClearRequest.URL = strings.Replace(d.ClearRequest.URL, "{{key}}", *d.Key, -1)
		if d.ClearRequest.Body != nil {
			bd, err := ioutil.ReadAll(d.ClearRequest.Body)
			if err != nil {
				l.Errorf("%+v", err)
				return err
			}
			bd = []byte(strings.Replace(string(bd), "{{key}}", *d.Key, -1))
			d.ClearRequest.Body = ioutil.NopCloser(bytes.NewReader(bd))
		}
	}
	req, err := http.NewRequest(d.ClearRequest.Method, d.ClearRequest.URL, d.ClearRequest.Body)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	for k, v := range d.ClearRequest.Headers {
		req.Header.Add(k, v)
	}
	if d.ClearRequest.ContentType != "" {
		req.Header.Add("Content-Type", d.ClearRequest.ContentType)
	}
	_, err = d.Client.Do(req)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	return nil
}

func (d *HTTP) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "http",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from http")
	if d.FailRequest == nil {
		return nil
	}
	if d.FailRequest.URL == "" {
		return nil
	}
	if d.FailRequest.Method == "" {
		d.FailRequest.Method = "GET"
	}
	if d.FailRequest.URL == "" {
		return errors.New("URL is nil")
	}
	if d.Key != nil && *d.Key != "" {
		d.FailRequest.URL = strings.Replace(d.FailRequest.URL, "{{key}}", *d.Key, -1)
		bd, err := ioutil.ReadAll(d.FailRequest.Body)
		if err != nil {
			l.Errorf("%+v", err)
			return err
		}
		bd = []byte(strings.Replace(string(bd), "{{key}}", *d.Key, -1))
		d.FailRequest.Body = ioutil.NopCloser(bytes.NewReader(bd))
	}
	req, err := http.NewRequest(d.FailRequest.Method, d.FailRequest.URL, d.FailRequest.Body)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	for k, v := range d.FailRequest.Headers {
		req.Header.Add(k, v)
	}
	if d.FailRequest.ContentType != "" {
		req.Header.Add("Content-Type", d.FailRequest.ContentType)
	}
	_, err = d.Client.Do(req)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	return nil
}

func (d *HTTP) Cleanup() error {
	return nil
}
