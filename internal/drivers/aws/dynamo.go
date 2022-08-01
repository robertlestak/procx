package aws

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"github.com/robertlestak/procx/internal/flags"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Dynamo struct {
	Client           *dynamodb.DynamoDB
	Table            string
	Region           string
	RoleARN          string
	QueryKeyJSONPath *string
	RetrieveQuery    *string
	ClearQuery       *string
	FailQuery        *string
	Key              *string
}

func (d *Dynamo) LoadEnv(prefix string) error {
	if os.Getenv(prefix+"AWS_REGION") != "" {
		d.Region = os.Getenv(prefix + "AWS_REGION")
	}
	if os.Getenv(prefix+"AWS_ROLE_ARN") != "" {
		d.RoleARN = os.Getenv(prefix + "AWS_SQS_ROLE_ARN")
	}
	if os.Getenv(prefix+"AWS_DYNAMO_TABLE") != "" {
		d.Table = os.Getenv(prefix + "AWS_DYNAMO_TABLE")
	}
	if os.Getenv(prefix+"AWS_DYNAMO_RETRIEVE_QUERY") != "" {
		q := os.Getenv(prefix + "AWS_DYNAMO_RETRIEVE_QUERY")
		d.RetrieveQuery = &q
	}
	if os.Getenv(prefix+"AWS_DYNAMO_CLEAR_QUERY") != "" {
		q := os.Getenv(prefix + "AWS_DYNAMO_CLEAR_QUERY")
		d.ClearQuery = &q
	}
	if os.Getenv(prefix+"AWS_DYNAMO_FAIL_QUERY") != "" {
		q := os.Getenv(prefix + "AWS_DYNAMO_FAIL_QUERY")
		d.FailQuery = &q
	}
	if os.Getenv(prefix+"AWS_DYNAMO_KEY_PATH") != "" {
		q := os.Getenv(prefix + "AWS_DYNAMO_KEY_PATH")
		d.QueryKeyJSONPath = &q
	}
	if os.Getenv(prefix+"AWS_LOAD_CONFIG") != "" || os.Getenv("AWS_SDK_LOAD_CONFIG") != "" {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
	return nil
}

func (d *Dynamo) LoadFlags() error {
	d.Table = *flags.AWSDynamoTable
	d.Region = *flags.AWSRegion
	d.RoleARN = *flags.AWSRoleARN
	d.RetrieveQuery = flags.AWSDynamoRetrieveQuery
	d.QueryKeyJSONPath = flags.AWSDynamoQueryKeyPath
	d.ClearQuery = flags.AWSDynamoClearQuery
	d.FailQuery = flags.AWSDynamoFailQuery
	if flags.AWSLoadConfig != nil && *flags.AWSLoadConfig {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
	return nil
}

func (d *Dynamo) Init() error {
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
	d.Client = dynamodb.New(sess, cfg)
	return err
}

func (d *Dynamo) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"action": "GetWork",
	})
	l.Debug("GetWork")
	if d.RetrieveQuery == nil {
		return nil, errors.New("retrieve query not set")
	}
	// execute statement
	statement := dynamodb.ExecuteStatementInput{
		Statement: d.RetrieveQuery,
		Limit:     aws.Int64(1),
	}
	resp, err := d.Client.ExecuteStatement(&statement)
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	l.Debug("GetWork response=%+v", resp)
	if resp.Items == nil {
		l.Debug("GetWork no items")
		return nil, nil
	}
	if len(resp.Items) == 0 {
		l.Debug("GetWork no items")
		return nil, nil
	}
	// get first item
	item := resp.Items[0]
	if item == nil {
		l.Debug("GetWork no items")
		return nil, nil
	}
	jd, err := json.Marshal(item)
	if err != nil {
		l.Errorf("%+v", err)
		return nil, err
	}
	rd := aws.String(string(jd))
	l.Debug("GetWork item=%s", rd)
	if d.QueryKeyJSONPath != nil {
		if err := d.extractKey(rd); err != nil {
			l.Errorf("%+v", err)
			return nil, err
		}
	}
	return rd, nil
}

func (d *Dynamo) ClearWork() error {
	l := log.WithFields(log.Fields{
		"action": "ClearWork",
	})
	l.Debug("ClearWork")
	if d.ClearQuery == nil || *d.ClearQuery == "" {
		return nil
	}
	// replace {{key}} with key
	q := *d.ClearQuery
	if d.Key != nil {
		q = strings.Replace(q, "{{key}}", *d.Key, -1)
	}
	// execute statement
	statement := dynamodb.ExecuteStatementInput{
		Statement: &q,
	}
	_, err := d.Client.ExecuteStatement(&statement)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	return nil
}

func (d *Dynamo) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"action": "HandleFailure",
	})
	l.Debug("HandleFailure")
	if d.FailQuery == nil || *d.FailQuery == "" {
		return nil
	}
	// replace {{key}} with key
	q := *d.FailQuery
	if d.Key != nil {
		q = strings.Replace(q, "{{key}}", *d.Key, -1)
	}
	// execute statement
	statement := dynamodb.ExecuteStatementInput{
		Statement: &q,
	}
	_, err := d.Client.ExecuteStatement(&statement)
	if err != nil {
		l.Errorf("%+v", err)
		return err
	}
	return nil
}

func (d *Dynamo) extractKey(data *string) error {
	l := log.WithFields(log.Fields{
		"action": "extractKey",
	})
	l.Debug("extractKey")
	if d.QueryKeyJSONPath == nil {
		return nil
	}
	if *d.QueryKeyJSONPath == "" {
		return nil
	}
	if data == nil {
		return nil
	}
	if *data == "" {
		return nil
	}
	value := gjson.Get(*data, *d.QueryKeyJSONPath)
	if value.Exists() {
		l.Debugf("extractKey value=%s", value.String())
		k := value.String()
		d.Key = &k
	}
	return nil
}
