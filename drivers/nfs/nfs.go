package nfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/go-nfs-client/nfs"
	"github.com/vmware/go-nfs-client/nfs/rpc"
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

type NFSMount struct {
	Mount  *nfs.Mount
	Target *nfs.Target
}

type NFS struct {
	Host      string
	Target    string
	Folder    string
	Key       string
	KeyPrefix string
	KeyRegex  string
	MountPath string
	ClearOp   *S3Op
	FailOp    *S3Op
	Client    *NFSMount
}

func (o *S3Op) GetKey() string {
	if o.KeyTemplate == "" && o.Key != "" {
		return o.Key
	}
	return strings.ReplaceAll(o.KeyTemplate, "{{key}}", o.Key)
}

func (d *NFS) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "LoadEnv",
	})
	l.Debug("LoadEnv")
	if os.Getenv(prefix+"NFS_HOST") != "" {
		d.Host = os.Getenv(prefix + "NFS_HOST")
	}
	if os.Getenv(prefix+"NFS_KEY") != "" {
		d.Key = os.Getenv(prefix + "NFS_KEY")
	}
	if os.Getenv(prefix+"NFS_KEY_PREFIX") != "" {
		d.KeyPrefix = os.Getenv(prefix + "NFS_KEY_PREFIX")
	}
	if os.Getenv(prefix+"NFS_KEY_REGEX") != "" {
		d.KeyRegex = os.Getenv(prefix + "NFS_KEY_REGEX")
	}
	if os.Getenv(prefix+"NFS_FOLDER") != "" {
		d.Folder = os.Getenv(prefix + "NFS_FOLDER")
	}
	if os.Getenv(prefix+"NFS_TARGET") != "" {
		d.Target = os.Getenv(prefix + "NFS_TARGET")
	}
	if os.Getenv(prefix+"NFS_MOUNT_PATH") != "" {
		d.MountPath = os.Getenv(prefix + "NFS_MOUNT_PATH")
	}
	if os.Getenv(prefix+"NFS_CLEAR_OP") != "" {
		d.ClearOp = &S3Op{}
		d.ClearOp.Operation = S3Operation(os.Getenv(prefix + "NFS_CLEAR_OP"))
	}
	if os.Getenv(prefix+"NFS_FAIL_OP") != "" {
		d.FailOp = &S3Op{}
		d.FailOp.Operation = S3Operation(os.Getenv(prefix + "NFS_FAIL_OP"))
	}
	if os.Getenv(prefix+"NFS_CLEAR_FOLDER") != "" {
		d.ClearOp.Bucket = os.Getenv(prefix + "NFS_CLEAR_FOLDER")
	}
	if os.Getenv(prefix+"NFS_CLEAR_KEY") != "" {
		d.ClearOp.Key = os.Getenv(prefix + "NFS_CLEAR_KEY")
	}
	if os.Getenv(prefix+"NFS_CLEAR_KEY_TEMPLATE") != "" {
		d.ClearOp.KeyTemplate = os.Getenv(prefix + "NFS_CLEAR_KEY_TEMPLATE")
	}
	if os.Getenv(prefix+"NFS_FAIL_FOLDER") != "" {
		d.FailOp.Bucket = os.Getenv(prefix + "NFS_FAIL_FOLDER")
	}
	if os.Getenv(prefix+"NFS_FAIL_KEY") != "" {
		d.FailOp.Key = os.Getenv(prefix + "NFS_FAIL_KEY")
	}
	if os.Getenv(prefix+"NFS_FAIL_KEY_TEMPLATE") != "" {
		d.FailOp.KeyTemplate = os.Getenv(prefix + "NFS_FAIL_KEY_TEMPLATE")
	}
	return nil
}

func (d *NFS) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.Host = *flags.NFSHost
	d.Target = *flags.NFSTarget
	d.Folder = *flags.NFSFolder
	d.Key = *flags.NFSKey
	d.MountPath = *flags.NFSMountPath
	if flags.NFSKeyRegex != nil {
		d.KeyRegex = *flags.NFSKeyRegex
	}
	if flags.NFSKeyPrefix != nil {
		d.KeyPrefix = *flags.NFSKeyPrefix
	}
	if flags.NFSClearOp != nil {
		d.ClearOp = &S3Op{}
		d.ClearOp.Operation = S3Operation(*flags.NFSClearOp)
		d.ClearOp.Bucket = *flags.NFSClearFolder
		d.ClearOp.Key = *flags.NFSClearKey
		d.ClearOp.KeyTemplate = *flags.NFSClearKeyTemplate
	}
	if flags.NFSFailOp != nil {
		d.FailOp = &S3Op{}
		d.FailOp.Operation = S3Operation(*flags.NFSFailOp)
		d.FailOp.Bucket = *flags.NFSFailFolder
		d.FailOp.Key = *flags.NFSFailKey
		d.FailOp.KeyTemplate = *flags.NFSFailKeyTemplate
	}
	return nil
}

func (d *NFS) Init() error {
	l := log.WithFields(
		log.Fields{
			"pkg": "aws",
			"fn":  "CreateNFSSession",
		},
	)
	l.Debug("CreateNFSSession")
	mount, err := nfs.DialMount(d.Host)
	if err != nil {
		log.Fatalf("unable to dial MOUNT service: %v", err)
	}
	auth := rpc.NewAuthUnix("hasselhoff", 1001, 1001)
	v, err := mount.Mount(d.Target, auth.Auth())
	if err != nil {
		log.Fatalf("unable to mount volume: %v", err)
	}
	d.Client = &NFSMount{
		Mount:  mount,
		Target: v,
	}
	return err
}

func (d *NFS) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
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

func (d *NFS) getObject() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg":    "aws",
		"fn":     "getObject",
		"folder": d.Folder,
		"key":    d.Key,
	})
	l.Debug("getObject")
	if d.Key == "" {
		return nil, errors.New("no key")
	}
	return d.Client.Target.Open(path.Join(d.Folder, d.Key))
}

func (t *NFSMount) FindPrefixRecursive(folder, prefix string) *string {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "FindPrefixRecursive",
	})
	l.Debugf("FindPrefixRecursive folder=%s prefix=%s", folder, prefix)
	entries, err := t.Target.ReadDirPlus(folder)
	if err != nil {
		l.Errorf("FindPrefixRecursive error=%v", err)
		return nil
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "." || entry.Name() == ".." {
				continue
			}
			if t.FindPrefixRecursive(path.Join(folder, entry.Name()), prefix) != nil {
				n := entry.Name()
				return &n
			}
		} else {
			if strings.HasPrefix(entry.Name(), prefix) {
				n := entry.Name()
				return &n
			}
		}
	}
	return nil
}

func (d *NFS) findObjectByPrefix() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "findObjectByPrefix",
	})
	l.Debug("findObjectByPrefix")
	if d.KeyPrefix == "" {
		return nil, fmt.Errorf("key prefix is required")
	}
	key := d.Client.FindPrefixRecursive(d.Folder, d.KeyPrefix)
	if key == nil {
		return nil, nil
	}
	l.Debugf("findObjectByPrefix key=%s", *key)
	d.Key = *key
	return d.getObject()
}

func (t *NFSMount) FindRegexRecursive(folder, prefix string) *string {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "FindRegexRecursive",
	})
	l.Debugf("FindPrefixRecursive folder=%s prefix=%s", folder, prefix)
	entries, err := t.Target.ReadDirPlus(folder)
	if err != nil {
		l.Errorf("FindPrefixRecursive error=%v", err)
		return nil
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "." || entry.Name() == ".." {
				continue
			}
			if t.FindPrefixRecursive(path.Join(folder, entry.Name()), prefix) != nil {
				n := entry.Name()
				return &n
			}
		} else {
			ok, err := regexp.MatchString(prefix, entry.Name())
			if err != nil {
				l.Errorf("FindRegexRecursive error=%v", err)
				return nil
			}
			if ok {
				n := entry.Name()
				return &n
			}
		}
	}
	return nil
}

func (d *NFS) findObjectByRegex() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "findObjectByRegex",
	})
	l.Debug("findObjectByRegex")
	if d.KeyRegex == "" {
		return nil, fmt.Errorf("key regex is required")
	}
	key := d.Client.FindRegexRecursive(d.Folder, d.KeyRegex)
	if key == nil {
		return nil, nil
	}
	l.Debugf("findObjectByRegex key=%s", *key)
	d.Key = *key
	return d.getObject()
}

func (d *NFS) deleteObject() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "deleteObject",
	})
	l.Debug("deleteObject")
	if d.Key == "" {
		return errors.New("no key")
	}
	return d.Client.Target.Remove(d.Key)
}

func (d *NFS) mvObject(destBucket string, destKey string) error {
	l := log.WithFields(log.Fields{
		"pkg":        "aws",
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
	_, _, err := d.Client.Target.Lookup(destBucket)
	if err != nil {
		// Create the bucket if it doesn't exist
		_, err = d.Client.Target.Mkdir(destBucket, 0755)
		if err != nil {
			l.Errorf("mvObject error=%v", err)
			return err
		}
	}
	w, err := d.Client.Target.OpenFile(path.Join(destBucket, destKey), 0644)
	if err != nil {
		return err
	}
	defer w.Close()
	r, err := d.getObject()
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	return d.Client.Target.Remove(path.Join(d.Folder, d.Key))
}

func (d *NFS) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
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

func (d *NFS) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
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

func (d *NFS) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "Cleanup",
	})
	l.Debug("Cleanup")
	if err := d.Client.Mount.Unmount(); err != nil {
		l.Errorf("%+v", err)
		return err
	}
	d.Client.Mount.Close()
	return nil
}
