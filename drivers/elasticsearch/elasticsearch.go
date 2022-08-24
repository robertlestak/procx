package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/robertlestak/procx/pkg/flags"
	"github.com/robertlestak/procx/pkg/schema"
	"github.com/robertlestak/procx/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type CloseOp string

var (
	CloseOpDelete   = CloseOp("delete")
	CloseOpPut      = CloseOp("put")
	CloseOpMergePut = CloseOp("merge-put")
	CloseOpMove     = CloseOp("move")
)

type Elasticsearch struct {
	Client   *elasticsearch8.Client
	Address  string
	Username string
	Password string
	// TLS
	EnableTLS     *bool
	TLSInsecure   *bool
	TLSCert       *string
	TLSKey        *string
	TLSCA         *string
	RetrieveIndex *string
	RetrieveQuery string
	ClearQuery    string
	ClearIndex    *string
	ClearOp       CloseOp
	FailQuery     string
	FailIndex     *string
	FailOp        CloseOp
	Key           *string
	data          []any
}

func (d *Elasticsearch) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	if os.Getenv(prefix+"ELASTICSEARCH_ADDRESS") != "" {
		d.Address = os.Getenv(prefix + "ELASTICSEARCH_ADDRESS")
	}
	if os.Getenv(prefix+"ELASTICSEARCH_USERNAME") != "" {
		d.Username = os.Getenv(prefix + "ELASTICSEARCH_USERNAME")
	}
	if os.Getenv(prefix+"ELASTICSEARCH_PASSWORD") != "" {
		d.Password = os.Getenv(prefix + "ELASTICSEARCH_PASSWORD")
	}
	if os.Getenv(prefix+"ELASTICSEARCH_TLS_SKIP_VERIFY") != "" {
		v := os.Getenv(prefix+"ELASTICSEARCH_TLS_SKIP_VERIFY") == "true"
		d.TLSInsecure = &v
	}
	if os.Getenv(prefix+"ELASTICSEARCH_RETRIEVE_QUERY") != "" {
		d.RetrieveQuery = os.Getenv(prefix + "ELASTICSEARCH_RETRIEVE_QUERY")
	}
	if os.Getenv(prefix+"ELASTICSEARCH_RETRIEVE_INDEX") != "" {
		v := os.Getenv(prefix + "ELASTICSEARCH_RETRIEVE_INDEX")
		d.RetrieveIndex = &v
	}
	if os.Getenv(prefix+"ELASTICSEARCH_CLEAR_QUERY") != "" {
		d.ClearQuery = os.Getenv(prefix + "ELASTICSEARCH_CLEAR_QUERY")
	}
	if os.Getenv(prefix+"ELASTICSEARCH_CLEAR_INDEX") != "" {
		v := os.Getenv(prefix + "ELASTICSEARCH_CLEAR_INDEX")
		d.ClearIndex = &v
	}
	if os.Getenv(prefix+"ELASTICSEARCH_CLEAR_OP") != "" {
		d.ClearOp = CloseOp(os.Getenv(prefix + "ELASTICSEARCH_CLEAR_OP"))
	}
	if os.Getenv(prefix+"ELASTICSEARCH_FAIL_QUERY") != "" {
		d.FailQuery = os.Getenv(prefix + "ELASTICSEARCH_FAIL_QUERY")
	}
	if os.Getenv(prefix+"ELASTICSEARCH_FAIL_INDEX") != "" {
		v := os.Getenv(prefix + "ELASTICSEARCH_FAIL_INDEX")
		d.FailIndex = &v
	}
	if os.Getenv(prefix+"ELASTICSEARCH_FAIL_OP") != "" {
		d.FailOp = CloseOp(os.Getenv(prefix + "ELASTICSEARCH_FAIL_OP"))
	}
	if os.Getenv(prefix+"ELASTICSEARCH_ENABLE_TLS") != "" {
		v := os.Getenv(prefix+"ELASTICSEARCH_ENABLE_TLS") == "true"
		d.EnableTLS = &v
	}
	if os.Getenv(prefix+"ELASTICSEARCH_TLS_CA_FILE") != "" {
		v := os.Getenv(prefix + "ELASTICSEARCH_TLS_CA_FILE")
		d.TLSCA = &v
	}
	if os.Getenv(prefix+"ELASTICSEARCH_TLS_CERT_FILE") != "" {
		v := os.Getenv(prefix + "ELASTICSEARCH_TLS_CERT_FILE")
		d.TLSCert = &v
	}
	if os.Getenv(prefix+"ELASTICSEARCH_TLS_KEY_FILE") != "" {
		v := os.Getenv(prefix + "ELASTICSEARCH_TLS_KEY_FILE")
		d.TLSKey = &v
	}
	return nil
}

func (d *Elasticsearch) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.Address = *flags.ElasticsearchAddress
	d.Username = *flags.ElasticsearchUsername
	d.Password = *flags.ElasticsearchPassword
	d.TLSInsecure = flags.ElasticsearchTLSSkipVerify
	d.EnableTLS = flags.ElasticsearchEnableTLS
	d.TLSCert = flags.ElasticsearchCertFile
	d.TLSKey = flags.ElasticsearchKeyFile
	d.TLSCA = flags.ElasticsearchCAFile
	d.RetrieveQuery = *flags.ElasticsearchRetrieveQuery
	d.RetrieveIndex = flags.ElasticsearchRetrieveIndex
	d.ClearQuery = *flags.ElasticsearchClearQuery
	d.ClearIndex = flags.ElasticsearchClearIndex
	d.ClearOp = CloseOp(*flags.ElasticsearchClearOp)
	d.FailQuery = *flags.ElasticsearchFailQuery
	d.FailIndex = flags.ElasticsearchFailIndex
	d.FailOp = CloseOp(*flags.ElasticsearchFailOp)
	return nil
}

func (d *Elasticsearch) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "Init",
	})
	l.Debug("Initializing elasticsearch driver")
	tc, err := utils.TlsConfig(d.EnableTLS, d.TLSInsecure, d.TLSCA, d.TLSCert, d.TLSKey)
	if err != nil {
		return err
	}
	client, err := elasticsearch8.NewClient(elasticsearch8.Config{
		Transport: &http.Transport{
			TLSClientConfig: tc,
		},
		Addresses: []string{d.Address},
		Username:  d.Username,
		Password:  d.Password,
	})
	if err != nil {
		l.Errorf("error creating client: %v", err)
		return err
	}
	d.Client = client
	return nil
}

func (d *Elasticsearch) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from elasticsearch")
	search := esapi.SearchRequest{
		Body: strings.NewReader(d.RetrieveQuery),
	}
	if d.RetrieveIndex != nil {
		search.Index = append(search.Index, *d.RetrieveIndex)
	}
	searchResponse, err := search.Do(context.Background(), d.Client)
	if err != nil {
		// if we get a 404, we have no work to do
		if searchResponse != nil && searchResponse.StatusCode == http.StatusNotFound {
			l.Debug("No work to do")
			return nil, nil
		}
		l.Errorf("error getting work: %v", err)
		return nil, err
	}
	if searchResponse.StatusCode != 200 {
		bd, err := ioutil.ReadAll(searchResponse.Body)
		if err != nil {
			l.Errorf("error reading response body: %v", err)
			return nil, err
		}
		l.Errorf("error getting work: %v", string(bd))
		return nil, errors.New("error getting work")
	}
	l.Debug("Got work")
	// parse response
	var r struct {
		Hits struct {
			Hits []struct {
				Index  string `json:"_index"`
				ID     string `json:"_id"`
				Source any    `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(searchResponse.Body).Decode(&r); err != nil {
		l.Errorf("error parsing response: %v", err)
		return nil, err
	}
	if len(r.Hits.Hits) == 0 {
		l.Debug("No work to do")
		return nil, nil
	}
	for _, hit := range r.Hits.Hits {
		l.Debugf("Got work: %v", hit.ID)
		// add _id to source
		s := hit.Source.(map[string]interface{})
		s["_id"] = hit.ID
		d.data = append(d.data, s)
	}
	jd, err := json.Marshal(d.data)
	if err != nil {
		l.Errorf("error marshalling source: %v", err)
		return nil, err
	}
	return bytes.NewReader(jd), nil
}

func mergeStringAndAny(s string, a any) (string, error) {
	var d1 map[string]interface{}
	if err := json.Unmarshal([]byte(s), &d1); err != nil {
		return "", err
	}
	var d2 map[string]interface{}
	jd, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(jd, &d2); err != nil {
		return "", err
	}
	// merge d1 and d2
	for k, v := range d1 {
		d2[k] = v
	}
	jd, err = json.Marshal(d2)
	if err != nil {
		return "", err
	}
	return string(jd), nil
}

func deleteDocument(c *elasticsearch8.Client, index *string, id string) error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "deleteDocument",
	})
	l.Debug("Deleting work")
	delete := esapi.DeleteRequest{
		DocumentID: id,
	}
	if index != nil {
		delete.Index = *index
	}
	deleteResponse, err := delete.Do(context.Background(), c)
	if err != nil {
		l.Errorf("error deleting work: %v", err)
		return err
	}
	if deleteResponse.StatusCode != 200 {
		bd, err := ioutil.ReadAll(deleteResponse.Body)
		if err != nil {
			l.Errorf("error reading response body: %v", err)
			return err
		}
		l.Errorf("error deleting work: %v", string(bd))
		return errors.New("error deleting work")
	}
	return nil
}

func deleteList(c *elasticsearch8.Client, index *string, id []string) error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "delete",
	})
	l.Debug("Deleting work")
	for _, i := range id {
		if err := deleteDocument(c, index, i); err != nil {
			return err
		}
	}
	return nil
}

func put(c *elasticsearch8.Client, index *string, id string, source string) error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "put",
	})
	l.Debug("Putting work")
	put := esapi.IndexRequest{
		DocumentID: id,
	}
	if index != nil {
		put.Index = *index
	}
	dd, err := json.Marshal(source)
	if err != nil {
		l.Errorf("error marshalling source: %v", err)
		return err
	}
	source = schema.ReplaceParamsString(dd, source)
	put.Body = strings.NewReader(source)
	putResponse, err := put.Do(context.Background(), c)
	if err != nil {
		l.Errorf("error putting work: %v", err)
		return err
	}
	if putResponse.StatusCode != 200 && putResponse.StatusCode != 201 {
		bd, err := ioutil.ReadAll(putResponse.Body)
		if err != nil {
			l.Errorf("error reading response body: %v", err)
			return err
		}
		l.Errorf("error putting work: %v", string(bd))
		return errors.New("error putting work")
	}
	return nil
}

func mergePut(c *elasticsearch8.Client, index *string, id string, query string, source any) error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "mergePut",
	})
	l.Debug("Merging and putting work")
	s := source.(map[string]interface{})
	delete(s, "_id")
	md, err := mergeStringAndAny(query, source)
	if err != nil {
		l.Errorf("error merging work: %v", err)
		return err
	}
	doc := `{"doc":` + md + `}`
	l.Debugf("merged work: %v", doc)
	put := esapi.UpdateRequest{
		DocumentID: id,
		Body:       strings.NewReader(doc),
	}
	if index != nil {
		put.Index = *index
	}
	putResponse, err := put.Do(context.Background(), c)
	if err != nil {
		l.Errorf("error putting work: %v", err)
		return err
	}
	if putResponse.StatusCode != 200 && putResponse.StatusCode != 201 {
		bd, err := ioutil.ReadAll(putResponse.Body)
		if err != nil {
			l.Errorf("error reading response body: %v", err)
			return err
		}
		l.Errorf("error putting work: %v", string(bd))
		return errors.New("error putting work")
	}
	return nil
}

func mergePutList(c *elasticsearch8.Client, origIndex *string, index *string, query string, source []any) error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "mergePutList",
	})
	l.Debug("Merging and putting work")
	jd, err := json.Marshal(source)
	if err != nil {
		l.Errorf("error marshalling source: %v", err)
		return err
	}
	query = schema.ReplaceParamsString(jd, query)
	for _, s := range source {
		ss := s.(map[string]interface{})
		id := ss["_id"].(string)
		if err := mergePut(c, index, id, query, s); err != nil {
			return err
		}
	}
	return nil
}

func move(c *elasticsearch8.Client, index *string, id string, newIndex *string, source any) error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "move",
	})
	l.Debug("Moving work")
	// remove _id field from source
	s := source.(map[string]interface{})
	delete(s, "_id")
	jd, err := json.Marshal(s)
	if err != nil {
		l.Errorf("error marshalling source: %v", err)
		return err
	}
	new := esapi.IndexRequest{
		DocumentID: id,
		Body:       bytes.NewReader(jd),
	}
	if newIndex != nil {
		new.Index = *newIndex
	} else {
		new.Index = *index
	}
	putRes, err := new.Do(context.Background(), c)
	if err != nil {
		l.Errorf("error putting work: %v", err)
		return err
	}
	if putRes.StatusCode != 200 && putRes.StatusCode != 201 {
		bd, err := ioutil.ReadAll(putRes.Body)
		if err != nil {
			l.Errorf("error reading response body: %v", err)
			return err
		}
		l.Errorf("error putting work: %v", string(bd))
		return errors.New("error putting work")
	}
	if err := deleteDocument(c, index, id); err != nil {
		l.Errorf("error deleting work: %v", err)
		return err
	}
	return nil
}

func moveList(c *elasticsearch8.Client, index *string, newIndex *string, sources []any) error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "moveList",
	})
	l.Debug("Moving work")
	for _, i := range sources {
		s := i.(map[string]interface{})
		id := s["_id"].(string)
		if err := move(c, index, id, newIndex, i); err != nil {
			return err
		}
	}
	return nil
}

func (d *Elasticsearch) getKeys() []string {
	var ks []string
	for _, v := range d.data {
		s := v.(map[string]interface{})
		if v, ok := s["_id"]; ok {
			ks = append(ks, v.(string))
		}
	}
	return ks
}

func (d *Elasticsearch) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from elasticsearch")
	if d.ClearIndex == nil || *d.ClearIndex == "" {
		d.ClearIndex = d.RetrieveIndex
	}
	switch d.ClearOp {
	case CloseOpDelete:
		return deleteList(d.Client, d.ClearIndex, d.getKeys())
	case CloseOpPut:
		var k string
		if d.Key != nil && *d.Key != "" {
			k = *d.Key
		}
		return put(d.Client, d.ClearIndex, k, d.ClearQuery)
	case CloseOpMergePut:
		return mergePutList(d.Client, d.RetrieveIndex, d.ClearIndex, d.ClearQuery, d.data)
	case CloseOpMove:
		return moveList(d.Client, d.RetrieveIndex, d.ClearIndex, d.data)
	}
	l.Debug("Cleared work")
	return nil
}

func (d *Elasticsearch) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure")
	if d.FailIndex == nil || *d.FailIndex == "" {
		d.FailIndex = d.RetrieveIndex
	}
	switch d.FailOp {
	case CloseOpDelete:
		return deleteDocument(d.Client, d.FailIndex, *d.Key)
	case CloseOpPut:
		var k string
		if d.Key != nil && *d.Key != "" {
			k = *d.Key
		}
		return put(d.Client, d.FailIndex, k, d.FailQuery)
	case CloseOpMergePut:
		return mergePutList(d.Client, d.RetrieveIndex, d.FailIndex, d.FailQuery, d.data)
	case CloseOpMove:
		return moveList(d.Client, d.RetrieveIndex, d.ClearIndex, d.data)
	}
	l.Debug("Handled failure")
	return nil
}

func (d *Elasticsearch) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "elasticsearch",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up")
	l.Debug("Cleaned up")
	return nil
}
