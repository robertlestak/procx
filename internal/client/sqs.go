package client

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var (
	SQSClient        *sqs.SQS
	SQSQueueURL      string
	SQSReceiptHandle string
)

// CreateAWSSession will connect to AWS with the account's credentials from vault
func CreateAWSSession(region, roleArn string) (*session.Session, *aws.Config, error) {
	l := log.WithFields(
		log.Fields{
			"action": "CreateAWSSession",
		},
	)
	l.Debug("CreateAWSSession")
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	if region == "" {
		region = "us-east-1"
	}
	cfg := &aws.Config{
		Region: aws.String(region),
	}
	sess, err := session.NewSession(cfg)
	reqId := uuid.New().String()
	if roleArn != "" {
		l.Debug("CreateAWSSession roleArn=%s requestId=%s", roleArn, reqId)
		creds := stscreds.NewCredentials(sess, roleArn, func(p *stscreds.AssumeRoleProvider) {
			p.RoleSessionName = "procx-" + reqId
		})
		cfg.Credentials = creds
	}
	if err != nil {
		l.Errorf("%+v", err)
	}
	return sess, cfg, nil
}

func CreateSQSClient(region, roleArn string) (*sqs.SQS, error) {
	sess, cfg, err := CreateAWSSession(region, roleArn)
	if err != nil {
		return nil, err
	}
	SQSClient = sqs.New(sess, cfg)
	return SQSClient, nil
}

// ReceiveMessage receives a single message from the queue
func ReceiveMessageSQS() (*sqs.Message, error) {
	var an []*string
	// assume some filtering would be done
	an = append(an, aws.String("All"))
	var man []*string
	man = append(man, aws.String("All"))
	rmi := &sqs.ReceiveMessageInput{
		// set queue URL
		QueueUrl:       aws.String(SQSQueueURL),
		AttributeNames: an,
		// retrieve all
		MessageAttributeNames: man,
		// retrieve one message at a time
		MaxNumberOfMessages: aws.Int64(1),
		// do not timeout visibility - for testing
		//VisibilityTimeout: aws.Int64(0),
	}
	m, err := SQSClient.ReceiveMessage(rmi)
	if err != nil {
		return nil, err
	}
	if len(m.Messages) < 1 {
		return nil, nil
	}
	return m.Messages[0], nil
}

// DeleteMessage deletes a message from the queue
func DeleteMessageSQS() error {
	di := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(SQSQueueURL),
		ReceiptHandle: aws.String(SQSReceiptHandle),
	}
	_, err := SQSClient.DeleteMessage(di)
	if err != nil {
		return err
	}
	return nil
}
