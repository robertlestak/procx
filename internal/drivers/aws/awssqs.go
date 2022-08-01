package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
	"github.com/robertlestak/procx/internal/flags"
	log "github.com/sirupsen/logrus"
)

type SQS struct {
	Client        *sqs.SQS
	Queue         string
	ReceiptHandle string
	Region        string
	RoleARN       string
}

func (d *SQS) LoadEnv(prefix string) error {
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
	return nil
}

func (d *SQS) LoadFlags() error {
	d.Queue = *flags.SQSQueueURL
	d.Region = *flags.AWSRegion
	d.RoleARN = *flags.AWSRoleARN
	if flags.AWSLoadConfig != nil && *flags.AWSLoadConfig {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
	return nil
}

func (d *SQS) Init() error {
	l := log.WithFields(
		log.Fields{
			"action": "CreateAWSSession",
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
	}
	d.Client = sqs.New(sess, cfg)
	return err
}

func (d *SQS) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "GetWork",
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
		return nil, err
	}
	if len(m.Messages) < 1 {
		return nil, nil
	}
	md := m.Messages[0]
	d.ReceiptHandle = *md.ReceiptHandle
	return md.Body, nil
}

func (d *SQS) ClearWork() error {
	l := log.WithFields(log.Fields{
		"action": "ClearWork",
	})
	l.Debug("ClearWork")
	di := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(d.Queue),
		ReceiptHandle: aws.String(d.ReceiptHandle),
	}
	_, err := d.Client.DeleteMessage(di)
	if err != nil {
		return err
	}
	return nil
}

func (d *SQS) HandleFailure() error {
	return nil
}
