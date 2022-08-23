package schema

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type SqlQuery struct {
	Query  string `json:"query"`
	Params []any  `json:"params"`
}

func RowsToMap(rows *sql.Rows) (map[string]any, error) {
	l := log.WithFields(log.Fields{
		"pkg": "sqlquery",
		"fn":  "RowsToMap",
	})
	l.Debug("Converting row to map")
	m := make(map[string]interface{})
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			l.Error(err)
			return nil, err
		}
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			l.Error(err)
			return nil, err
		}
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
			l.Debugf("%s: %s", colName, *val)
		}
	}
	l.Debug("Converted row to map")
	return m, nil
}

func HandleField(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

func MapStringAnyToJSON(m map[string]any) ([]byte, error) {
	l := log.WithFields(log.Fields{
		"pkg": "schema",
		"fn":  "MapStringAnyToJSON",
	})
	l.Debug("Converting map to JSON")
	for k, v := range m {
		m[k] = HandleField(v)
	}
	jd, err := json.Marshal(m)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	return jd, nil
}

func SliceMapStringAnyToJSON(m []map[string]any) ([]byte, error) {
	l := log.WithFields(log.Fields{
		"pkg": "schema",
		"fn":  "SliceMapStringAnyToJSON",
	})
	l.Debug("Converting map to JSON")
	for k, v := range m {
		for kk, vv := range v {
			m[k][kk] = HandleField(vv)
		}
	}
	jd, err := json.Marshal(m)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	return jd, nil
}

func ExtractMustacheKey(s string) string {
	l := log.WithFields(log.Fields{
		"pkg": "schema",
		"fn":  "ExtractMustacheKey",
	})
	l.Debug("Extracting mustache key")
	var key string
	for _, k := range strings.Split(s, "{{") {
		if strings.Contains(k, "}}") {
			key = strings.Split(k, "}}")[0]
			break
		}
	}
	return key
}

func ExtractMustacheKeys(s string) []string {
	l := log.WithFields(log.Fields{
		"pkg": "schema",
		"fn":  "ExtractMustacheKeys",
	})
	l.Debug("Extracting mustache keys")
	keys := []string{}
	for _, k := range strings.Split(s, "{{") {
		if strings.Contains(k, "}}") {
			keys = append(keys, strings.Split(k, "}}")[0])
		}
	}
	l.Debug("Extracted mustache keys: ", keys)
	return keys
}

func ReplaceParams(bd []byte, params []any) []any {
	for i, v := range params {
		sv := fmt.Sprintf("%s", v)
		if sv == "{{procx_payload}}" {
			params[i] = bd
		} else if strings.Contains(sv, "{{") {
			key := ExtractMustacheKey(sv)
			params[i] = gjson.GetBytes(bd, key).String()
		}
	}
	return params
}

func ReplaceParamsMap(data map[string]any, params []any) []any {
	for i, v := range params {
		sv := fmt.Sprintf("%s", v)
		if sv == "{{procx_payload}}" {
			jd, err := json.Marshal(data)
			if err != nil {
				log.Error(err)
			}
			params[i] = jd
		} else if strings.Contains(sv, "{{") {
			key := ExtractMustacheKey(sv)
			params[i] = data[key]
		}
	}
	return params
}

func ReplaceJSONKey(query string, k string, v any) string {
	l := log.WithFields(log.Fields{
		"pkg": "schema",
		"fn":  "ReplaceJSONKey",
		"k":   k,
		"v":   v,
	})
	l.Debug("Replacing JSON key")
	return strings.ReplaceAll(query, "{{"+k+"}}", fmt.Sprint(HandleField(v)))
}

func ReplaceParamsString(bd []byte, params string) string {
	l := log.WithFields(log.Fields{
		"pkg":    "schema",
		"fn":     "ReplaceParamsString",
		"params": params,
	})
	l.Debug("Replacing params")
	s := strings.ReplaceAll(params, "{{procx_payload}}", string(bd))
	keys := ExtractMustacheKeys(s)
	for _, k := range keys {
		jv := gjson.GetBytes(bd, k)
		s = ReplaceJSONKey(s, k, jv.String())
	}
	l.Debug("Replaced params")
	return s
}

func ReplaceParamsMapString(data map[string]any, params string) string {
	l := log.WithFields(log.Fields{
		"pkg":    "schema",
		"fn":     "ReplaceParamsMapString",
		"params": params,
	})
	l.Debug("Replacing params map string")
	jd, err := MapStringAnyToJSON(data)
	if err != nil {
		l.Error(err)
	}
	s := strings.ReplaceAll(params, "{{procx_payload}}", string(jd))
	keys := ExtractMustacheKeys(s)
	for _, k := range keys {
		jv := gjson.GetBytes(jd, k)
		s = ReplaceJSONKey(s, k, jv.String())
	}
	l.Debug("Replaced params map string: ", s)
	return s
}

func ReplaceParamsSliceMapString(data []map[string]any, params string) string {
	l := log.WithFields(log.Fields{
		"pkg":    "schema",
		"fn":     "ReplaceParamsSliceMapString",
		"params": params,
	})
	l.Debug("Replacing params map string")
	jd, err := SliceMapStringAnyToJSON(data)
	if err != nil {
		l.Error(err)
	}
	s := strings.ReplaceAll(params, "{{procx_payload}}", string(jd))
	keys := ExtractMustacheKeys(s)
	for _, k := range keys {
		jv := gjson.GetBytes(jd, k)
		s = ReplaceJSONKey(s, k, jv.String())
	}
	l.Debug("Replaced params map string: ", s)
	return s
}
