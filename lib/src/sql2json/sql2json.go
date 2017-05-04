// sql2json project sql2json.go
package sql2json

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func GetJson(ctx context.Context, db *sql.DB, strSql string, out_buf *bytes.Buffer) error {
	if "" == strings.Trim(strSql, " ") {
		return errors.New("err msg")
	}

	rows, err := db.Query(strSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()

	if err != nil {
		return err
	}

	//fix bug time.Time nil
	//values := make([]sql.RawBytes, len(columns))
	values := make([]sql.NullString, len(columns))
	scans := make([]interface{}, len(columns))

	for i := range values {
		scans[i] = &values[i]
	}

	type Jitem struct {
		Item string `json:"e"`
	}
	var jitem Jitem

	out_buf.WriteByte('[')
	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			panic(err.Error())
		}

		out_buf.WriteByte('{')
		var strVal string
		for i, col := range values {
			if !col.Valid {
				strVal = "null"
			} else {
				jitem.Item = col.String
				bs, _ := json.Marshal(&jitem)
				strVal = string(bs[6 : len(bs)-2])
			}

			columName := strings.ToLower(columns[i])
			cell := fmt.Sprintf(`"%v":"%v"`, columName, strVal)
			out_buf.WriteString(cell + ",")
		}
		out_buf.Bytes()[out_buf.Len()-1] = '}'
		out_buf.WriteByte(',')
	}
	if out_buf.Len() > 1 {
		out_buf.Bytes()[out_buf.Len()-1] = ']'
	} else {
		out_buf.WriteByte(']')
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}
