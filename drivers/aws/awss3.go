package aws

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/google/uuid"
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

type S3 struct {
	Client    *s3.S3
	sts       *STSSession
	Bucket    string
	Key       string
	KeyRegex  string
	KeyPrefix string
	Region    string
	RoleARN   string
	ClearOp   *S3Op
	FailOp    *S3Op
}

func (d *S3) LogIdentity() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "LogIdentity",
	})
	l.Debug("LogIdentity")
	streq := &sts.GetCallerIdentityInput{}
	var sc *sts.STS
	if d.sts.Config != nil {
		sc = sts.New(d.sts.Session, d.sts.Config)
	} else {
		sc = sts.New(d.sts.Session)
	}
	r, err := sc.GetCallerIdentity(streq)
	if err != nil {
		l.Errorf("%+v", err)
	} else {
		l.Debugf("%+v", r)
	}
	return nil
}

func (o *S3Op) GetKey() string {
	if o.KeyTemplate == "" && o.Key != "" {
		return o.Key
	}
	return strings.ReplaceAll(o.KeyTemplate, "{{key}}", o.Key)
}

func (d *S3) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "LoadEnv",
	})
	l.Debug("LoadEnv")
	if os.Getenv(prefix+"AWS_REGION") != "" {
		d.Region = os.Getenv(prefix + "AWS_REGION")
	}
	if os.Getenv(prefix+"AWS_ROLE_ARN") != "" {
		d.RoleARN = os.Getenv(prefix + "AWS_ROLE_ARN")
	}
	if os.Getenv(prefix+"AWS_S3_BUCKET") != "" {
		d.Bucket = os.Getenv(prefix + "AWS_S3_BUCKET")
	}
	if os.Getenv(prefix+"AWS_S3_KEY") != "" {
		d.Key = os.Getenv(prefix + "AWS_S3_KEY")
	}
	if os.Getenv(prefix+"AWS_S3_KEY_REGEX") != "" {
		d.KeyRegex = os.Getenv(prefix + "AWS_S3_KEY_REGEX")
	}
	if os.Getenv(prefix+"AWS_S3_KEY_PREFIX") != "" {
		d.KeyPrefix = os.Getenv(prefix + "AWS_S3_KEY_PREFIX")
	}
	if os.Getenv(prefix+"AWS_S3_CLEAR_OP") != "" {
		d.ClearOp = &S3Op{}
		d.ClearOp.Operation = S3Operation(os.Getenv(prefix + "AWS_S3_CLEAR_OP"))
	}
	if os.Getenv(prefix+"AWS_S3_FAIL_OP") != "" {
		d.FailOp = &S3Op{}
		d.FailOp.Operation = S3Operation(os.Getenv(prefix + "AWS_S3_FAIL_OP"))
	}
	if os.Getenv(prefix+"AWS_S3_CLEAR_BUCKET") != "" {
		d.ClearOp.Bucket = os.Getenv(prefix + "AWS_S3_CLEAR_BUCKET")
	}
	if os.Getenv(prefix+"AWS_S3_CLEAR_KEY") != "" {
		d.ClearOp.Key = os.Getenv(prefix + "AWS_S3_CLEAR_KEY")
	}
	if os.Getenv(prefix+"AWS_S3_CLEAR_KEY_TEMPLATE") != "" {
		d.ClearOp.KeyTemplate = os.Getenv(prefix + "AWS_S3_CLEAR_KEY_TEMPLATE")
	}
	if os.Getenv(prefix+"AWS_S3_FAIL_BUCKET") != "" {
		d.FailOp.Bucket = os.Getenv(prefix + "AWS_S3_FAIL_BUCKET")
	}
	if os.Getenv(prefix+"AWS_S3_FAIL_KEY") != "" {
		d.FailOp.Key = os.Getenv(prefix + "AWS_S3_FAIL_KEY")
	}
	if os.Getenv(prefix+"AWS_S3_FAIL_KEY_TEMPLATE") != "" {
		d.FailOp.KeyTemplate = os.Getenv(prefix + "AWS_S3_FAIL_KEY_TEMPLATE")
	}
	if os.Getenv(prefix+"AWS_LOAD_CONFIG") != "" || os.Getenv("AWS_SDK_LOAD_CONFIG") != "" {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
	return nil
}

func (d *S3) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.Bucket = *flags.AWSS3Bucket
	d.Key = *flags.AWSS3Key
	d.Region = *flags.AWSRegion
	d.RoleARN = *flags.AWSRoleARN
	if flags.AWSS3KeyRegex != nil {
		d.KeyRegex = *flags.AWSS3KeyRegex
	}
	if flags.AWSS3KeyPrefix != nil {
		d.KeyPrefix = *flags.AWSS3KeyPrefix
	}
	if flags.AWSS3ClearOp != nil {
		d.ClearOp = &S3Op{}
		d.ClearOp.Operation = S3Operation(*flags.AWSS3ClearOp)
		d.ClearOp.Bucket = *flags.AWSS3ClearBucket
		d.ClearOp.Key = *flags.AWSS3ClearKey
		d.ClearOp.KeyTemplate = *flags.AWSS3ClearKeyTemplate
	}
	if flags.AWSS3FailOp != nil {
		d.FailOp = &S3Op{}
		d.FailOp.Operation = S3Operation(*flags.AWSS3FailOp)
		d.FailOp.Bucket = *flags.AWSS3FailBucket
		d.FailOp.Key = *flags.AWSS3FailKey
		d.FailOp.KeyTemplate = *flags.AWSS3FailKeyTemplate
	}
	if flags.AWSLoadConfig != nil && *flags.AWSLoadConfig {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
	return nil
}

func (d *S3) Init() error {
	l := log.WithFields(
		log.Fields{
			"pkg": "aws",
			"fn":  "CreateAWSSession",
		},
	)
	l.Debug("CreateAWSSession")
	if d.Region == "" {
		d.Region = os.Getenv("AWS_REGION")
	}
	if d.Region == "" {
		d.Region = "us-east-1"
	}
	cfg := &aws.Config{
		Region: aws.String(d.Region),
	}
	sess, err := session.NewSession(cfg)
	reqId := uuid.New().String()
	if d.RoleARN != "" {
		l.Debug("CreateAWSSession roleArn=%s requestId=%s", d.RoleARN, reqId)
		creds := stscreds.NewCredentials(sess, d.RoleARN, func(p *stscreds.AssumeRoleProvider) {
			p.RoleSessionName = "procx-" + reqId
		})
		cfg.Credentials = creds
	}
	if err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return err

	}
	d.Client = s3.New(sess, cfg)
	d.sts = &STSSession{
		Session: sess,
		Config:  cfg,
	}
	return err
}

func (d *S3) GetWork() (io.Reader, error) {
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

func (d *S3) getObject() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "getObject",
	})
	l.Debug("getObject")
	if d.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}
	if d.Key == "" {
		return nil, fmt.Errorf("key is required")
	}
	g := &s3.GetObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(d.Key),
	}
	resp, err := d.Client.GetObject(g)
	if err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return nil, err
	}
	return resp.Body, nil
}

func (d *S3) findObjectByPrefix() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "findObjectByPrefix",
	})
	l.Debug("findObjectByPrefix")
	if d.Bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}
	if d.KeyPrefix == "" {
		return nil, fmt.Errorf("key prefix is required")
	}
	g := &s3.ListObjectsInput{
		Bucket: aws.String(d.Bucket),
		Prefix: aws.String(d.KeyPrefix),
	}
	resp, err := d.Client.ListObjects(g)
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	if len(resp.Contents) == 0 {
		return nil, fmt.Errorf("no objects found")
	}
	ng := &s3.GetObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    resp.Contents[0].Key,
	}
	d.Key = *ng.Key
	ns, err := d.Client.GetObject(ng)
	if err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return nil, err
	}
	return ns.Body, nil
}

func (d *S3) findObjectByRegex() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
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
	var marker *string
FindKey:
	for key == "" {
		l.Debugf("findObjectByRegex bucket=%s keyRegex=%s marker=%s", d.Bucket, d.KeyRegex, marker)
		gl := &s3.ListObjectsInput{
			Bucket: aws.String(d.Bucket),
			Marker: marker,
		}
		resp, err := d.Client.ListObjects(gl)
		if err != nil {
			l.Errorf("%+v", err)
			return nil, err
		}
		if len(resp.Contents) == 0 {
			return nil, fmt.Errorf("no objects found")
		}
		for _, c := range resp.Contents {
			l.Debugf("findObjectByRegex bucket=%s keyRegex=%s marker=%s key=%s", d.Bucket, d.KeyRegex, marker, *c.Key)
			// check if the key matches the regex
			var ok bool
			var err error
			if ok, err = regexp.MatchString(d.KeyRegex, *c.Key); err != nil {
				l.Errorf("%+v", err)
				return nil, err
			}
			if ok {
				key = *c.Key
				break FindKey
			}
		}
		marker = resp.NextMarker
		if marker == nil {
			break
		}
	}
	if key == "" {
		return nil, fmt.Errorf("no objects found")
	}
	g := &s3.GetObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(key),
	}
	d.Key = key
	resp, err := d.Client.GetObject(g)
	if err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return nil, err
	}
	return resp.Body, nil
}

func (d *S3) deleteObject() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "deleteObject",
	})
	l.Debug("deleteObject")
	g := &s3.DeleteObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(d.Key),
	}
	_, err := d.Client.DeleteObject(g)
	if err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return err
	}
	return nil
}

func (d *S3) mvObject(destBucket string, destKey string) error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "mvObject",
	})
	l.Debug("mvObject")
	if destBucket == "" {
		return fmt.Errorf("destBucket is empty")
	}
	if destKey == "" {
		destKey = d.Key
	}
	g := &s3.CopyObjectInput{
		Bucket:     aws.String(destBucket),
		CopySource: aws.String(d.Bucket + "/" + d.Key),
		Key:        aws.String(destKey),
	}
	_, err := d.Client.CopyObject(g)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	if err := d.deleteObject(); err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return err
	}
	return nil
}

func (d *S3) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "ClearWork",
	})
	l.Debug("ClearWork")
	if d.ClearOp == nil {
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

func (d *S3) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "HandleFailure",
	})
	l.Debug("HandleFailure")
	if d.FailOp == nil {
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
		return errors.New("unknown operation")
	}
}

func (d *S3) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "Cleanup",
	})
	l.Debug("Cleanup")
	return nil
}
