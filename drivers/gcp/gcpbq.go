package gcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/robertlestak/procx/pkg/flags"
	"google.golang.org/api/iterator"

	log "github.com/sirupsen/logrus"
)

type BQ struct {
	Client        *bigquery.Client
	ProjectID     string
	Key           *string
	QueryKey      *bool
	RetrieveQuery *string
	ClearQuery    *string
	FailQuery     *string
}

func (d *BQ) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "bq",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment variables")
	if os.Getenv(prefix+"GCP_BQ_RETRIEVE_QUERY") != "" {
		q := os.Getenv(prefix + "GCP_BQ_RETRIEVE_QUERY")
		d.RetrieveQuery = &q
	}
	if os.Getenv(prefix+"GCP_BQ_CLEAR_QUERY") != "" {
		q := os.Getenv(prefix + "GCP_BQ_CLEAR_QUERY")
		d.ClearQuery = &q
	}
	if os.Getenv(prefix+"GCP_BQ_FAIL_QUERY") != "" {
		q := os.Getenv(prefix + "GCP_BQ_FAIL_QUERY")
		d.FailQuery = &q
	}
	if os.Getenv(prefix+"GCP_BQ_QUERY_KEY") != "" {
		tval := os.Getenv(prefix+"GCP_BQ_QUERY_KEY") == "true"
		d.QueryKey = &tval
	}
	if os.Getenv(prefix+"GCP_PROJECT_ID") != "" {
		d.ProjectID = os.Getenv(prefix + "GCP_PROJECT_ID")
	}
	return nil
}

func (d *BQ) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "bq",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.ProjectID = *flags.GCPProjectID
	if *flags.GCPBQQueryKey {
		d.QueryKey = flags.GCPBQQueryKey
	}
	if *flags.GCPBQRetrieveQuery != "" {
		d.RetrieveQuery = flags.GCPBQRetrieveQuery
	}
	if *flags.GCPBQClearQuery != "" {
		d.ClearQuery = flags.GCPBQClearQuery
	}
	if *flags.GCPBQFailQuery != "" {
		d.FailQuery = flags.GCPBQFailQuery
	}
	return nil
}

func (d *BQ) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "bq",
		"fn":  "Init",
		"prj": d.ProjectID,
	})
	l.Debug("Initializing GCP_BQ client")
	var err error
	ctx := context.Background()
	c, err := bigquery.NewClient(ctx, d.ProjectID)
	if err != nil {
		return err
	}
	d.Client = c
	l.Debug("Connected to bq")
	return nil
}

func (d *BQ) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "bq",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from GCP_BQ")
	var err error
	var result string
	var key string
	if d.RetrieveQuery == nil || *d.RetrieveQuery == "" {
		l.Error("RetrieveQuery is nil or empty")
		return nil, errors.New("RetrieveQuery is nil or empty")
	}
	qry := d.Client.Query(*d.RetrieveQuery)
	it, err := qry.Read(context.Background())
	if err != nil {
		l.Error(err)
		return nil, err
	}
	if d.QueryKey != nil && *d.QueryKey {
		for {
			var values []bigquery.Value
			err := it.Next(&values)
			if err == iterator.Done {
				break
			}
			if err != nil {
				l.Error(err)
				return nil, err
			}
			if values == nil {
				return nil, nil
			}
			if len(values) < 2 {
				return nil, errors.New("invalid query result")
			}
			if values[0] == nil || values[1] == nil {
				return nil, errors.New("invalid query result")
			}
			key = fmt.Sprintf("%v", values[0])
			result = fmt.Sprintf("%v", values[1])
			l.WithFields(log.Fields{
				"key":  key,
				"data": result,
			}).Debug("Got work")
			if key != "" && result != "" {
				break
			}
		}
	} else {
		for {
			var values []bigquery.Value
			err := it.Next(&values)
			if err == iterator.Done {
				break
			}
			if err != nil {
				l.Error(err)
				return nil, err
			}
			if values == nil {
				l.Debug("values is nil")
				return nil, nil
			}
			v := values[0]
			if v == nil {
				l.Debug("value is nil")
				return nil, errors.New("invalid query result")
			}
			// parse result as string
			result = fmt.Sprintf("%v", v)
			l.Debugf("result: %v", result)
			if result != "" {
				break
			}
		}
	}
	if result == "" {
		l.Debug("No work found")
		return nil, nil
	}
	d.Key = &key
	l.Debug("Got work")
	return strings.NewReader(result), nil
}

func (d *BQ) clearQuery() string {
	if d.ClearQuery == nil {
		return ""
	}
	return strings.ReplaceAll(*d.ClearQuery, "{{key}}", *d.Key)
}

func (d *BQ) failQuery() string {
	if d.FailQuery == nil {
		return ""
	}
	return strings.ReplaceAll(*d.FailQuery, "{{key}}", *d.Key)
}

func (d *BQ) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "bq",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from GCP_BQ")
	var err error
	if d.ClearQuery == nil || *d.ClearQuery == "" {
		return nil
	}
	qry := d.Client.Query(d.clearQuery())
	_, err = qry.Read(context.Background())
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleared work")
	return nil
}

func (d *BQ) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "bq",
		"fn":  "HandleFailure",
	})
	l.Debug("Handling failure in GCP_BQ")
	var err error
	if d.FailQuery == nil || *d.FailQuery == "" {
		return nil
	}
	qry := d.Client.Query(d.failQuery())
	_, err = qry.Read(context.Background())
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Handled failure")
	return nil
}

func (d *BQ) Cleanup() error {
	l := log.WithFields(log.Fields{
		"pkg": "bq",
		"fn":  "Cleanup",
	})
	l.Debug("Cleaning up GCP_BQ")
	err := d.Client.Close()
	if err != nil {
		l.Error(err)
		return err
	}
	l.Debug("Cleaned up")
	return nil
}
