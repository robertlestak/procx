package smb

import (
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/hirochachacha/go-smb2"
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
	Key         string
	KeyTemplate string
}

type SMBClient struct {
	Client *smb2.Session
	Conn   net.Conn
	Share  *smb2.Share
}

type SMB struct {
	Host     string
	Port     int
	Username *string
	Password *string
	Share    *string
	Key      string
	KeyGlob  *string
	ClearOp  *S3Op
	FailOp   *S3Op
	Client   *SMBClient
	file     *smb2.File
}

func (o *S3Op) GetKey() string {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "GetKey",
		"key": o.Key,
	})
	l.Debug("GetKey")
	if o.KeyTemplate == "" && o.Key != "" {
		return o.Key
	}
	// get base key
	sp := strings.Split(o.Key, "\\")
	baseKey := sp[len(sp)-1]
	return strings.ReplaceAll(o.KeyTemplate, "{{key}}", baseKey)
}

func (d *SMB) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "LoadEnv",
	})
	l.Debug("LoadEnv")
	if os.Getenv(prefix+"SMB_HOST") != "" {
		d.Host = os.Getenv(prefix + "SMB_HOST")
	}
	if os.Getenv(prefix+"SMB_PORT") != "" {
		pv, err := strconv.Atoi(os.Getenv(prefix + "SMB_PORT"))
		if err != nil {
			l.Errorf("%+v", err)
			return err
		}
		d.Port = pv
	}
	if os.Getenv(prefix+"SMB_USER") != "" {
		v := os.Getenv(prefix + "SMB_USER")
		d.Username = &v
	}
	if os.Getenv(prefix+"SMB_PASS") != "" {
		v := os.Getenv(prefix + "SMB_PASS")
		d.Password = &v
	}
	if os.Getenv(prefix+"SMB_SHARE") != "" {
		v := os.Getenv(prefix + "SMB_SHARE")
		d.Share = &v
	}
	if os.Getenv(prefix+"SMB_KEY") != "" {
		d.Key = os.Getenv(prefix + "SMB_KEY")
	}
	if os.Getenv(prefix+"SMB_KEY_GLOB") != "" {
		v := os.Getenv(prefix + "SMB_KEY_GLOB")
		d.KeyGlob = &v
	}
	if d.ClearOp == nil {
		d.ClearOp = &S3Op{}
	}
	if d.FailOp == nil {
		d.FailOp = &S3Op{}
	}
	if os.Getenv(prefix+"SMB_CLEAR_OP") != "" {
		d.ClearOp.Operation = S3Operation(os.Getenv(prefix + "SMB_CLEAR_OP"))
	}
	if os.Getenv(prefix+"SMB_CLEAR_KEY") != "" {
		d.ClearOp.Key = os.Getenv(prefix + "SMB_CLEAR_KEY")
	}
	if os.Getenv(prefix+"SMB_CLEAR_KEY_TEMPLATE") != "" {
		d.ClearOp.KeyTemplate = os.Getenv(prefix + "SMB_CLEAR_KEY_TEMPLATE")
	}
	if os.Getenv(prefix+"SMB_FAIL_OP") != "" {
		d.FailOp.Operation = S3Operation(os.Getenv(prefix + "SMB_FAIL_OP"))
	}
	if os.Getenv(prefix+"SMB_FAIL_KEY") != "" {
		d.FailOp.Key = os.Getenv(prefix + "SMB_FAIL_KEY")
	}
	if os.Getenv(prefix+"SMB_FAIL_KEY_TEMPLATE") != "" {
		d.FailOp.KeyTemplate = os.Getenv(prefix + "SMB_FAIL_KEY_TEMPLATE")
	}
	return nil
}

func (d *SMB) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.Host = *flags.SMBHost
	d.Port = *flags.SMBPort
	d.Username = flags.SMBUser
	d.Password = flags.SMBPass
	d.Key = *flags.SMBKey
	d.Share = flags.SMBShare
	if d.ClearOp == nil {
		d.ClearOp = &S3Op{}
	}
	if d.FailOp == nil {
		d.FailOp = &S3Op{}
	}
	d.KeyGlob = flags.SMBKeyGlob
	d.ClearOp.Operation = S3Operation(*flags.SMBClearOp)
	d.ClearOp.Key = *flags.SMBClearKey
	d.ClearOp.KeyTemplate = *flags.SMBClearKeyTemplate
	d.FailOp.Operation = S3Operation(*flags.SMBFailOp)
	d.FailOp.Key = *flags.SMBFailKey
	d.FailOp.KeyTemplate = *flags.SMBFailKeyTemplate
	return nil
}

func (d *SMB) Init() error {
	l := log.WithFields(
		log.Fields{
			"pkg": "nfs",
			"fn":  "CreateNFSSession",
		},
	)
	l.Debug("CreateNFSSession")
	if d.Host == "" || d.Port == 0 || d.Username == nil || d.Password == nil || d.Share == nil {
		return errors.New("invalid SMB configuration")
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", d.Host, d.Port))
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	sd := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     *d.Username,
			Password: *d.Password,
		},
	}
	s, err := sd.Dial(conn)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	m, err := s.Mount(*d.Share)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	c := &SMBClient{
		Client: s,
		Conn:   conn,
		Share:  m,
	}
	d.Client = c
	return err
}

func (d *SMB) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "GetWork",
	})
	l.Debug("GetWork")
	if d.Key != "" {
		l.Debugf("GetWork key=%s", d.Key)
		return d.getObject()
	} else if *d.KeyGlob != "" {
		l.Debugf("GetWork keyGlob=%s", *d.KeyGlob)
		_, err := d.findObjectGlob()
		if err != nil {
			l.Errorf("%+v", err)
			return nil, err
		}
		return d.getObject()
	} else {
		l.Debug("GetWork no key")
		return nil, errors.New("no key")
	}
}

func (d *SMB) findObjectGlob() (string, error) {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "findObjectGlob",
	})
	l.Debug("findObjectGlob")
	if d.KeyGlob == nil || *d.KeyGlob == "" {
		l.Error("findObjectGlob no keyGlob")
		return "", errors.New("no keyGlob")
	}
	l.Debugf("findObjectGlob keyGlob=%s", *d.KeyGlob)
	matches, err := iofs.Glob(d.Client.Share.DirFS("."), *d.KeyGlob)
	if err != nil {
		l.Errorf("%+v", err)
		return "", err
	}
	if len(matches) == 0 {
		l.Error("findObjectGlob no matches")
		return "", errors.New("no matches")
	}
	l.Debugf("findObjectGlob matches=%d", len(matches))
	d.Key = matches[0]
	return d.Key, nil
}

func (d *SMB) getObject() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "getObject",
		"key": d.Key,
	})
	l.Debug("getObject")
	if d.Key == "" {
		return nil, errors.New("no key")
	}
	if d.Client == nil {
		return nil, errors.New("no client")
	}
	f, err := d.Client.Share.Open(d.Key)
	if err != nil {
		l.Errorf("unable to open file: %v", err)
		return nil, err
	}
	d.file = f
	return f, nil
}

func (d *SMB) deleteObject() error {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "deleteObject",
	})
	l.Debug("deleteObject")
	if d.Key == "" {
		return errors.New("no key")
	}
	return d.Client.Share.Remove(d.Key)
}

func (d *SMB) mvObject(destKey string) error {
	l := log.WithFields(log.Fields{
		"pkg":     "nfs",
		"fn":      "mvObject",
		"destKey": destKey,
	})
	l.Debug("mvObject")
	if destKey == "" {
		destKey = d.Key
	}
	if d.Key == "" {
		return errors.New("no key")
	}
	if err := d.Client.Share.Rename(d.Key, destKey); err != nil {
		l.Errorf("unable to rename file: %v", err)
		return err
	}
	return nil
}

func (d *SMB) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "ClearWork",
	})
	l.Debug("ClearWork")
	d.file.Close()
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
		return d.mvObject(opk)
	default:
		return errors.New("unknown operation")
	}
}

func (d *SMB) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "HandleFailure",
	})
	l.Debug("HandleFailure")
	d.file.Close()
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
		return d.mvObject(opk)
	default:
		return nil
	}
}

func (d *SMB) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "nfs",
		"fn":  "Cleanup",
	})
	l.Debug("Cleanup")
	if d.Client == nil {
		return nil
	}
	d.Client.Share.Umount()
	d.Client.Client.Logoff()
	d.Client.Conn.Close()
	return nil
}
