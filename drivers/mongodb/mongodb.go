package mongodb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/schema"
	"github.com/robertlestak/procx/pkg/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	Client        *mongo.Client
	Host          string
	Port          int
	User          string
	Password      string
	DB            string
	Collection    string
	RetrieveQuery *string
	Limit         *int64
	ClearQuery    *string
	FailQuery     *string
	AuthSource    *string
	// TLS
	EnableTLS   *bool
	TLSInsecure *bool
	TLSCert     *string
	TLSKey      *string
	TLSCA       *string
	data        []bson.M
}

func (d *Mongo) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "mongo",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment variables")
	if os.Getenv(prefix+"MONGO_HOST") != "" {
		d.Host = os.Getenv(prefix + "MONGO_HOST")
	}
	if os.Getenv(prefix+"MONGO_PORT") != "" {
		pval, err := strconv.Atoi(os.Getenv(prefix + "MONGO_PORT"))
		if err != nil {
			return err
		}
		d.Port = pval
	}
	if os.Getenv(prefix+"MONGO_USER") != "" {
		d.User = os.Getenv(prefix + "MONGO_USER")
	}
	if os.Getenv(prefix+"MONGO_PASSWORD") != "" {
		d.Password = os.Getenv(prefix + "MONGO_PASSWORD")
	}
	if os.Getenv(prefix+"MONGO_DATABASE") != "" {
		d.DB = os.Getenv(prefix + "MONGO_DATABASE")
	}
	if os.Getenv(prefix+"MONGO_RETRIEVE_QUERY") != "" {
		qv := os.Getenv(prefix + "MONGO_RETRIEVE_QUERY")
		d.RetrieveQuery = &qv
	}
	if os.Getenv(prefix+"MONGO_CLEAR_QUERY") != "" {
		qv := os.Getenv(prefix + "MONGO_CLEAR_QUERY")
		d.ClearQuery = &qv
	}
	if os.Getenv(prefix+"MONGO_FAIL_QUERY") != "" {
		qv := os.Getenv(prefix + "MONGO_FAIL_QUERY")
		d.FailQuery = &qv
	}
	if os.Getenv(prefix+"MONGO_COLLECTION") != "" {
		c := os.Getenv(prefix + "MONGO_COLLECTION")
		d.Collection = strings.TrimSpace(c)
	}
	if os.Getenv(prefix+"MONGO_ENABLE_TLS") != "" {
		v := os.Getenv(prefix+"MONGO_ENABLE_TLS") == "true"
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"MONGO_TLS_INSECURE") != "" {
		v := os.Getenv(prefix+"MONGO_TLS_INSECURE") == "true"
		d.TLSInsecure = &v
	}
	if os.Getenv(prefix+"MONGO_TLS_CERT_FILE") != "" {
		v := os.Getenv(prefix + "MONGO_TLS_CERT_FILE")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"MONGO_TLS_KEY_FILE") != "" {
		v := os.Getenv(prefix + "MONGO_TLS_KEY_FILE")
		d.TLSKey = &v
	}
	if os.Getenv(prefix+"MONGO_TLS_CA_FILE") != "" {
		v := os.Getenv(prefix + "MONGO_TLS_CA_FILE")
		d.TLSCA = &v
	}
	if os.Getenv(prefix+"MONGO_AUTH_SOURCE") != "" {
		v := os.Getenv(prefix + "MONGO_AUTH_SOURCE")
		d.AuthSource = &v
	}
	if os.Getenv(prefix+"MONGO_LIMIT") != "" {
		v, err := strconv.Atoi(os.Getenv(prefix + "MONGO_LIMIT"))
		if err != nil {
			return err
		}
		iv := int64(v)
		d.Limit = &iv
	}
	return nil
}

func (d *Mongo) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "mongo",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	pv, err := strconv.Atoi(*flags.MongoPort)
	if err != nil {
		return err
	}
	d.Host = *flags.MongoHost
	d.Port = pv
	d.User = *flags.MongoUser
	d.Password = *flags.MongoPassword
	d.DB = *flags.MongoDatabase
	iv := int64(*flags.MongoLimit)
	d.Limit = &iv
	d.Collection = *flags.MongoCollection
	d.RetrieveQuery = flags.MongoRetrieveQuery
	d.ClearQuery = flags.MongoClearQuery
	d.FailQuery = flags.MongoFailQuery
	d.EnableTLS = flags.MongoEnableTLS
	d.TLSInsecure = flags.MongoTLSInsecure
	d.TLSCert = flags.MongoCertFile
	d.TLSKey = flags.MongoKeyFile
	d.TLSCA = flags.MongoCAFile
	d.AuthSource = flags.MongoAuthSource
	return nil
}

func (d *Mongo) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "mongo",
		"fn":  "Init",
	})
	l.Debug("Initializing mongo client")
	var err error
	var uri string
	if d.AuthSource != nil && *d.AuthSource != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s", d.User, d.Password, d.Host, d.Port, d.DB, *d.AuthSource)
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", d.User, d.Password, d.Host, d.Port, d.DB)
	}
	l.Debug("uri: ", uri)
	opts := options.Client().ApplyURI(uri)
	if d.EnableTLS != nil && *d.EnableTLS {
		l.Debug("TLS enabled")
		tc, err := utils.TlsConfig(d.EnableTLS, d.TLSInsecure, d.TLSCA, d.TLSCert, d.TLSKey)
		if err != nil {
			return err
		}
		opts.SetTLSConfig(tc)
	}
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		l.Error(err)
		return err
	}
	d.Client = client
	// ping the database to check if it is alive
	err = d.Client.Ping(context.TODO(), nil)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *Mongo) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "mongo",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from mongo")
	if d.RetrieveQuery == nil || *d.RetrieveQuery == "" {
		l.Error("query is empty")
		return nil, errors.New("query is empty")
	}
	var err error
	coll := d.Client.Database(d.DB).Collection(d.Collection)
	// unmarshal query string into struct
	var bsonMap bson.M
	err = json.Unmarshal([]byte(*d.RetrieveQuery), &bsonMap)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	var arr []bson.M
	// find all documents that match the query and return them in an array
	findOptions := options.Find()
	if d.Limit != nil && *d.Limit > 0 {
		findOptions.SetLimit(*d.Limit)
	}
	curr, err := coll.Find(context.TODO(), bsonMap, findOptions)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	for curr.Next(context.TODO()) {
		var elem bson.M
		err := curr.Decode(&elem)
		if err != nil {
			l.Error(err)
			return nil, err
		}
		arr = append(arr, elem)
	}
	if err := curr.Err(); err != nil {
		l.Error(err)
		return nil, err
	}
	curr.Close(context.TODO())
	if len(arr) == 0 {
		l.Debug("No work found")
		return nil, nil
	}
	d.data = arr
	jd, err := json.Marshal(arr)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	return bytes.NewReader(jd), nil
}

func (d *Mongo) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg":   "mongo",
		"fn":    "ClearWork",
		"query": d.ClearQuery,
	})
	if d.ClearQuery == nil || *d.ClearQuery == "" {
		return nil
	}
	l.Debug("Clearing work from mongo")
	var err error
	dbconn := d.Client.Database(d.DB)
	var result bson.M
	jd, err := json.Marshal(d.data)
	if err != nil {
		l.Error(err)
		return err
	}
	query := schema.ReplaceParamsString(jd, *d.ClearQuery)
	l = l.WithField("newQuery", query)
	var command bson.D
	err = bson.UnmarshalExtJSON([]byte(query), true, &command)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("command: ", command)
	err = dbconn.RunCommand(
		context.TODO(),
		command,
	).Decode(&result)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	l.Debug("result: ", result)
	return nil
}

func (d *Mongo) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "mongo",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure from mongo")
	if d.FailQuery == nil || *d.FailQuery == "" {
		return nil
	}
	var err error
	dbconn := d.Client.Database(d.DB)
	var result bson.M
	jd, err := json.Marshal(d.data)
	if err != nil {
		l.Error(err)
		return err
	}
	query := schema.ReplaceParamsString(jd, *d.FailQuery)
	l = l.WithField("newQuery", query)
	var command bson.D
	err = bson.UnmarshalExtJSON([]byte(query), true, &command)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("command: ", command)
	err = dbconn.RunCommand(
		context.TODO(),
		command,
	).Decode(&result)
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("handled failed work")
	return nil
}

func (d *Mongo) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "mongo",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up mongo")
	if d.Client == nil {
		return nil
	}
	err := d.Client.Disconnect(context.TODO())
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleaned up mongo")
	return nil
}
