package schema

import (
	"cloud.google.com/go/bigquery"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

func BqRowsToMap(qry *bigquery.RowIterator) ([]map[string]bigquery.Value, error) {
	l := log.WithFields(log.Fields{
		"pkg": "sqlquery",
		"fn":  "BqRowsToMap",
	})
	l.Debug("Converting rows to map")
	var data []map[string]bigquery.Value
	for {
		var row map[string]bigquery.Value
		err := qry.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		data = append(data, row)
	}
	l.Debug("Converted rows to map")
	return data, nil
}
