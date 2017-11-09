// godbs project godbs.go
package godbs

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tealeg/xlsx"

	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
)

var (
	sql_godbs_user = `		
		CREATE TABLE IF NOT EXISTS godbs_user 
		(
			id 			VARCHAR(64) PRIMARY KEY NOT NULL,     
    		pass		VARCHAR(128) NOT NULL, 
			#sign		VARCHAR(32) NOT NULL DEFAULT "", 
    		#trustzone 	VARCHAR(512) NOT NULL DEFAULT "*", 
    		status		INTEGER NOT NULL DEFAULT 0, 
    		#login_time	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    		create_time	TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`

	sql_godbs_dsn = `		
		CREATE TABLE IF NOT EXISTS godbs_dsn 
		(
			id		INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    		driver	VARCHAR(64) NOT NULL, 
    		dsn		VARCHAR(1024) NOT NULL,
			status 	INTEGER NOT NULL DEFAULT 0,
			info	VARCHAR(128) NOT NULL DEFAULT "",
			#update_time	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    		create_time	TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`

	sql_godbs_service = `		
		CREATE TABLE IF NOT EXISTS godbs_service 
		(
			sn		VARCHAR(128) PRIMARY KEY NOT NULL, 
    		content		VARCHAR(4096) NOT NULL, 
    		dsn_id		INTEGER NOT NULL DEFAULT -1,
			status 		INTEGER NOT NULL DEFAULT 0, 
    		#update_time	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    		create_time	TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`

	sql_godbs_service_test = `insert into godbs_service (sn,content) values('test','select now() as server_time')`
)

func dbOpen(driver, dsn string, maxOpen, maxIdle int) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)

	err = db.Ping()
	if err != nil {
		return db, err
	}

	return db, err
}

func Query2Json(outBuf *bytes.Buffer, db *sql.DB, strSql string, args ...interface{}) error {
	rows, err := db.Query(strSql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	outBuf.Reset()

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
	outBuf.WriteByte('[')

	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		outBuf.WriteByte('{')
		for i, col := range values {
			jitem.Item = col.String
			bs, _ := json.Marshal(&jitem)
			outBuf.WriteString(fmt.Sprintf(`"%v":"%v",`, columns[i], string(bs[6:len(bs)-2])))
		}

		outBuf.Bytes()[outBuf.Len()-1] = '}'
		outBuf.WriteByte(',')
	}

	if outBuf.Len() > 1 {
		outBuf.Bytes()[outBuf.Len()-1] = ']'
	} else {
		outBuf.WriteByte(']')
	}

	return nil
}

func addRow2Sheet(s *xlsx.Sheet, args ...string) {
	row := s.AddRow()
	cell := row.AddCell()
	cell.Value = ""

	for _, v := range args {
		cell := row.AddCell()
		cell.Value = v
	}
}

func Query2Xlsx(outBuf *bytes.Buffer, db *sql.DB, strSql string, args ...interface{}) error {
	rows, err := db.Query(strSql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	outBuf.Reset()

	f := xlsx.NewFile()
	sheet, err := f.AddSheet("Sheet1")
	if err != nil {
		return err
	}

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

		for i, col := range values {
			cv[i] = col.String
		}
		addRow2Sheet(sheet, cv[0:]...)
	}

	if err != nil {
		return err
	}

	err = f.Write(outBuf)
	return err
}
