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
	"github.com/robertlestak/procx/pkg/schema"

	log "github.com/sirupsen/logrus"
)

type BQ struct {
	Client        *bigquery.Client
	ProjectID     string
	Key           *string
	RetrieveField *string
	RetrieveQuery *string
	ClearQuery    *string
	FailQuery     *string
	data          map[string]bigquery.Value
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
	if os.Getenv(prefix+"GCP_BQ_RETRIEVE_FIELD") != "" {
		f := os.Getenv(prefix + "GCP_BQ_RETRIEVE_FIELD")
		d.RetrieveField = &f
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
	d.RetrieveField = flags.GCPBQRetrieveField
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
	var result string
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
	m, err := schema.BqRowToMap(it)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	if len(m) == 0 {
		l.Debug("No work found")
		return nil, nil
	}
	d.data = m
	if d.RetrieveField != nil && *d.RetrieveField != "" {
		result = fmt.Sprintf("%s", schema.HandleField(m[*d.RetrieveField]))
	} else {
		td := make(map[string]any)
		for k, v := range d.data {
			td[k] = v
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

func (d *BQ) clearQuery() string {
	if d.ClearQuery == nil {
		return ""
	}
	td := make(map[string]any)
	for k, v := range d.data {
		td[k] = v
	}
	return schema.ReplaceParamsMapString(td, *d.ClearQuery)
}

func (d *BQ) failQuery() string {
	if d.FailQuery == nil {
		return ""
	}
	td := make(map[string]any)
	for k, v := range d.data {
		td[k] = v
	}
	return schema.ReplaceParamsMapString(td, *d.FailQuery)
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
