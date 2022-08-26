package schema

import (
	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"
)

func CqlRowsToMapSlice(qry *gocql.Query) ([]map[string]any, error) {
	l := log.WithFields(log.Fields{
		"pkg": "sqlquery",
		"fn":  "CqlRowsToMapSlice",
	})
	l.Debug("Converting row to map")
	iter := qry.Iter()
	return iter.SliceMap()
}
