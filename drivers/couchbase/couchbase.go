package couchbase

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/schema"
	"github.com/robertlestak/procx/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type CouchbaseOp string

var (
	CouchbaseOpRM    = CouchbaseOp("rm")
	CouchbaseOpMV    = CouchbaseOp("mv")
	CouchbaseOpSet   = CouchbaseOp("set")
	CouchbaseOpMerge = CouchbaseOp("merge")
)

type CouchbaseDoc struct {
	Op         CouchbaseOp    `json:"op"`
	Doc        map[string]any `json:"doc"`
	Bucket     string         `json:"bucket"`
	Scope      string         `json:"scope"`
	Collection string         `json:"collection"`
	ID         string         `json:"id"`
}

type Couchbase struct {
	Client        *gocb.Cluster
	Address       string
	User          *string
	Password      *string
	BucketName    *string
	Scope         *string
	Collection    *string
	ID            *string
	RetrieveQuery *schema.SqlQuery
	Clear         *CouchbaseDoc
	Fail          *CouchbaseDoc
	// TLS
	EnableTLS   *bool
	TLSInsecure *bool
	TLSCert     *string
	TLSKey      *string
	TLSCA       *string
	doc         []map[string]any
}

func (d *Couchbase) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment variables")
	if os.Getenv(prefix+"COUCHBASE_USER") != "" {
		v := os.Getenv(prefix + "COUCHBASE_USER")
		d.User = &v
	}
	if os.Getenv(prefix+"COUCHBASE_PASSWORD") != "" {
		v := os.Getenv(prefix + "COUCHBASE_PASSWORD")
		d.Password = &v
	}
	if os.Getenv(prefix+"COUCHBASE_COLLECTION") != "" {
		c := os.Getenv(prefix + "COUCHBASE_COLLECTION")
		d.Collection = &c
	}
	if os.Getenv(prefix+"COUCHBASE_SCOPE") != "" {
		s := os.Getenv(prefix + "COUCHBASE_SCOPE")
		d.Scope = &s
	}
	if os.Getenv(prefix+"COUCHBASE_BUCKET_NAME") != "" {
		b := os.Getenv(prefix + "COUCHBASE_BUCKET_NAME")
		d.BucketName = &b
	}
	if os.Getenv(prefix+"COUCHBASE_ID") != "" {
		i := os.Getenv(prefix + "COUCHBASE_ID")
		d.ID = &i
	}
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if os.Getenv(prefix+"COUCHBASE_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery.Query = os.Getenv(prefix + "COUCHBASE_RETRIEVE_QUERY")
	}
	if os.Getenv(prefix+"COUCHBASE_RETRIEVE_PARAMS") != "" {
		p := strings.Split(os.Getenv(prefix+"COUCHBASE_RETRIEVE_PARAMS"), ",")
		for _, v := range p {
			d.RetrieveQuery.Params = append(d.RetrieveQuery.Params, v)
		}
	}
	if os.Getenv(prefix+"COUCHBASE_ENABLE_TLS") != "" {
		v := os.Getenv(prefix+"COUCHBASE_ENABLE_TLS") == "true"
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"COUCHBASE_TLS_INSECURE") != "" {
		v := os.Getenv(prefix+"COUCHBASE_TLS_INSECURE") == "true"
		d.TLSInsecure = &v
	}
	if os.Getenv(prefix+"COUCHBASE_TLS_CERT_FILE") != "" {
		v := os.Getenv(prefix + "COUCHBASE_TLS_CERT_FILE")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"COUCHBASE_TLS_KEY_FILE") != "" {
		v := os.Getenv(prefix + "COUCHBASE_TLS_KEY_FILE")
		d.TLSKey = &v
	}
	if os.Getenv(prefix+"COUCHBASE_TLS_CA_FILE") != "" {
		v := os.Getenv(prefix + "COUCHBASE_TLS_CA_FILE")
		d.TLSCA = &v
	}
	if d.Clear == nil {
		d.Clear = &CouchbaseDoc{}
	}
	if os.Getenv(prefix+"COUCHBASE_CLEAR_OP") != "" {
		d.Clear.Op = CouchbaseOp(os.Getenv(prefix + "COUCHBASE_CLEAR_OP"))
	}
	if os.Getenv(prefix+"COUCHBASE_CLEAR_BUCKET") != "" {
		d.Clear.Bucket = os.Getenv(prefix + "COUCHBASE_CLEAR_BUCKET")
	}
	if os.Getenv(prefix+"COUCHBASE_CLEAR_SCOPE") != "" {
		d.Clear.Scope = os.Getenv(prefix + "COUCHBASE_CLEAR_SCOPE")
	}
	if os.Getenv(prefix+"COUCHBASE_CLEAR_COLLECTION") != "" {
		d.Clear.Collection = os.Getenv(prefix + "COUCHBASE_CLEAR_COLLECTION")
	}
	if os.Getenv(prefix+"COUCHBASE_CLEAR_ID") != "" {
		d.Clear.ID = os.Getenv(prefix + "COUCHBASE_CLEAR_ID")
	}
	if os.Getenv(prefix+"COUCHBASE_CLEAR_DOC") != "" {
		var doc map[string]any
		err := json.Unmarshal([]byte(os.Getenv(prefix+"COUCHBASE_CLEAR_DOC")), &doc)
		if err != nil {
			l.WithError(err).Error("Failed to unmarshal clear doc")
			return err
		}
		d.Clear.Doc = doc
	}
	if d.Fail == nil {
		d.Fail = &CouchbaseDoc{}
	}
	if os.Getenv(prefix+"COUCHBASE_FAIL_OP") != "" {
		d.Fail.Op = CouchbaseOp(os.Getenv(prefix + "COUCHBASE_FAIL_OP"))
	}
	if os.Getenv(prefix+"COUCHBASE_FAIL_BUCKET") != "" {
		d.Fail.Bucket = os.Getenv(prefix + "COUCHBASE_FAIL_BUCKET")
	}
	if os.Getenv(prefix+"COUCHBASE_FAIL_SCOPE") != "" {
		d.Fail.Scope = os.Getenv(prefix + "COUCHBASE_FAIL_SCOPE")
	}
	if os.Getenv(prefix+"COUCHBASE_FAIL_COLLECTION") != "" {
		d.Fail.Collection = os.Getenv(prefix + "COUCHBASE_FAIL_COLLECTION")
	}
	if os.Getenv(prefix+"COUCHBASE_FAIL_ID") != "" {
		d.Fail.ID = os.Getenv(prefix + "COUCHBASE_FAIL_ID")
	}
	if os.Getenv(prefix+"COUCHBASE_FAIL_DOC") != "" {
		var doc map[string]any
		err := json.Unmarshal([]byte(os.Getenv(prefix+"COUCHBASE_FAIL_DOC")), &doc)
		if err != nil {
			l.WithError(err).Error("Failed to unmarshal fail doc")
			return err
		}
		d.Fail.Doc = doc
	}
	return nil
}

func (d *Couchbase) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.User = flags.CouchbaseUser
	d.Password = flags.CouchbasePassword
	d.BucketName = flags.CouchbaseBucketName
	d.Scope = flags.CouchbaseScope
	d.Address = *flags.CouchbaseAddress
	d.Collection = flags.CouchbaseCollection
	d.ID = flags.CouchbaseID
	var rps []any
	if *flags.CouchbaseRetrieveParams != "" {
		s := strings.Split(*flags.CouchbaseRetrieveParams, ",")
		for _, v := range s {
			rps = append(rps, v)
		}
	}
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &schema.SqlQuery{}
	}
	if *flags.CouchbaseRetrieveQuery != "" {
		d.RetrieveQuery.Query = *flags.CouchbaseRetrieveQuery
	}
	d.RetrieveQuery.Params = rps
	d.EnableTLS = flags.CouchbaseEnableTLS
	d.TLSInsecure = flags.CouchbaseTLSInsecure
	d.TLSCert = flags.CouchbaseCertFile
	d.TLSKey = flags.CouchbaseKeyFile
	d.TLSCA = flags.CouchbaseCAFile
	if d.Clear == nil {
		d.Clear = &CouchbaseDoc{}
	}
	d.Clear.Op = CouchbaseOp(*flags.CouchbaseClearOp)
	d.Clear.Bucket = *flags.CouchbaseClearBucket
	d.Clear.Scope = *flags.CouchbaseClearScope
	d.Clear.Collection = *flags.CouchbaseClearCollection
	d.Clear.ID = *flags.CouchbaseClearID
	if *flags.CouchbaseClearDoc != "" && len(d.Clear.Doc) == 0 {
		cd := schema.ReplaceParamsSliceMapString(d.doc, *flags.CouchbaseClearDoc)
		flags.CouchbaseClearDoc = &cd
		if err := json.Unmarshal([]byte(*flags.CouchbaseClearDoc), &d.Clear.Doc); err != nil {
			return err
		}
	}
	if d.Fail == nil {
		d.Fail = &CouchbaseDoc{}
	}
	d.Fail.Op = CouchbaseOp(*flags.CouchbaseFailOp)
	d.Fail.Bucket = *flags.CouchbaseFailBucket
	d.Fail.Scope = *flags.CouchbaseFailScope
	d.Fail.Collection = *flags.CouchbaseFailCollection
	d.Fail.ID = *flags.CouchbaseFailID
	if *flags.CouchbaseFailDoc != "" && len(d.Fail.Doc) == 0 {
		fd := schema.ReplaceParamsSliceMapString(d.doc, *flags.CouchbaseFailDoc)
		flags.CouchbaseFailDoc = &fd
		if err := json.Unmarshal([]byte(*flags.CouchbaseFailDoc), &d.Fail.Doc); err != nil {
			return err
		}
	}
	return nil
}

func (d *Couchbase) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "Init",
	})
	l.Debug("Initializing couchbase client")
	opts := gocb.ClusterOptions{}
	if d.User != nil && *d.User != "" && d.Password != nil && *d.Password != "" {
		opts.Authenticator = gocb.PasswordAuthenticator{
			Username: *d.User,
			Password: *d.Password,
		}
	}
	if d.EnableTLS != nil && *d.EnableTLS {
		tc, err := utils.TlsConfig(d.EnableTLS, d.TLSInsecure, d.TLSCA, d.TLSCert, d.TLSKey)
		if err != nil {
			return err
		}
		sc := gocb.SecurityConfig{
			TLSSkipVerify: d.TLSInsecure != nil && *d.TLSInsecure,
			TLSRootCAs:    tc.RootCAs,
		}
		opts.SecurityConfig = sc
	}
	cluster, err := gocb.Connect(d.Address, opts)
	if err != nil {
		l.Error(err)
		return err
	}
	d.Client = cluster
	return nil
}

func (d *Couchbase) getByID() error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "getByID",
	})
	l.Debug("Getting by ID")
	var err error
	if d.BucketName == nil || *d.BucketName == "" {
		return errors.New("bucket name is required")
	}
	if d.Scope == nil || *d.Scope == "" {
		return errors.New("scope is required")
	}
	if d.Collection == nil || *d.Collection == "" {
		return errors.New("collection is required")
	}
	bucket := d.Client.Bucket(*d.BucketName)
	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		l.Error(err)
		return err
	}
	col := bucket.Scope(*d.Scope).Collection(*d.Collection)
	var doc map[string]any
	res, err := col.Get(*d.ID, nil)
	if err != nil {
		l.Error(err)
		return err
	}
	err = res.Content(&doc)
	if err != nil {
		l.Error(err)
		return err
	}
	d.doc = []map[string]any{doc}
	return nil
}

func (d *Couchbase) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from couchbase")
	if d.ID != nil && *d.ID != "" {
		if err := d.getByID(); err != nil {
			return nil, err
		}
		jd, err := json.Marshal(d.doc)
		if err != nil {
			l.Error(err)
			return nil, err
		}
		return bytes.NewReader(jd), nil
	}
	res, err := d.Client.Query(
		d.RetrieveQuery.Query,
		&gocb.QueryOptions{PositionalParameters: d.RetrieveQuery.Params},
	)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	for res.Next() {
		var r map[string]any
		if err := res.Row(&r); err != nil {
			l.Error(err)
			return nil, err
		}
		d.doc = append(d.doc, r)
	}
	if err := res.Err(); err != nil {
		l.Error(err)
		return nil, err
	}
	jd, err := json.Marshal(d.doc)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	return bytes.NewReader(jd), nil
}

func (d *Couchbase) rmDoc(doc *CouchbaseDoc) error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "rmDoc",
	})
	l.Debug("Removing doc")
	b := d.Client.Bucket(doc.Bucket)
	err := b.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		l.Error(err)
		return err
	}
	col := b.Scope(doc.Scope).Collection(doc.Collection)
	_, err = col.Remove(doc.ID, nil)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *Couchbase) upsertDoc(doc *CouchbaseDoc) error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "upsertDoc",
	})
	l.Debug("Upserting doc")
	b := d.Client.Bucket(doc.Bucket)
	err := b.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		l.Error(err)
		return err
	}
	col := b.Scope(doc.Scope).Collection(doc.Collection)
	_, err = col.Upsert(doc.ID, doc.Doc, nil)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *Couchbase) mvDoc(source *CouchbaseDoc, dest *CouchbaseDoc) error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "mvDoc",
	})
	l.Debug("Moving doc")
	if source.Bucket == dest.Bucket && source.Scope == dest.Scope && source.Collection == dest.Collection {
		return errors.New("source and destination are the same")
	}
	dest.Doc = source.Doc
	if err := d.upsertDoc(dest); err != nil {
		l.Error(err)
		return err
	}
	if err := d.rmDoc(source); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *Couchbase) handleOp(doc *CouchbaseDoc) error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "handleOp",
	})
	l.Debug("Handling op")
	switch doc.Op {
	case CouchbaseOpMV:
		rd := &CouchbaseDoc{
			Bucket:     *d.BucketName,
			Scope:      *d.Scope,
			Collection: *d.Collection,
			ID:         *d.ID,
			Doc:        doc.Doc,
		}
		return d.mvDoc(rd, doc)
	case CouchbaseOpRM:
		return d.rmDoc(doc)
	case CouchbaseOpSet:
		return d.upsertDoc(doc)
	case CouchbaseOpMerge:
		return d.upsertDoc(doc)
	default:
		return errors.New("unknown op")
	}
}

func (d *Couchbase) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from couchbase")
	if d.Clear == nil {
		return nil
	}
	if d.Clear.Bucket == "" {
		d.Clear.Bucket = *d.BucketName
	}
	if d.Clear.Scope == "" {
		d.Clear.Scope = *d.Scope
	}
	if d.Clear.Collection == "" {
		d.Clear.Collection = *d.Collection
	}
	if d.Clear.ID == "" {
		d.Clear.ID = *d.ID
	}
	if d.Clear.Op == "" {
		return nil
	}
	var errs []error
	for _, doc := range d.doc {
		if len(d.Clear.Doc) > 0 && (d.Clear.Op == CouchbaseOpMerge || d.Clear.Op == CouchbaseOpMV) {
			for k, v := range d.Clear.Doc {
				doc[k] = v
			}
			d.Clear.Doc = doc
		}
		if len(d.Clear.Doc) == 0 {
			d.Clear.Doc = doc
		}
		errs = append(errs, d.handleOp(d.Clear))
	}
	for _, err := range errs {
		if err != nil {
			l.Error(err)
			return err
		}
	}
	return nil
}

func (d *Couchbase) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure from couchbase")
	if d.Fail == nil {
		return nil
	}
	if d.Fail.Bucket == "" {
		d.Fail.Bucket = *d.BucketName
	}
	if d.Fail.Scope == "" {
		d.Fail.Scope = *d.Scope
	}
	if d.Fail.Collection == "" {
		d.Fail.Collection = *d.Collection
	}
	if d.Fail.ID == "" {
		d.Fail.ID = *d.ID
	}
	if d.Fail.Op == "" {
		return nil
	}
	var errs []error
	for _, doc := range d.doc {
		if len(d.Fail.Doc) > 0 && (d.Fail.Op == CouchbaseOpMerge || d.Fail.Op == CouchbaseOpMV) {
			for k, v := range d.Fail.Doc {
				doc[k] = v
			}
			d.Fail.Doc = doc
		}
		if len(d.Fail.Doc) == 0 {
			d.Fail.Doc = doc
		}
		errs = append(errs, d.handleOp(d.Fail))
	}
	for _, err := range errs {
		if err != nil {
			l.Error(err)
			return err
		}
	}
	return nil
}

func (d *Couchbase) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "couchbase",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up couchbase")
	if d.Client == nil {
		return nil
	}
	d.Client.Close(nil)
	l.Debug("Cleaned up couchbase")
	return nil
}
