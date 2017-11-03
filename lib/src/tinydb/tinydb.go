// tinydb project tinydb.go
package tinydb

import (
	//	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/tealeg/xlsx"

	_ "github.com/go-sql-driver/mysql"
	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
)

func OpenDb(timeout time.Duration, driver, dsn string, maxOpen, maxIdle int) (*sql.DB, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	mydb, err := _openDb(driver, dsn, maxOpen, maxIdle)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return mydb, err
}

func _openDb(driver, dsn string, maxOpen, maxIdle int) (*sql.DB, error) {
	mydb, err := sql.Open(driver, dsn)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	mydb.SetMaxOpenConns(maxOpen)
	mydb.SetMaxIdleConns(maxIdle)

	err = mydb.Ping()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return mydb, err
}

//for insert update delete
func ModifyTab(timeout time.Duration, mydb *sql.DB, strsql string, args ...interface{}) (RowsAffected int64, ok bool) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	rowCount, ok := _modifyTab(mydb, strsql, args...)

	select {
	case <-ctx.Done():
		return -1, false
	default:
	}
	return rowCount, ok
}

func _modifyTab(mydb *sql.DB, strsql string, args ...interface{}) (RowsAffected int64, ok bool) {
	rst, err := mydb.Exec(strsql, args...)
	if err != nil {
		return 0, false
	}

	rowCount, _ := rst.RowsAffected()
	return rowCount, true
}

func Sql2Writer(timeout time.Duration, mydb *sql.DB, strSql string, w io.Writer, dataType string) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)

	var err error
	switch strings.ToLower(dataType) {
	case "xls":
		fallthrough
	case "xlsx":
		err = _xlsx2Writer(ctx, mydb, strSql, w)
	case "json":
		fallthrough
	default:
		err = _json2Writer(ctx, mydb, strSql, w)
	}
	return err
}

func _json2Writer(ctx context.Context, mydb *sql.DB, strSql string, w io.Writer) error {
	rows, err := mydb.Query(strSql)
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
	w.Write([]byte("["))
	rowCnt := 0

	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		if rowCnt > 0 {
			w.Write([]byte(","))
		}
		rowCnt += 1
		w.Write([]byte("{"))

		var strVal string
		for i, col := range values {
			if !col.Valid {
				strVal = "null"
			} else {
				jitem.Item = col.String
				bs, _ := json.Marshal(&jitem)
				strVal = string(bs[6 : len(bs)-2])
			}

			if i > 0 {
				w.Write([]byte(","))
			}
			w.Write([]byte(fmt.Sprintf(`"%v":"%v"`, columns[i], strVal)))
		}
		w.Write([]byte("}"))
	}

	w.Write([]byte("]"))
	return nil
}

/*
func SQL2Json(timeout time.Duration, mydb *sql.DB, strSql string, out_buf *bytes.Buffer) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	return _sql2Json(ctx, mydb, strSql, out_buf)
}

func _sql2Json(ctx context.Context, mydb *sql.DB, strSql string, out_buf *bytes.Buffer) error {
	if "" == strings.Trim(strSql, " ") {
		return errors.New("err msg")
	}

	rows, err := mydb.Query(strSql)
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
			fmt.Println(err.Error())
			return err
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

func Sql2Xlsx(timeout time.Duration, mydb *sql.DB, strSql string, w io.Writer) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	return _sql2Xlsx(ctx, mydb, strSql, w)
}
*/
func addRow2Sheet(s *xlsx.Sheet, args ...string) error {
	row := s.AddRow()
	cell := row.AddCell()
	cell.Value = ""

	for _, v := range args {
		cell := row.AddCell()
		cell.Value = v
	}

	return nil
}

func _xlsx2Writer(ctx context.Context, mydb *sql.DB, strSql string, w io.Writer) error {
	if "" == strings.Trim(strSql, " ") {
		return errors.New("err msg")
	}

	f := xlsx.NewFile()
	sheet, err := f.AddSheet("GoSheet1")
	if err != nil {
		return err
	}

	rows, err := mydb.Query(strSql)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()

	if err != nil {
		return err
	}

	addRow2Sheet(sheet, columns[0:]...)

	values := make([]sql.NullString, len(columns))
	scans := make([]interface{}, len(columns))
	cv := make([]string, len(columns))

	for i := range values {
		scans[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			return err
		}

		var strVal string
		for i, col := range values {
			if !col.Valid {
				strVal = ""
			} else {
				strVal = col.String
			}
			cv[i] = strVal
		}

		addRow2Sheet(sheet, cv[0:]...)
	}

	if err != nil {
		return err
	}

	if w != nil {
		err = f.Write(w)
		if err != nil {
			fmt.Println(err)
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}
