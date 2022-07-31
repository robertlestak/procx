package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robertlestak/procx/internal/flags"
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
}

func (d *Mongo) LoadEnv(prefix string) error {
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
	return nil
}

func (d *Mongo) LoadFlags() error {
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
	return nil
}

func (d *Mongo) Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreateMongoClient",
	})
	l.Debug("Initializing mongo client")
	var err error
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", d.User, d.Password, d.Host, d.Port, d.DB)
	l.Debug("uri: ", uri)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
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

func (d *Mongo) GetWork() (*string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWorkMongo",
	})
	l.Debug("Getting work from mongo")
	if d.RetrieveQuery == nil || *d.RetrieveQuery == "" {
		l.Error("query is empty")
		return nil, errors.New("query is empty")
	}
	var err error
	var result string
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
	result = string(jd)
	d.Key = &key
	return &result, nil
}

func (d *Mongo) ClearWork() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkMongo",
		"query":   d.ClearQuery,
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
		"package": "cache",
		"method":  "HandleFailureMongo",
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
