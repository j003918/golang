// tinydb project tinydb.go
package tinydb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tealeg/xlsx"

	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
)

type Tinydb struct {
	mydb *sql.DB
}

func NewTinydb() *Tinydb {
	return &Tinydb{
		mydb: nil,
	}
}

func (td *Tinydb) Open(driver, dsn string, maxOpen, maxIdle int) error {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	var err error
	td.mydb, err = sql.Open(driver, dsn)
	if err != nil {
		return err
	}

	td.mydb.SetMaxOpenConns(maxOpen)
	td.mydb.SetMaxIdleConns(maxIdle)

	err = td.mydb.Ping()
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (td *Tinydb) Query(timeout time.Duration, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)

	rows, err := td.mydb.Query(query, args...)
	//rows, err := td.mydb.Query(query)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return rows, err
}

func (td *Tinydb) Exec(timeout time.Duration, strsql string, args ...interface{}) (RowsAffected int64, ok bool) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)

	rst, err := td.mydb.Exec(strsql, args...)
	if err != nil {
		return 0, false
	}

	rowCount, err := rst.RowsAffected()
	if err != nil {
		return -1, false
	}

	select {
	case <-ctx.Done():
		return -2, false
	default:
	}
	return rowCount, true
}

func (td *Tinydb) Query2Json(timeout time.Duration, buf *bytes.Buffer, query string, args ...interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	rows, err := td.mydb.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	buf.Reset()

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
	buf.WriteByte('[')
	rowCnt := 0

	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		if rowCnt > 0 {
			buf.WriteByte(',')
		}
		rowCnt += 1
		buf.WriteByte('{')

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
				buf.WriteByte(',')
			}
			buf.WriteString(fmt.Sprintf(`"%v":"%v"`, columns[i], strVal))
		}
		buf.WriteByte('}')
	}
	buf.WriteByte(']')

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (td *Tinydb) addRow2Sheet(s *xlsx.Sheet, args ...string) {
	row := s.AddRow()
	cell := row.AddCell()
	cell.Value = ""

	for _, v := range args {
		cell := row.AddCell()
		cell.Value = v
	}
}

func (td *Tinydb) Query2Xlsx(timeout time.Duration, buf *bytes.Buffer, query string, args ...interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	rows, err := td.mydb.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	buf.Reset()

	f := xlsx.NewFile()
	sheet, err := f.AddSheet("Sheet1")
	if err != nil {
		return err
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	td.addRow2Sheet(sheet, columns[0:]...)

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
		td.addRow2Sheet(sheet, cv[0:]...)
	}

	if err != nil {
		return err
	}

	err = f.Write(buf)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return err
}

/*********************************************************************************************/
/*
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

func _xlsx2Writer(ctx context.Context, mydb *sql.DB, strSql string, w io.Writer) error {
	if "" == strings.Trim(strSql, " ") {
		return errors.New("err msg")
	}

	f := xlsx.NewFile()
	sheet, err := f.AddSheet("Sheet1")
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
*/
