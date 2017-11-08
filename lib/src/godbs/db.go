// godbs project godbs.go
package godbs

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tealeg/xlsx"

	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
)

func (this *GoDBS) opendb(driver, dsn string, maxOpen, maxIdle int) error {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	var err error
	this.db, err = sql.Open(driver, dsn)
	if err != nil {
		return err
	}

	this.db.SetMaxOpenConns(maxOpen)
	this.db.SetMaxIdleConns(maxIdle)

	err = this.db.Ping()
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

func (this *GoDBS) Query(timeout time.Duration, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)

	rows, err := this.db.Query(query, args...)
	//rows, err := td.mydb.Query(query)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return rows, err
}

func (this *GoDBS) Exec(timeout time.Duration, strsql string, args ...interface{}) (RowsAffected int64, ok bool) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)

	rst, err := this.db.Exec(strsql, args...)
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

func (this *GoDBS) Query2Json(timeout time.Duration, buf *bytes.Buffer, query string, args ...interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	rows, err := this.db.Query(query, args...)
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

	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		buf.WriteByte('{')
		for i, col := range values {
			jitem.Item = col.String
			bs, _ := json.Marshal(&jitem)
			buf.WriteString(fmt.Sprintf(`"%v":"%v",`, columns[i], string(bs[6:len(bs)-2])))
		}

		buf.Bytes()[buf.Len()-1] = '}'
		buf.WriteByte(',')
	}

	if buf.Len() > 1 {
		buf.Bytes()[buf.Len()-1] = ']'
	} else {
		buf.WriteByte(']')
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (td *GoDBS) addRow2Sheet(s *xlsx.Sheet, args ...string) {
	row := s.AddRow()
	cell := row.AddCell()
	cell.Value = ""

	for _, v := range args {
		cell := row.AddCell()
		cell.Value = v
	}
}

func (this *GoDBS) Query2Xlsx(timeout time.Duration, buf *bytes.Buffer, query string, args ...interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	rows, err := this.db.Query(query, args...)
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
	this.addRow2Sheet(sheet, columns[0:]...)

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

		for i, col := range values {
			cv[i] = col.String
		}
		this.addRow2Sheet(sheet, cv[0:]...)
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

func (this *GoDBS) initService() {
	strSql := `		
		CREATE TABLE IF NOT EXISTS godbs 
		(
			sn			VARCHAR(64)		PRIMARY KEY NOT NULL,     
    		content		VARCHAR(4096) 	NOT NULL, 
			name		VARCHAR(128) 	DEFAULT NULL, 
    		create_time	TIMESTAMP		NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`

	this.Exec(30, strSql)
	this.Exec(30, `insert into godbs (sn,content,name) values('test','select * from information_schema.columns where table_name=''#tn#'' ','测试')`)
}

func (this *GoDBS) loadService() {
	strSql := "select sn,content from godbs"
	rows, err := this.Query(10, strSql)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	strSN, strContent := "", ""
	for rows.Next() {
		err = rows.Scan(&strSN, &strContent)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		this.mapService.Store(strings.ToLower(strSN), strContent)
	}
}
