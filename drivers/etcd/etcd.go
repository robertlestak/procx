package etcd

import (
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/schema"
	"github.com/robertlestak/procx/pkg/utils"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Operation string

var (
	OperationRM  = Operation("rm")
	OperationPut = Operation("put")
	OperationMV  = Operation("mv")
)

type Etcd struct {
	Client     *clientv3.Client
	Hosts      []string
	Username   *string
	Password   *string
	Key        string
	WithPrefix *bool
	ClearOp    *Operation
	ClearKey   *string
	ClearVal   *string
	FailOp     *Operation
	FailKey    *string
	FailVal    *string
	// TLS
	EnableTLS   *bool
	TLSInsecure *bool
	TLSCert     *string
	TLSKey      *string
	TLSCA       *string
	data        []byte
}

func (d *Etcd) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "etcd",
		"fn":  "LoadEnv",
	})
	l.Debug("LoadEnv")
	if os.Getenv(prefix+"ETCD_HOSTS") != "" {
		d.Hosts = strings.Split(os.Getenv(prefix+"ETCD_HOSTS"), ",")
	}
	if os.Getenv(prefix+"ETCD_USERNAME") != "" {
		v := os.Getenv(prefix + "ETCD_USERNAME")
		d.Username = &v
	}
	if os.Getenv(prefix+"ETCD_PASSWORD") != "" {
		v := os.Getenv(prefix + "ETCD_PASSWORD")
		d.Password = &v
	}
	if os.Getenv(prefix+"ETCD_RETRIEVE_KEY") != "" {
		d.Key = os.Getenv(prefix + "ETCD_RETRIEVE_KEY")
	}
	if os.Getenv(prefix+"ETCD_CLEAR_OP") != "" {
		v := Operation(os.Getenv(prefix + "ETCD_CLEAR_OP"))
		d.ClearOp = &v
	}
	if os.Getenv(prefix+"ETCD_CLEAR_KEY") != "" {
		v := os.Getenv(prefix + "ETCD_CLEAR_KEY")
		d.ClearKey = &v
	}
	if os.Getenv(prefix+"ETCD_CLEAR_VAL") != "" {
		v := os.Getenv(prefix + "ETCD_CLEAR_VAL")
		d.ClearVal = &v
	}
	if os.Getenv(prefix+"ETCD_FAIL_OP") != "" {
		v := Operation(os.Getenv(prefix + "ETCD_FAIL_OP"))
		d.FailOp = &v
	}
	if os.Getenv(prefix+"ETCD_FAIL_KEY") != "" {
		v := os.Getenv(prefix + "ETCD_FAIL_KEY")
		d.FailKey = &v
	}
	if os.Getenv(prefix+"ETCD_FAIL_VAL") != "" {
		v := os.Getenv(prefix + "ETCD_FAIL_VAL")
		d.FailVal = &v
	}
	if os.Getenv(prefix+"ETCD_WITH_PREFIX") != "" {
		v, err := strconv.ParseBool(os.Getenv(prefix + "ETCD_WITH_PREFIX"))
		if err != nil {
			l.Error(err)
			return err
		}
		d.WithPrefix = &v
	}
	if os.Getenv(prefix+"ETCD_TLS_ENABLE") != "" {
		v, err := strconv.ParseBool(os.Getenv(prefix + "ETCD_TLS_ENABLE"))
		if err != nil {
			l.Error(err)
			return err
		}
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"ETCD_TLS_INSECURE") != "" {
		v, err := strconv.ParseBool(os.Getenv(prefix + "ETCD_TLS_INSECURE"))
		if err != nil {
			l.Error(err)
			return err
		}
		d.TLSInsecure = &v
	}
	if os.Getenv(prefix+"ETCD_TLS_CERT") != "" {
		v := os.Getenv(prefix + "ETCD_TLS_CERT")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"ETCD_TLS_KEY") != "" {
		v := os.Getenv(prefix + "ETCD_TLS_KEY")
		d.TLSKey = &v
	}
	if os.Getenv(prefix+"ETCD_TLS_CA") != "" {
		v := os.Getenv(prefix + "ETCD_TLS_CA")
		d.TLSCA = &v
	}
	return nil
}

func (d *Etcd) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "etcd",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.Username = flags.EtcdUsername
	d.Password = flags.EtcdPassword
	d.Hosts = strings.Split(*flags.EtcdHosts, ",")
	d.Key = *flags.EtcdKey
	d.WithPrefix = flags.EtcdWithPrefix
	op := Operation(*flags.EtcdClearOp)
	d.ClearOp = &op
	d.ClearKey = flags.EtcdClearKey
	d.ClearVal = flags.EtcdClearVal
	fop := Operation(*flags.EtcdFailOp)
	d.FailOp = &fop
	d.FailKey = flags.EtcdFailKey
	d.FailVal = flags.EtcdFailVal
	d.EnableTLS = flags.EtcdTLSEnable
	d.TLSInsecure = flags.EtcdTLSInsecure
	d.TLSCert = flags.EtcdTLSCert
	d.TLSKey = flags.EtcdTLSKey
	d.TLSCA = flags.EtcdTLSCA
	return nil
}

func (d *Etcd) Init() error {
	l := log.WithFields(
		log.Fields{
			"pkg": "etcd",
			"fn":  "CreateFSSession",
		},
	)
	l.Debug("CreateFSSession")
	cfg := clientv3.Config{
		Endpoints:   d.Hosts,
		DialTimeout: 5 * time.Second,
	}
	if d.Username != nil && *d.Username != "" && d.Password != nil && *d.Password != "" {
		cfg.Username = *d.Username
		cfg.Password = *d.Password
	}
	if d.EnableTLS != nil && *d.EnableTLS {
		t, err := utils.TlsConfig(d.EnableTLS, d.TLSInsecure, d.TLSCA, d.TLSCert, d.TLSKey)
		if err != nil {
			l.Errorf("%+v", err)
			return err
		}
		cfg.TLS = t
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		l.Error(err)
		return err
	}
	d.Client = cli
	return nil
}

func (d *Etcd) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "fs",
		"fn":  "GetWork",
	})
	l.Debug("GetWork")
	var opts []clientv3.OpOption
	if d.WithPrefix != nil && *d.WithPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}
	resp, err := d.Client.Get(d.Client.Ctx(), d.Key, opts...)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	d.data = resp.Kvs[0].Value
	return strings.NewReader(string(d.data)), nil
}

func (d *Etcd) rmKey(key string) error {
	l := log.WithFields(log.Fields{
		"pkg": "etcd",
		"fn":  "rmKey",
	})
	l.Debug("rmKey")
	_, err := d.Client.Delete(d.Client.Ctx(), key)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *Etcd) put(key, value string) error {
	l := log.WithFields(log.Fields{
		"pkg": "etcd",
		"fn":  "PutWork",
	})
	l.Debug("PutWork")
	_, err := d.Client.Put(d.Client.Ctx(), key, value)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *Etcd) move(destKey string) error {
	l := log.WithFields(log.Fields{
		"pkg": "etcd",
		"fn":  "move",
	})
	l.Debug("move")
	if d.data == nil {
		return nil
	}
	if err := d.put(destKey, string(d.data)); err != nil {
		l.Error(err)
		return err
	}
	if err := d.rmKey(d.Key); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *Etcd) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "etcd",
		"fn":  "ClearWork",
		"op":  *d.ClearOp,
	})
	l.Debug("Clearing work from etcd")
	if d.ClearOp == nil || *d.ClearOp == "" {
		return nil
	}
	if d.ClearKey == nil || *d.ClearKey == "" {
		d.ClearKey = &d.Key
	}
	if d.ClearVal != nil && *d.ClearVal != "" {
		v := schema.ReplaceParamsString(d.data, *d.ClearVal)
		d.ClearVal = &v
	}
	switch *d.ClearOp {
	case OperationRM:
		return d.rmKey(*d.ClearKey)
	case OperationMV:
		return d.move(*d.ClearKey)
	case OperationPut:
		return d.put(*d.ClearKey, *d.ClearVal)
	default:
		return nil
	}
}

func (d *Etcd) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "etcd",
		"fn":  "HandleFailure",
		"op":  *d.FailOp,
	})
	l.Debug("Handling failure")
	if d.FailOp == nil || *d.FailOp == "" {
		return nil
	}
	if d.FailKey == nil || *d.FailKey == "" {
		d.FailKey = &d.Key
	}
	if d.FailVal != nil && *d.FailVal != "" {
		v := schema.ReplaceParamsString(d.data, *d.FailVal)
		d.FailVal = &v
	}
	switch *d.FailOp {
	case OperationRM:
		return d.rmKey(*d.FailKey)
	case OperationMV:
		return d.move(*d.FailKey)
	case OperationPut:
		return d.put(*d.FailKey, *d.FailVal)
	default:
		return nil
	}
}

func (d *Etcd) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "etcd",
		"fn":  "Cleanup",
	})
	l.Debug("Cleanup")
	if err := d.Client.Close(); err != nil {
		l.Error(err)
		return err
	}
	return nil
}
