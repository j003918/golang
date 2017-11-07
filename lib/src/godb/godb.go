// godb project godb.go
package godb

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tealeg/xlsx"

	//_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
)

type GoDB struct {
	db         *sql.DB
	srv        *http.Server
	mapService sync.Map
}

func NewGoDB() *GoDB {
	return &GoDB{
		db:  nil,
		srv: nil,
	}
}

func (this *GoDB) Init(db_driver, db_dsn string, db_maxOpen, db_maxIdle int, httpAddr string) bool {
	err := this.opendb(db_driver, db_dsn, db_maxOpen, db_maxIdle)
	if err != nil {
		return false
	}

	this.srv = &http.Server{
		Addr:           httpAddr,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	this.initConf()
	return true
}

func (this *GoDB) Set(db *sql.DB, srv *http.Server) {
	this.db = db
	this.srv = srv
	this.initConf()
}

func (this *GoDB) opendb(driver, dsn string, maxOpen, maxIdle int) error {
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

func (this *GoDB) Query(timeout time.Duration, query string, args ...interface{}) (*sql.Rows, error) {
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

func (this *GoDB) Exec(timeout time.Duration, strsql string, args ...interface{}) (RowsAffected int64, ok bool) {
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

func (this *GoDB) Query2Json(timeout time.Duration, buf *bytes.Buffer, query string, args ...interface{}) error {
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

func (td *GoDB) addRow2Sheet(s *xlsx.Sheet, args ...string) {
	row := s.AddRow()
	cell := row.AddCell()
	cell.Value = ""

	for _, v := range args {
		cell := row.AddCell()
		cell.Value = v
	}
}

func (this *GoDB) Query2Xlsx(timeout time.Duration, buf *bytes.Buffer, query string, args ...interface{}) error {
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

		var strVal string
		for i, col := range values {
			if !col.Valid {
				strVal = ""
			} else {
				strVal = col.String
			}
			cv[i] = strVal
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

func (this *GoDB) initConf() {
	strSql := `		
		CREATE TABLE IF NOT EXISTS godbs 
		(
			sn			VARCHAR(64)		PRIMARY KEY NOT NULL,     
    		content		VARCHAR(4096) 	NOT NULL, 
			name		VARCHAR(128) 	DEFAULT NULL, 
    		create_time	TIMESTAMP		NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`

	//tinydb.ModifyTab(5, dbConn, strSql)
	this.Exec(5, strSql)
}

func (this *GoDB) loadService() {
	strSql := "select sn,content from godbs"
	//rows, err := dbConn.Query(strSql)
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

func (this *GoDB) dbs(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	strDT := r.FormValue("dt")
	strSqlObj, ok := this.mapService.Load(strings.ToLower(r.FormValue("sn")))
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte(http.StatusText(404)))
		return
	}
	strSql := strSqlObj.(string)
	/*
		for k, _ := range r.Form {
			strSql = strings.Replace(strSql, "<"+k+">", r.Form.Get(k), -1)
		}
	*/
	w.Header().Set("Connection", "close")
	w.Header().Set("CharacterEncoding", "utf-8")

	var buf bytes.Buffer
	var err error
	switch strings.ToLower(strDT) {
	case "xls":
		fallthrough
	case "xlsx":
		err = this.Query2Xlsx(30, &buf, strSql)
		if err == nil {
			w.Header().Set("Content-Type", "application/vnd.ms-excel") //application/vnd.ms-excel or application/x-xls
			w.Header().Set("Content-Disposition", "attachment;filename="+r.FormValue("sn")+".xlsx")
		}
	default:
		err = this.Query2Json(30, &buf, strSql)
		if err == nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
	}

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(http.StatusText(500)))
		return
	}
	w.Write(buf.Bytes())
}

func (this *GoDB) AddUrl(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	//http.Request
	http.HandleFunc(pattern, handler)
}
func (this *GoDB) RunHttp() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	http.Handle("/", http.FileServer(http.Dir("./static/")))
	http.HandleFunc("/dbs", this.dbs)

	http.HandleFunc("/dbs/sys/reload", func(w http.ResponseWriter, r *http.Request) {
		this.loadService()
	})

	http.HandleFunc("/dbs/sys/list", func(w http.ResponseWriter, r *http.Request) {
		this.mapService.Range(func(k, v interface{}) bool {
			w.Write([]byte(k.(string) + ": " + v.(string) + "\r\n"))
			return true
		})
	})

	http.HandleFunc("/dbs/sys/add", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		strSN := strings.ToLower(r.FormValue("sn"))
		strContent := r.FormValue("content")
		_, ok := this.mapService.LoadOrStore(strSN, strContent)

		if !ok {
			w.Write([]byte(http.StatusText(200)))
		} else {
			w.WriteHeader(500)
			w.Write([]byte(http.StatusText(500)))
		}
	})

	http.HandleFunc("/dbs/sys/del", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		strSN := strings.ToLower(r.FormValue("sn"))
		this.mapService.Delete(strSN)
	})

	// add ffxlc insert function
	//	http.HandleFunc("/gdi", gdi)
	//

	go func() {
		fmt.Println(this.srv.ListenAndServe())
	}()

	<-stopChan
	fmt.Println("Shutting down server...")

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	this.srv.Shutdown(ctx)

	if nil != ctx.Err() {
		fmt.Println(ctx.Err().Error())
	}

	fmt.Println("Exit OK!")
}
