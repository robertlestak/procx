package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
)

func CreateMongoClient(host string, port int, user string, pass string, db string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "CreateMongoClient",
	})
	l.Debug("Initializing mongo client")
	var err error
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", user, pass, host, port, db)
	l.Debug("uri: ", uri)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		l.Error(err)
		return err
	}
	MongoClient = client
	// ping the database to check if it is alive
	err = MongoClient.Ping(context.TODO(), nil)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func GetWorkMongo(db, collection, query string) (*string, *string, error) {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "GetWorkMongo",
	})
	l.Debug("Getting work from mongo")
	if query == "" {
		l.Error("query is empty")
		return nil, nil, errors.New("query is empty")
	}
	var err error
	var result string
	var key string
	coll := MongoClient.Database(db).Collection(collection)
	// unmarshal query string into struct
	var bsonMap bson.M
	err = json.Unmarshal([]byte(query), &bsonMap)
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}
	var res bson.M
	// get the first document from the collection that matches the query
	err = coll.FindOne(context.TODO(), bsonMap).Decode(&res)
	if err != nil {
		// if no document is found, return nil
		if err == mongo.ErrNoDocuments {
			l.Debug("no documents found")
			return nil, nil, nil
		}
		l.Error(err)
		return nil, nil, err
	}
	// get string id
	id := res["_id"].(primitive.ObjectID).Hex()
	l.Debug("id: ", id)
	key = id
	jd, err := json.Marshal(res)
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}
	result = string(jd)
	return &result, &key, nil
}

func ClearWorkMongo(db, collection string, query string, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "ClearWorkMongo",
		"query":   query,
	})
	if key != nil {
		l = l.WithField("key", *key)
	}
	if key == nil {
		return nil
	}
	l.Debug("Clearing work from mongo")
	var err error
	dbconn := MongoClient.Database(db)
	var result bson.M
	// replace object ID in query string with string id
	query = strings.Replace(query, "{{key}}", *key, -1)
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

func HandleFailureMongo(db, collection string, query string, key *string) error {
	l := log.WithFields(log.Fields{
		"package": "cache",
		"method":  "HandleFailureMongo",
	})
	l.Debug("Handling failure from mongo")
	if key == nil {
		return nil
	}
	if query == "" {
		return nil
	}
	var err error
	dbconn := MongoClient.Database(db)
	var result bson.M
	// replace object ID in query string with string id
	query = strings.Replace(query, "{{key}}", *key, -1)
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
