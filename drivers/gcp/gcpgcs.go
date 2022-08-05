package gcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

type GCSOperation string

var (
	GCSOperationRM = GCSOperation("rm")
	GCSOperationMV = GCSOperation("mv")
)

type GCSOp struct {
	Operation   GCSOperation
	Bucket      string
	Key         string
	KeyTemplate string
}

type GCS struct {
	Client    *storage.Client
	Bucket    string
	Key       string
	KeyRegex  string
	KeyPrefix string
	ClearOp   *GCSOp
	FailOp    *GCSOp
}

func (o *GCSOp) GetKey() string {
	if o.KeyTemplate == "" && o.Key != "" {
		return o.Key
	}
	return strings.ReplaceAll(o.KeyTemplate, "{{key}}", o.Key)
}

func (d *GCS) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "LoadEnv",
	})
	l.Debug("LoadEnv")
	if os.Getenv(prefix+"GCP_GCS_BUCKET") != "" {
		d.Bucket = os.Getenv(prefix + "GCP_GCS_BUCKET")
	}
	if os.Getenv(prefix+"GCP_GCS_KEY") != "" {
		d.Key = os.Getenv(prefix + "GCP_GCS_KEY")
	}
	if os.Getenv(prefix+"GCP_GCS_KEY_REGEX") != "" {
		d.KeyRegex = os.Getenv(prefix + "GCP_GCS_KEY_REGEX")
	}
	if os.Getenv(prefix+"GCP_GCS_KEY_PREFIX") != "" {
		d.KeyPrefix = os.Getenv(prefix + "GCP_GCS_KEY_PREFIX")
	}
	if os.Getenv(prefix+"GCP_GCS_CLEAR_OP") != "" {
		d.ClearOp = &GCSOp{}
		d.ClearOp.Operation = GCSOperation(os.Getenv(prefix + "GCP_GCS_CLEAR_OP"))
	}
	if os.Getenv(prefix+"GCP_GCS_FAIL_OP") != "" {
		d.FailOp = &GCSOp{}
		d.FailOp.Operation = GCSOperation(os.Getenv(prefix + "GCP_GCS_FAIL_OP"))
	}
	if os.Getenv(prefix+"GCP_GCS_CLEAR_BUCKET") != "" {
		d.ClearOp.Bucket = os.Getenv(prefix + "GCP_GCS_CLEAR_BUCKET")
	}
	if os.Getenv(prefix+"GCP_GCS_CLEAR_KEY") != "" {
		d.ClearOp.Key = os.Getenv(prefix + "GCP_GCS_CLEAR_KEY")
	}
	if os.Getenv(prefix+"GCP_GCS_CLEAR_KEY_TEMPLATE") != "" {
		d.ClearOp.KeyTemplate = os.Getenv(prefix + "GCP_GCS_CLEAR_KEY_TEMPLATE")
	}
	if os.Getenv(prefix+"GCP_GCS_FAIL_BUCKET") != "" {
		d.FailOp.Bucket = os.Getenv(prefix + "GCP_GCS_FAIL_BUCKET")
	}
	if os.Getenv(prefix+"GCP_GCS_FAIL_KEY") != "" {
		d.FailOp.Key = os.Getenv(prefix + "GCP_GCS_FAIL_KEY")
	}
	if os.Getenv(prefix+"GCP_GCS_FAIL_KEY_TEMPLATE") != "" {
		d.FailOp.KeyTemplate = os.Getenv(prefix + "GCP_GCS_FAIL_KEY_TEMPLATE")
	}
	return nil
}

func (d *GCS) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.Bucket = *flags.GCPGCSBucket
	d.Key = *flags.GCPGCSKey
	if flags.GCPGCSKeyRegex != nil {
		d.KeyRegex = *flags.GCPGCSKeyRegex
	}
	if flags.GCPGCSKeyPrefix != nil {
		d.KeyPrefix = *flags.GCPGCSKeyPrefix
	}
	if flags.GCPGCSClearOp != nil {
		d.ClearOp = &GCSOp{}
		d.ClearOp.Operation = GCSOperation(*flags.GCPGCSClearOp)
		d.ClearOp.Bucket = *flags.GCPGCSClearBucket
		d.ClearOp.Key = *flags.GCPGCSClearKey
		d.ClearOp.KeyTemplate = *flags.GCPGCSClearKeyTemplate
	}
	if flags.GCPGCSFailOp != nil {
		d.FailOp = &GCSOp{}
		d.FailOp.Operation = GCSOperation(*flags.GCPGCSFailOp)
		d.FailOp.Bucket = *flags.GCPGCSFailBucket
		d.FailOp.Key = *flags.GCPGCSFailKey
		d.FailOp.KeyTemplate = *flags.GCPGCSFailKeyTemplate
	}
	return nil
}

func (d *GCS) Init() error {
	l := log.WithFields(
		log.Fields{
			"pkg": "gcp",
			"fn":  "CreateGCPSession",
		},
	)
	l.Debug("CreateGCPSession")
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	d.Client = client
	return err
}

func (d *GCS) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "GetWork",
	})
	l.Debug("GetWork")
	if d.Key != "" {
		l.Debugf("GetWork key=%s", d.Key)
		return d.getObject()
	} else if d.KeyPrefix != "" {
		l.Debugf("GetWork keyPrefix=%s", d.KeyPrefix)
		return d.findObjectByPrefix()
	} else if d.KeyRegex != "" {
		l.Debugf("GetWork keyRegex=%s", d.KeyRegex)
		return d.findObjectByRegex()
	} else {
		l.Debug("GetWork no key")
		return nil, errors.New("no key")
	}
}

func (d *GCS) getObject() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "getObject",
	})
	l.Debug("getObject")
	if d.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}
	if d.Key == "" {
		return nil, fmt.Errorf("key is required")
	}
	ctx := context.Background()
	rc, err := d.Client.Bucket(d.Bucket).Object(d.Key).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (d *GCS) findObjectByPrefix() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "findObjectByPrefix",
	})
	l.Debug("findObjectByPrefix")
	if d.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}
	if d.KeyPrefix == "" {
		return nil, fmt.Errorf("key prefix is required")
	}
	ctx := context.Background()
	q := &storage.Query{
		Prefix: d.KeyPrefix,
	}
	it := d.Client.Bucket(d.Bucket).Objects(ctx, q)
	var foundKey string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			l.Errorf("findObjectByPrefix error: %s", err)
			return nil, err
		}
		foundKey = attrs.Name
		break
	}
	if foundKey == "" {
		return nil, nil
	}
	d.Key = foundKey
	return d.getObject()
}

func (d *GCS) findObjectByRegex() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "findObjectByRegex",
	})
	l.Debug("findObjectByRegex")
	if d.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}
	if d.KeyRegex == "" {
		return nil, fmt.Errorf("keyRegex is required")
	}
	var key string
	ctx := context.Background()
	q := &storage.Query{
		Delimiter: "/",
		Prefix:    d.KeyPrefix,
	}
	it := d.Client.Bucket(d.Bucket).Objects(ctx, q)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			l.Errorf("findObjectByRegex error: %s", err)
			return nil, err
		}
		if matched, err := regexp.MatchString(d.KeyRegex, attrs.Name); err != nil {
			l.Errorf("findObjectByRegex error: %s", err)
			return nil, err
		} else if matched {
			key = attrs.Name
			break
		}
	}
	if key == "" {
		return nil, nil
	}
	d.Key = key
	return d.getObject()
}

func (d *GCS) deleteObject() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "deleteObject",
	})
	l.Debug("deleteObject")
	if d.Bucket == "" {
		return fmt.Errorf("bucket is required")
	}
	if d.Key == "" {
		return fmt.Errorf("key is required")
	}
	ctx := context.Background()
	err := d.Client.Bucket(d.Bucket).Object(d.Key).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *GCS) mvObject(destBucket string, destKey string) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "mvObject",
	})
	l.Debug("mvObject")
	if destBucket == "" {
		return fmt.Errorf("destBucket is empty")
	}
	if destKey == "" {
		destKey = d.Key
	}
	ctx := context.Background()
	src := d.Client.Bucket(d.Bucket).Object(d.Key)
	dst := d.Client.Bucket(destBucket).Object(destKey)
	_, err := dst.CopierFrom(src).Run(ctx)
	if err != nil {
		return err
	}
	return d.deleteObject()
}

func (d *GCS) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "ClearWork",
	})
	l.Debug("ClearWork")
	if d.ClearOp == nil {
		return nil
	}
	if d.ClearOp.Operation == "" {
		return nil
	}
	switch d.ClearOp.Operation {
	case GCSOperationRM:
		return d.deleteObject()
	case GCSOperationMV:
		if d.ClearOp.Key == "" {
			d.ClearOp.Key = d.Key
		}
		opk := d.ClearOp.GetKey()
		return d.mvObject(d.ClearOp.Bucket, opk)
	default:
		return errors.New("unknown operation")
	}
}

func (d *GCS) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "HandleFailure",
	})
	l.Debug("HandleFailure")
	if d.FailOp == nil {
		return nil
	}
	if d.FailOp.Operation == "" {
		return nil
	}
	switch d.FailOp.Operation {
	case GCSOperationRM:
		return d.deleteObject()
	case GCSOperationMV:
		if d.FailOp.Key == "" {
			d.FailOp.Key = d.Key
		}
		opk := d.FailOp.GetKey()
		return d.mvObject(d.FailOp.Bucket, opk)
	default:
		return errors.New("unknown operation")
	}
}

func (d *GCS) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "Cleanup",
	})
	l.Debug("Cleanup")
	return nil
}
