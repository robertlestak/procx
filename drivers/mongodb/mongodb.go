package mongodb

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ClearQuery    *string
	FailQuery     *string
	Key           *string
	// TLS
	EnableTLS   *bool
	TLSInsecure *bool
	TLSCert     *string
	TLSKey      *string
	TLSCA       *string
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
	d.Collection = *flags.MongoCollection
	d.RetrieveQuery = flags.MongoRetrieveQuery
	d.ClearQuery = flags.MongoClearQuery
	d.FailQuery = flags.MongoFailQuery
	d.EnableTLS = flags.MongoEnableTLS
	d.TLSInsecure = flags.MongoTLSInsecure
	d.TLSCert = flags.MongoCertFile
	d.TLSKey = flags.MongoKeyFile
	d.TLSCA = flags.MongoCAFile
	return nil
}

func (d *Mongo) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "mongo",
		"fn":  "Init",
	})
	l.Debug("Initializing mongo client")
	var err error
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", d.User, d.Password, d.Host, d.Port, d.DB)
	l.Debug("uri: ", uri)
	opts := options.Client().ApplyURI(uri)
	if d.EnableTLS != nil && *d.EnableTLS {
		l.Debug("TLS enabled")
		tlsConfig := &tls.Config{
			InsecureSkipVerify: *d.TLSInsecure,
		}
		if *d.TLSCert != "" && *d.TLSKey != "" {
			l.Debug("TLS cert and key provided")
			cert, err := tls.LoadX509KeyPair(*d.TLSCert, *d.TLSKey)
			if err != nil {
				return err
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		if *d.TLSCA != "" {
			l.Debug("TLS CA provided")
			caCert, err := ioutil.ReadFile(*d.TLSCA)
			if err != nil {
				return err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
		opts.SetTLSConfig(tlsConfig)
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
	var key string
	coll := d.Client.Database(d.DB).Collection(d.Collection)
	// unmarshal query string into struct
	var bsonMap bson.M
	err = json.Unmarshal([]byte(*d.RetrieveQuery), &bsonMap)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	var res bson.M
	// get the first document from the collection that matches the query
	err = coll.FindOne(context.TODO(), bsonMap).Decode(&res)
	if err != nil {
		// if no document is found, return nil
		if err == mongo.ErrNoDocuments {
			l.Debug("no documents found")
			return nil, nil
		}
		l.Error(err)
		return nil, err
	}
	// get string id
	id := res["_id"].(primitive.ObjectID).Hex()
	l.Debug("id: ", id)
	key = id
	jd, err := json.Marshal(res)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	d.Key = &key
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
	if d.Key != nil {
		l = l.WithField("key", *d.Key)
	}
	if d.Key == nil {
		return nil
	}
	l.Debug("Clearing work from mongo")
	var err error
	dbconn := d.Client.Database(d.DB)
	var result bson.M
	// replace object ID in query string with string id
	query := strings.Replace(*d.ClearQuery, "{{key}}", *d.Key, -1)
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
	if d.Key == nil {
		return nil
	}
	if d.FailQuery == nil || *d.FailQuery == "" {
		return nil
	}
	var err error
	dbconn := d.Client.Database(d.DB)
	var result bson.M
	// replace object ID in query string with string id
	query := strings.Replace(*d.FailQuery, "{{key}}", *d.Key, -1)
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
