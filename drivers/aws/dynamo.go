package aws

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/google/uuid"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Dynamo struct {
	Client           *dynamodb.DynamoDB
	sts              *STSSession
	Table            string
	Region           string
	RetrieveField    *string
	Limit            *int64
	NextToken        *string
	IncludeNextToken bool
	RoleARN          string
	RetrieveQuery    *string
	ClearQuery       *string
	FailQuery        *string
	data             map[string]any
}

func (d *Dynamo) LogIdentity() error {
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

func (d *Dynamo) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"fn":  "LoadEnv",
		"pkg": "aws",
	})
	l.Debug("LoadEnv")
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
	if os.Getenv(prefix+"AWS_DYNAMO_RETRIEVE_FIELD") != "" {
		f := os.Getenv(prefix + "AWS_DYNAMO_RETRIEVE_FIELD")
		d.RetrieveField = &f
	}
	if os.Getenv(prefix+"AWS_DYNAMO_INCLUDE_NEXT_TOKEN") != "" {
		v := os.Getenv(prefix+"AWS_DYNAMO_INCLUDE_NEXT_TOKEN") == "true"
		d.IncludeNextToken = v
	}
	if os.Getenv(prefix+"AWS_DYNAMO_LIMIT") != "" {
		v := os.Getenv(prefix + "AWS_DYNAMO_LIMIT")
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			l.Errorf("%+v", err)
			return err
		}
		d.Limit = &i
	}
	if os.Getenv(prefix+"AWS_DYNAMO_NEXT_TOKEN") != "" {
		v := os.Getenv(prefix + "AWS_DYNAMO_NEXT_TOKEN")
		d.NextToken = &v
	}
	if os.Getenv(prefix+"AWS_LOAD_CONFIG") != "" || os.Getenv("AWS_SDK_LOAD_CONFIG") != "" {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
	return nil
}

func (d *Dynamo) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"fn":  "LoadFlags",
		"pkg": "aws",
	})
	l.Debug("LoadFlags")
	d.Table = *flags.AWSDynamoTable
	d.Region = *flags.AWSRegion
	d.RoleARN = *flags.AWSRoleARN
	d.RetrieveQuery = flags.AWSDynamoRetrieveQuery
	d.RetrieveField = flags.AWSDynamoRetrieveField
	d.ClearQuery = flags.AWSDynamoClearQuery
	d.FailQuery = flags.AWSDynamoFailQuery
	d.IncludeNextToken = *flags.AWSDynamoIncludeNextToken
	iv := int64(*flags.AWSDynamoLimit)
	d.Limit = &iv
	d.NextToken = flags.AWSDynamoNextToken
	if flags.AWSLoadConfig != nil && *flags.AWSLoadConfig {
		os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	}
	return nil
}

func (d *Dynamo) Init() error {
	l := log.WithFields(
		log.Fields{
			"fn":  "CreateAWSSession",
			"pkg": "aws",
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
	d.Client = dynamodb.New(sess, cfg)
	d.sts = &STSSession{
		Session: sess,
		Config:  cfg,
	}
	return err
}

func (d *Dynamo) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"fn":  "GetWork",
		"pkg": "aws",
	})
	l.Debug("GetWork")
	if d.RetrieveQuery == nil {
		return nil, errors.New("retrieve query not set")
	}
	// execute statement
	statement := dynamodb.ExecuteStatementInput{
		Statement: d.RetrieveQuery,
	}
	if *d.Limit > 0 {
		statement.Limit = d.Limit
	}
	if d.NextToken != nil && *d.NextToken != "" {
		statement.NextToken = d.NextToken
	}
	resp, err := d.Client.ExecuteStatement(&statement)
	if err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
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
	err = dynamodbattribute.UnmarshalMap(item, &d.data)
	if err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return nil, err
	}
	var result string
	if d.RetrieveField != nil && *d.RetrieveField != "" {
		bd, err := json.Marshal(d.data)
		if err != nil {
			l.Errorf("%+v", err)
			if err := d.LogIdentity(); err != nil {
				l.Errorf("%+v", err)
			}
			return nil, err
		}
		l.Debug("GetWork item=%s", string(bd))
		result = gjson.GetBytes(bd, *d.RetrieveField).String()
	} else {
		td := make(map[string]any)
		for k, v := range d.data {
			td[k] = v
		}
		if d.IncludeNextToken {
			td["_nextToken"] = resp.NextToken
		}
		jd, err := schema.MapStringAnyToJSON(td)
		if err != nil {
			l.Error(err)
			return nil, err
		}
		result = string(jd)
	}
	// if result is empty, return nil
	if result == "" {
		l.Debug("result is empty")
		return nil, nil
	}
	l.Debug("Got work")
	return strings.NewReader(result), nil
}

func (d *Dynamo) clearQuery() string {
	l := log.WithFields(log.Fields{
		"fn":  "clearQuery",
		"pkg": "aws",
	})
	l.Debug("clearQuery")
	if d.ClearQuery == nil {
		return ""
	}
	td := make(map[string]any)
	for k, v := range d.data {
		td[k] = v
	}
	q := schema.ReplaceParamsMapString(td, *d.ClearQuery)
	l.Debugf("clearQuery query=%s", q)
	return q
}

func (d *Dynamo) failQuery() string {
	l := log.WithFields(log.Fields{
		"fn":  "failQuery",
		"pkg": "aws",
	})
	l.Debug("failQuery")
	if d.FailQuery == nil {
		return ""
	}
	td := make(map[string]any)
	for k, v := range d.data {
		td[k] = v
	}
	q := schema.ReplaceParamsMapString(td, *d.FailQuery)
	l.Debugf("failQuery query=%s", q)
	return q
}

func (d *Dynamo) ClearWork() error {
	l := log.WithFields(log.Fields{
		"fn":  "ClearWork",
		"pkg": "aws",
	})
	l.Debug("ClearWork")
	if d.ClearQuery == nil || *d.ClearQuery == "" {
		return nil
	}
	// replace {{key}} with key
	q := d.clearQuery()
	// execute statement
	statement := dynamodb.ExecuteStatementInput{
		Statement: &q,
	}
	_, err := d.Client.ExecuteStatement(&statement)
	if err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return err
	}
	return nil
}

func (d *Dynamo) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"fn":  "HandleFailure",
		"pkg": "aws",
	})
	l.Debug("HandleFailure")
	if d.FailQuery == nil || *d.FailQuery == "" {
		return nil
	}
	// replace {{key}} with key
	q := d.failQuery()
	// execute statement
	statement := dynamodb.ExecuteStatementInput{
		Statement: &q,
	}
	_, err := d.Client.ExecuteStatement(&statement)
	if err != nil {
		l.Errorf("%+v", err)
		if err := d.LogIdentity(); err != nil {
			l.Errorf("%+v", err)
		}
		return err
	}
	return nil
}

func (d *Dynamo) Cleanup() error {
	l := log.WithFields(log.Fields{
		"fn":  "Cleanup",
		"pkg": "aws",
	})
	l.Debug("Cleanup")
	return nil
}
