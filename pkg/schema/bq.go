package schema

import (
	"cloud.google.com/go/bigquery"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

func BqRowToMap(qry *bigquery.RowIterator) (map[string]bigquery.Value, error) {
	l := log.WithFields(log.Fields{
		"pkg": "sqlquery",
		"fn":  "CqlRowsToMap",
	})
	l.Debug("Converting row to map")
	m := make(map[string]bigquery.Value)
	for {
		err := qry.Next(&m)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	l.Debug("Converted row to map")
	return m, nil
}
