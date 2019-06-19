// json
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"strings"
)

func rows2Json(rows *sql.Rows, out_buf *bytes.Buffer) bool {
	colKeys, err := rows.Columns()
	if err != nil {
		log.Println(err)
		return false
	}

	colVals := make([]sql.NullString, len(colKeys))
	colValsPtr := make([]interface{}, len(colKeys))
	var valBuf []byte

	for i, _ := range colKeys {
		colKeys[i] = strings.ToLower(colKeys[i])
		colValsPtr[i] = &colVals[i]
	}

	out_buf.WriteByte('[')
	for rows.Next() {
		err = rows.Scan(colValsPtr...)
		if err != nil {
			log.Println(err)
			return false
		}

		out_buf.WriteByte('{')
		for i, val := range colVals {
			valBuf, err = json.Marshal(&val.String)
			if err != nil {
				log.Println(err)
				return false
			}
			out_buf.WriteString(`"` + colKeys[i] + `":` + string(valBuf) + `,`)
		}
		out_buf.Bytes()[out_buf.Len()-1] = '}'
		out_buf.WriteByte(',')
	}

	if out_buf.Len() > 1 {
		out_buf.Bytes()[out_buf.Len()-1] = ']'
	} else {
		out_buf.WriteByte(']')
	}

	return true
}
