package gcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"google.golang.org/api/iterator"
)

type FirestoreOp string

var (
	FirestoreRMOp     = FirestoreOp("rm")
	FirestoreMVOp     = FirestoreOp("mv")
	FirestoreUpdateOp = FirestoreOp("update")
)

type GCPFirestoreQuery struct {
	Path    *string
	Op      *string
	Value   any
	OrderBy *string
	Order   *firestore.Direction
}

type GCPFirestore struct {
	Client                  *firestore.Client
	RetrieveCollection      *string
	RetrieveDocument        *string
	RetrieveQuery           *GCPFirestoreQuery
	RetrieveDocumentJSONKey *string
	Limit                   *int
	ClearOp                 *FirestoreOp
	ClearUpdate             *map[string]any
	ClearCollection         *string
	FailOp                  *FirestoreOp
	FailUpdate              *map[string]any
	FailCollection          *string
	ProjectID               string
	doc                     []map[string]any
}

func (d *GCPFirestore) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "LoadEnv",
	})
	l.Debug("LoadEnv")
	if os.Getenv(prefix+"GCP_PROJECT_ID") != "" {
		d.ProjectID = os.Getenv(prefix + "GCP_PROJECT_ID")
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_RETRIEVE_COLLECTION") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_RETRIEVE_COLLECTION")
		d.RetrieveCollection = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_RETRIEVE_DOCUMENT") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_RETRIEVE_DOCUMENT")
		d.RetrieveDocument = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_RETRIEVE_DOCUMENT_JSON_KEY") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_RETRIEVE_DOCUMENT_JSON_KEY")
		d.RetrieveDocumentJSONKey = &v
	}
	if d.RetrieveQuery == nil {
		d.RetrieveQuery = &GCPFirestoreQuery{}
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_RETRIEVE_QUERY_PATH") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_RETRIEVE_QUERY_PATH")
		d.RetrieveQuery.Path = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_RETRIEVE_QUERY_OP") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_RETRIEVE_QUERY_OP")
		d.RetrieveQuery.Op = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_RETRIEVE_QUERY_VALUE") != "" {
		vs := os.Getenv(prefix + "GCP_FIRESTORE_RETRIEVE_QUERY_VALUE")
		var v interface{}
		jd, err := json.Marshal(vs)
		if err != nil {
			l.Error(err)
			return err
		}
		err = json.Unmarshal(jd, &v)
		if err != nil {
			l.Error(err)
			return err
		}
		d.RetrieveQuery.Value = v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_RETRIEVE_QUERY_ORDER_BY") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_RETRIEVE_QUERY_ORDER_BY")
		d.RetrieveQuery.OrderBy = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_RETRIEVE_QUERY_ORDER") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_RETRIEVE_QUERY_ORDER")
		if strings.EqualFold(v, "asc") {
			tv := firestore.Asc
			d.RetrieveQuery.Order = &tv
		} else {
			tv := firestore.Desc
			d.RetrieveQuery.Order = &tv
		}
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_RETRIEVE_LIMIT") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_RETRIEVE_LIMIT")
		iv, err := strconv.Atoi(v)
		if err != nil {
			l.Error(err)
			return err
		}
		d.Limit = &iv
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_CLEAR_OP") != "" {
		v := FirestoreOp(os.Getenv(prefix + "GCP_FIRESTORE_CLEAR_OP"))
		d.ClearOp = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_CLEAR_COLLECTION") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_CLEAR_COLLECTION")
		d.ClearCollection = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_FAIL_OP") != "" {
		v := FirestoreOp(os.Getenv(prefix + "GCP_FIRESTORE_FAIL_OP"))
		d.FailOp = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_FAIL_COLLECTION") != "" {
		v := os.Getenv(prefix + "GCP_FIRESTORE_FAIL_COLLECTION")
		d.FailCollection = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_FAIL_UPDATE") != "" {
		vs := os.Getenv(prefix + "GCP_FIRESTORE_FAIL_UPDATE")
		var v map[string]interface{}
		err := json.Unmarshal([]byte(vs), &v)
		if err != nil {
			l.Error(err)
			return err
		}
		d.FailUpdate = &v
	}
	if os.Getenv(prefix+"GCP_FIRESTORE_CLEAR_UPDATE") != "" {
		vs := os.Getenv(prefix + "GCP_FIRESTORE_CLEAR_UPDATE")
		var v map[string]interface{}
		err := json.Unmarshal([]byte(vs), &v)
		if err != nil {
			l.Error(err)
			return err
		}
		d.ClearUpdate = &v
	}
	return nil
}

func (d *GCPFirestore) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "LoadFlags",
	})
	l.Debug("LoadFlags")
	d.ProjectID = *flags.GCPProjectID
	d.RetrieveCollection = flags.GCPFirestoreRetrieveCollection
	d.RetrieveDocument = flags.GCPFirestoreRetrieveDocument
	d.RetrieveDocumentJSONKey = flags.GCPFirestoreRetrieveDocumentJSONKey
	d.ClearCollection = flags.GCPFirestoreClearCollection
	d.FailCollection = flags.GCPFirestoreFailCollection
	if flags.GCPFirestoreClearUpdate != nil && *flags.GCPFirestoreClearUpdate != "" {
		var v map[string]interface{}
		err := json.Unmarshal([]byte(*flags.GCPFirestoreClearUpdate), &v)
		if err != nil {
			l.Error(err)
			return err
		}
		d.ClearUpdate = &v
	}
	if flags.GCPFirestoreFailUpdate != nil && *flags.GCPFirestoreFailUpdate != "" {
		var v map[string]interface{}
		err := json.Unmarshal([]byte(*flags.GCPFirestoreFailUpdate), &v)
		if err != nil {
			l.Error(err)
			return err
		}
		d.FailUpdate = &v
	}

	var v interface{}
	if flags.GCPFirestoreRetrieveQueryValue != nil && *flags.GCPFirestoreRetrieveQueryValue != "" {
		jd, err := json.Marshal(*flags.GCPFirestoreRetrieveQueryValue)
		if err != nil {
			l.Error(err)
			return err
		}
		err = json.Unmarshal(jd, &v)
		if err != nil {
			l.Error(err)
			return err
		}
	}
	d.RetrieveQuery = &GCPFirestoreQuery{
		Path:    flags.GCPFirestoreRetrieveQueryPath,
		Op:      flags.GCPFirestoreRetrieveQueryOp,
		Value:   v,
		OrderBy: flags.GCPFirestoreRetrieveQueryOrderBy,
	}
	if strings.EqualFold(*flags.GCPFirestoreRetrieveQueryOrder, "asc") {
		v := firestore.Asc
		d.RetrieveQuery.Order = &v
	} else {
		v := firestore.Desc
		d.RetrieveQuery.Order = &v
	}
	if *flags.GCPFirestoreClearOp != "" {
		o := FirestoreOp(*flags.GCPFirestoreClearOp)
		d.ClearOp = &o
	}
	if *flags.GCPFirestoreFailOp != "" {
		o := FirestoreOp(*flags.GCPFirestoreFailOp)
		d.FailOp = &o
	}
	return nil
}

func (d *GCPFirestore) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "Init",
	})
	l.Debug("Initializing gcp firestore driver")
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, d.ProjectID)
	if err != nil {
		return err
	}
	d.Client = client
	return nil
}

func (d *GCPFirestore) Data(obj []map[string]interface{}) (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "Data",
	})
	l.Debug("Getting data from gcp firestore driver")
	jd, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	if d.RetrieveDocumentJSONKey == nil || *d.RetrieveDocumentJSONKey == "" {
		return bytes.NewReader(jd), nil
	} else {
		val := gjson.ParseBytes(jd).Get(*d.RetrieveDocumentJSONKey)
		return strings.NewReader(val.String()), nil
	}
}

func (d *GCPFirestore) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from gcp firestore driver")
	if d.RetrieveCollection == nil {
		return nil, nil
	}
	ctx := context.Background()
	if d.RetrieveDocument != nil && *d.RetrieveDocument != "" {
		p := *d.RetrieveCollection + "/" + *d.RetrieveDocument
		l.Debug("Getting document: " + p)
		doc, err := d.Client.Doc(p).Get(ctx)
		if err != nil {
			return nil, err
		}
		dd := doc.Data()
		dd["_id"] = doc.Ref.ID
		return d.Data([]map[string]any{dd})
	} else if d.RetrieveQuery != nil && *d.RetrieveQuery.Path != "" {
		qry := d.Client.Collection(*d.RetrieveCollection).
			Where(*d.RetrieveQuery.Path, *d.RetrieveQuery.Op, d.RetrieveQuery.Value)
		if d.RetrieveQuery.OrderBy != nil && *d.RetrieveQuery.OrderBy != "" {
			qry = qry.OrderBy(*d.RetrieveQuery.OrderBy, *d.RetrieveQuery.Order)
		}
		// get first document
		if d.Limit != nil && *d.Limit > 0 {
			qry = qry.Limit(*d.Limit)
		}
		iter := qry.Documents(ctx)
		for {
			doc, err := iter.Next()
			if err != nil {
				if err == iterator.Done {
					break
				}
				return nil, err
			}
			dd := doc.Data()
			dd["_id"] = doc.Ref.ID
			d.doc = append(d.doc, dd)
		}
		return d.Data(d.doc)
	} else {
		// get the first document in the collection
		qry := d.Client.Collection(*d.RetrieveCollection).Limit(1)
		doc, err := qry.Documents(ctx).Next()
		if err != nil {
			if err == iterator.Done {
				return nil, nil
			}
			return nil, err
		}
		dd := doc.Data()
		dd["_id"] = doc.Ref.ID
		return d.Data([]map[string]any{dd})
	}
}

func (d *GCPFirestore) rmDoc(collection, id string) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "rmDoc",
	})
	l.Debug("Removing document from gcp firestore driver")
	if id == "" {
		return errors.New("no document to remove")
	}
	ctx := context.Background()
	if collection == "" {
		return errors.New("no collection to remove document from")
	}
	_, err := d.Client.Doc(collection + "/" + id).Delete(ctx)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *GCPFirestore) createDoc(collection string, id string) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "createDoc",
	})
	l.Debug("Creating document in gcp firestore driver")
	if collection == "" {
		return errors.New("no collection to create document in")
	}
	if id == "" {
		// generate new id
		id = uuid.New().String()
	}
	ctx := context.Background()
	_, err := d.Client.Doc(collection+"/"+id).Create(ctx, d.doc)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *GCPFirestore) mvDoc(collection string, id string) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "mvDoc",
	})
	l.Debug("Moving document in gcp firestore driver")
	if collection == "" {
		return errors.New("no collection to move document to")
	}
	if id == "" {
		return errors.New("no document to move")
	}
	if err := d.createDoc(collection, id); err != nil {
		return err
	}
	if err := d.rmDoc(collection, id); err != nil {
		return err
	}
	return nil
}

func (d *GCPFirestore) merge(merge map[string]interface{}) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "merge",
	})
	l.Debug("Merging document in gcp firestore driver")
	for i, _ := range d.doc {
		for k, v := range merge {
			d.doc[i][k] = v
		}
	}
	return nil
}

func (d *GCPFirestore) updateDoc(collection, id string, merge map[string]interface{}) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "updateDoc",
	})
	l.Debug("Updating document in gcp firestore driver")
	if id == "" {
		return errors.New("no document to update")
	}
	if collection == "" {
		return errors.New("no collection to update document in")
	}
	ctx := context.Background()
	_, err := d.Client.Doc(collection+"/"+id).Set(ctx, d.doc)
	if err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (d *GCPFirestore) rmDocs() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "rmDocs",
	})
	l.Debug("Removing documents from gcp firestore driver")
	var errs []error
	for _, doc := range d.doc {
		if err := d.rmDoc(*d.RetrieveCollection, doc["_id"].(string)); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.New("errors removing documents")
	}
	return nil
}

func (d *GCPFirestore) mvDocs(collection string) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "mvDocs",
	})
	l.Debug("Moving documents in gcp firestore driver")
	var errs []error
	for _, doc := range d.doc {
		if err := d.mvDoc(collection, doc["_id"].(string)); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.New("errors moving documents")
	}
	return nil
}

func (d *GCPFirestore) updateDocs(merge map[string]interface{}) error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "updateDocs",
	})
	l.Debug("Updating documents in gcp firestore driver")
	var errs []error
	for _, doc := range d.doc {
		if err := d.updateDoc(*d.RetrieveCollection, doc["_id"].(string), merge); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.New("errors updating documents")
	}
	return nil
}

func (d *GCPFirestore) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from gcp firestore driver")
	if d.ClearOp == nil {
		return nil
	}
	if d.ClearUpdate != nil && len(*d.ClearUpdate) > 0 {
		if err := d.merge(*d.ClearUpdate); err != nil {
			return err
		}
	}
	switch *d.ClearOp {
	case FirestoreRMOp:
		return d.rmDocs()
	case FirestoreMVOp:
		return d.mvDocs(*d.ClearCollection)
	case FirestoreUpdateOp:
		return d.updateDocs(*d.ClearUpdate)
	}
	return nil
}

func (d *GCPFirestore) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure in gcp firestore driver")
	if d.FailOp == nil {
		return nil
	}
	if d.FailUpdate != nil && len(*d.FailUpdate) > 0 {
		if err := d.merge(*d.FailUpdate); err != nil {
			return err
		}
	}
	switch *d.FailOp {
	case FirestoreRMOp:
		return d.rmDocs()
	case FirestoreMVOp:
		return d.mvDocs(*d.FailCollection)
	case FirestoreUpdateOp:
		return d.updateDocs(*d.FailUpdate)
	}
	return nil
}

func (d *GCPFirestore) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "gcp",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up gcp firestore driver")
	return nil
}
