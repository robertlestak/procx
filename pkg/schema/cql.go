package schema

import (
	"github.com/gocql/gocql"
	log "github.com/sirupsen/logrus"
)

func CqlRowsToMap(qry *gocql.Query) (map[string]any, error) {
	l := log.WithFields(log.Fields{
		"pkg": "sqlquery",
		"fn":  "CqlRowsToMap",
	})
	l.Debug("Converting row to map")
	m := make(map[string]interface{})
	iter := qry.Iter()
	cols := iter.Columns()
	columns := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}
	iter.Scan(columnPointers...)
	if err := iter.Close(); err != nil {
		l.Error(err)
		return nil, err
	}
	for i, colName := range cols {
		val := columnPointers[i].(*interface{})
		m[colName.Name] = *val
		l.Debugf("%s: %s", colName, *val)
	}
	l.Debug("Converted row to map")
	return m, nil
}

func CqlRowsToMapSlice(qry *gocql.Query) ([]map[string]any, error) {
	l := log.WithFields(log.Fields{
		"pkg": "sqlquery",
		"fn":  "CqlRowsToMap",
	})
	l.Debug("Converting row to map")
	var sm []map[string]any
	iter := qry.Iter()
	cols := iter.Columns()
	columns := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}
	for iter.Scan(columnPointers...) {
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName.Name] = *val
			l.Debugf("%s: %s", colName, *val)
		}
		sm = append(sm, m)
	}
	if err := iter.Close(); err != nil {
		l.Error(err)
		return nil, err
	}
	l.Debug("Converted row to map")
	return sm, nil
}
