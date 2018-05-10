// dbs project dbs.go
package dbs

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/tealeg/xlsx"
)

func mssqlInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func Rows2Json(rows *sql.Rows, out_buf *bytes.Buffer) error {
	colKeys, err := rows.Columns()
	if err != nil {
		return err
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
			return nil
		}

		out_buf.WriteByte('{')
		for i, val := range colVals {
			valBuf, err = json.Marshal(&val.String)
			if err != nil {
				return err
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

	return nil
}

func Rows2Xlsx(rows *sql.Rows, out_buf *bytes.Buffer) error {
	colKeys, err := rows.Columns()
	if err != nil {
		return err
	}

	colVals := make([]sql.NullString, len(colKeys))
	colValsPtr := make([]interface{}, len(colKeys))

	cells := make([]string, len(colKeys))
	tmp := xlsx.NewFile()
	sheet, err := tmp.AddSheet("sheet1")
	if err != nil {
		return err
	}

	for i, _ := range colKeys {
		colKeys[i] = strings.ToLower(colKeys[i])
		colValsPtr[i] = &colVals[i]
	}

	sheet.AddRow().WriteSlice(&colKeys, -1)
	for rows.Next() {
		err = rows.Scan(colValsPtr...)
		if err != nil {
			return nil
		}

		for i, val := range colVals {
			cells[i] = val.String
		}
		sheet.AddRow().WriteSlice(&cells, -1)
	}

	//tmp.Save("bb.xlsx")
	tmp.Write(out_buf)

	return nil
}
