package fs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type S3Operation string

var (
	S3OperationRM = S3Operation("rm")
	S3OperationMV = S3Operation("mv")
)

type S3Op struct {
	Operation   S3Operation
	Bucket      string
	Key         string
	KeyTemplate string
}

type FS struct {
	Folder    string
	Key       string
	KeyPrefix string
	KeyRegex  string
	ClearOp   *S3Op
	FailOp    *S3Op
}

func (o *S3Op) GetKey() string {
	if o.KeyTemplate == "" && o.Key != "" {
		return o.Key
	}
	return strings.ReplaceAll(o.KeyTemplate, "{{key}}", o.Key)
}

func (d *FS) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
		"fn":  "LoadEnv",
	})
	l.Debug("LoadEnv")
	if os.Getenv(prefix+"FS_KEY") != "" {
		d.Key = os.Getenv(prefix + "FS_KEY")
	}
	if os.Getenv(prefix+"FS_KEY_PREFIX") != "" {
		d.KeyPrefix = os.Getenv(prefix + "FS_KEY_PREFIX")
	}
	if os.Getenv(prefix+"FS_KEY_REGEX") != "" {
		d.KeyRegex = os.Getenv(prefix + "FS_KEY_REGEX")
	}
	if os.Getenv(prefix+"FS_FOLDER") != "" {
		d.Folder = os.Getenv(prefix + "FS_FOLDER")
	}
	if os.Getenv(prefix+"FS_CLEAR_OP") != "" {
		d.ClearOp = &S3Op{}
		d.ClearOp.Operation = S3Operation(os.Getenv(prefix + "FS_CLEAR_OP"))
	}
	if os.Getenv(prefix+"FS_FAIL_OP") != "" {
		d.FailOp = &S3Op{}
		d.FailOp.Operation = S3Operation(os.Getenv(prefix + "FS_FAIL_OP"))
	}
	if os.Getenv(prefix+"FS_CLEAR_FOLDER") != "" {
		d.ClearOp.Bucket = os.Getenv(prefix + "FS_CLEAR_FOLDER")
	}
	if os.Getenv(prefix+"FS_CLEAR_KEY") != "" {
		d.ClearOp.Key = os.Getenv(prefix + "FS_CLEAR_KEY")
	}
	if os.Getenv(prefix+"FS_CLEAR_KEY_TEMPLATE") != "" {
		d.ClearOp.KeyTemplate = os.Getenv(prefix + "FS_CLEAR_KEY_TEMPLATE")
	}
	if os.Getenv(prefix+"FS_FAIL_FOLDER") != "" {
		d.FailOp.Bucket = os.Getenv(prefix + "FS_FAIL_FOLDER")
	}
	if os.Getenv(prefix+"FS_FAIL_KEY") != "" {
		d.FailOp.Key = os.Getenv(prefix + "FS_FAIL_KEY")
	}
	if os.Getenv(prefix+"FS_FAIL_KEY_TEMPLATE") != "" {
		d.FailOp.KeyTemplate = os.Getenv(prefix + "FS_FAIL_KEY_TEMPLATE")
	}
	return nil
}

func (d *FS) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.Folder = *flags.FSFolder
	d.Key = *flags.FSKey
	if flags.FSKeyRegex != nil {
		d.KeyRegex = *flags.FSKeyRegex
	}
	if flags.FSKeyPrefix != nil {
		d.KeyPrefix = *flags.FSKeyPrefix
	}
	if flags.FSClearOp != nil {
		d.ClearOp = &S3Op{}
		d.ClearOp.Operation = S3Operation(*flags.FSClearOp)
		d.ClearOp.Bucket = *flags.FSClearFolder
		d.ClearOp.Key = *flags.FSClearKey
		d.ClearOp.KeyTemplate = *flags.FSClearKeyTemplate
	}
	if flags.FSFailOp != nil {
		d.FailOp = &S3Op{}
		d.FailOp.Operation = S3Operation(*flags.FSFailOp)
		d.FailOp.Bucket = *flags.FSFailFolder
		d.FailOp.Key = *flags.FSFailKey
		d.FailOp.KeyTemplate = *flags.FSFailKeyTemplate
	}
	return nil
}

func (d *FS) Init() error {
	l := log.WithFields(
		log.Fields{
			"pkg": "fs",
			"fn":  "CreateFSSession",
		},
	)
	l.Debug("CreateFSSession")
	return nil
}

func (d *FS) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
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

func (d *FS) getObject() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg":    "fs",
		"fn":     "getObject",
		"folder": d.Folder,
		"key":    d.Key,
	})
	l.Debug("getObject")
	if d.Key == "" {
		return nil, errors.New("no key")
	}
	return os.Open(path.Join(d.Folder, d.Key))
}

func (t *FS) FindPrefixRecursive(folder, prefix string) *string {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
		"fn":  "FindPrefixRecursive",
	})
	l.Debugf("FindPrefixRecursive folder=%s prefix=%s", folder, prefix)
	var found *string
	err := filepath.WalkDir(folder,
		func(path string, d fs.DirEntry, err error) error {
			fn := strings.Replace(path, folder+"/", "", 1)
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasPrefix(fn, prefix) {
				l.Debugf("FindPrefixRecursive found %s", path)
				found = &fn
				return filepath.SkipDir
			}
			return nil
		})
	if err != nil && err != filepath.SkipDir {
		l.Errorf("FindPrefixRecursive error %s", err)
		return nil
	}
	return found
}

func (d *FS) findObjectByPrefix() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
		"fn":  "findObjectByPrefix",
	})
	l.Debug("findObjectByPrefix")
	if d.KeyPrefix == "" {
		return nil, fmt.Errorf("key prefix is required")
	}
	key := d.FindPrefixRecursive(d.Folder, d.KeyPrefix)
	if key == nil {
		return nil, nil
	}
	l.Debugf("findObjectByPrefix key=%s", *key)
	d.Key = *key
	return d.getObject()
}

func (t *FS) FindRegexRecursive(folder, prefix string) *string {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
		"fn":  "FindRegexRecursive",
	})
	l.Debugf("FindPrefixRecursive folder=%s prefix=%s", folder, prefix)

	var found *string
	err := filepath.WalkDir(folder,
		func(path string, d fs.DirEntry, err error) error {
			fn := strings.Replace(path, folder+"/", "", 1)
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ok, err := regexp.MatchString(prefix, fn)
			if err != nil {
				l.Errorf("FindRegexRecursive error=%v", err)
				return nil
			}
			if ok {
				l.Debugf("FindRegexRecursive found %s", path)
				found = &fn
				return filepath.SkipDir
			}
			return nil
		})
	if err != nil && err != filepath.SkipDir {
		l.Errorf("FindPrefixRecursive error %s", err)
		return nil
	}
	return found
}

func (d *FS) findObjectByRegex() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
		"fn":  "findObjectByRegex",
	})
	l.Debug("findObjectByRegex")
	if d.KeyRegex == "" {
		return nil, fmt.Errorf("key regex is required")
	}
	key := d.FindRegexRecursive(d.Folder, d.KeyRegex)
	if key == nil {
		return nil, nil
	}
	l.Debugf("findObjectByRegex key=%s", *key)
	d.Key = *key
	return d.getObject()
}

func (d *FS) deleteObject() error {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
		"fn":  "deleteObject",
	})
	l.Debug("deleteObject")
	if d.Key == "" {
		return errors.New("no key")
	}
	return os.Remove(path.Join(d.Folder, d.Key))
}

func (d *FS) mvObject(destBucket string, destKey string) error {
	l := log.WithFields(log.Fields{
		"pkg":        "fs",
		"fn":         "mvObject",
		"destBucket": destBucket,
		"destKey":    destKey,
	})
	l.Debug("mvObject")
	if destBucket == "" {
		return fmt.Errorf("destBucket is empty")
	}
	if destKey == "" {
		destKey = d.Key
	}
	if d.Key == "" {
		return errors.New("no key")
	}
	of := path.Join(d.Folder, d.Key)
	nf := path.Join(destBucket, destKey)
	l.Debugf("mvObject of=%s nf=%s", of, nf)
	if err := os.Rename(of, nf); err != nil {
		l.Errorf("mvObject error=%v", err)
		return err
	}
	return nil
}

func (d *FS) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
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
	case S3OperationRM:
		return d.deleteObject()
	case S3OperationMV:
		if d.ClearOp.Key == "" {
			d.ClearOp.Key = d.Key
		}
		opk := d.ClearOp.GetKey()
		return d.mvObject(d.ClearOp.Bucket, opk)
	default:
		return errors.New("unknown operation")
	}
}

func (d *FS) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
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
	case S3OperationRM:
		return d.deleteObject()
	case S3OperationMV:
		if d.FailOp.Key == "" {
			d.FailOp.Key = d.Key
		}
		opk := d.FailOp.GetKey()
		return d.mvObject(d.FailOp.Bucket, opk)
	default:
		return nil
	}
}

func (d *FS) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
		"fn":  "Cleanup",
	})
	l.Debug("Cleanup")
	return nil
}
