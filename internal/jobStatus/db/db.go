package db

import (
	"fmt"
	"time"

	"go-slo/internal"
)

func QueryToWhere(q map[string]string) (string, []any, error) {
	var vals []any
	where := ""
	fieldNum := 1

	for nm, val := range q {
		dbVal, err := FieldToDb(nm, val)
		if err != nil {
			return "", nil, err
		}
		if fieldNum > 1 {
			where += " AND "
		}
		where += fmt.Sprintf(`"%s" = $%d`, nm, fieldNum)
		vals = append(vals, dbVal)
		fieldNum++
	}
	return where, vals, nil
}

// fieldToDb converts string values from the query map to database values
// we can use to run queries.
func FieldToDb(nm string, val string) (any, error) {
	// explicitly convert fields that are not string
	switch nm {
	case "JobStatusTimestamp": // time.Time
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		return t.UTC(), nil
	case "BusinessDate": // internal.Date
		dt, err := internal.NewDate(val)
		if err != nil {
			return nil, err
		}
		return dt.AsTime(), nil
	default: // all the rest are strings
		return val, nil
	}
}
