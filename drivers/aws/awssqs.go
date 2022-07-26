package aws

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/google/uuid"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
)

type STSSession struct {
	Session *session.Session
	Config  *aws.Config
}

type SQS struct {
	Client        *sqs.SQS
	sts           *STSSession
	Queue         string
	ReceiptHandle string
	Region        string
	RoleARN       string
	IncludeID     bool
}

func (d *SQS) LoadEnv(prefix string) error {
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
	if os.Getenv(prefix+"AWS_SQS_QUEUE_URL") != "" {
		d.Queue = os.Getenv(prefix + "AWS_SQS_QUEUE_URL")
	}
	if os.Getenv(prefix+"AWS_LOAD_CONFIG") != "" || os.Getenv("AWS_SDK_LOAD_CONFIG") != "" {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
	if os.Getenv(prefix+"AWS_SQS_INCLUDE_ID") == "true" {
		d.IncludeID = true
	}
	return nil
}

func (d *SQS) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.Queue = *flags.SQSQueueURL
	d.Region = *flags.AWSRegion
	d.RoleARN = *flags.AWSRoleARN
	d.IncludeID = *flags.AWSSQSIncludeID
	if flags.AWSLoadConfig != nil && *flags.AWSLoadConfig {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
	return nil
}

func (d *SQS) LogIdentity() error {
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

func (d *SQS) Init() error {
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
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
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
	d.Client = sqs.New(sess, cfg)
	d.sts = &STSSession{
		Session: sess,
		Config:  cfg,
	}
	return err
}

func (d *SQS) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "GetWork",
	})
	l.Debug("GetWork")
	var an []*string
	// assume some filtering would be done
	an = append(an, aws.String("All"))
	var man []*string
	man = append(man, aws.String("All"))
	rmi := &sqs.ReceiveMessageInput{
		// set queue URL
		QueueUrl:       aws.String(d.Queue),
		AttributeNames: an,
		// retrieve all
		MessageAttributeNames: man,
		// retrieve one message at a time
		MaxNumberOfMessages: aws.Int64(1),
		// do not timeout visibility - for testing
		//VisibilityTimeout: aws.Int64(0),
	}
	m, err := d.Client.ReceiveMessage(rmi)
	if err != nil {
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return nil, err
	}
	if len(m.Messages) < 1 {
		return nil, nil
	}
	md := m.Messages[0]
	var body string
	if d.IncludeID {
		var resp struct {
			ID   string `json:"id"`
			Body string `json:"body"`
		}
		resp.ID = *md.MessageId
		resp.Body = *md.Body
		b, err := json.Marshal(resp)
		if err != nil {
			return nil, err
		}
		body = string(b)
	} else {
		body = *md.Body
	}
	d.ReceiptHandle = *md.ReceiptHandle
	return strings.NewReader(body), nil
}

func (d *SQS) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "ClearWork",
	})
	l.Debug("ClearWork")
	di := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(d.Queue),
		ReceiptHandle: aws.String(d.ReceiptHandle),
	}
	_, err := d.Client.DeleteMessage(di)
	if err != nil {
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return err
	}
	return nil
}

func (d *SQS) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "HandleFailure",
	})
	l.Debug("HandleFailure")
	return nil
}

func (d *SQS) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "aws",
		"fn":  "Cleanup",
	})
	l.Debug("Cleanup")
	return nil
}
