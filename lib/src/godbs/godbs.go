// godbs project godbs.go
package godbs

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type GoDBS struct {
	db         *sql.DB
	srv        *http.Server
	mapService sync.Map
}

func NewGoDBS() *GoDBS {
	return &GoDBS{
		db:  nil,
		srv: nil,
	}
}

func (this *GoDBS) InitDBS(db_driver, db_dsn string, db_maxOpen, db_maxIdle int, httpAddr string) bool {
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

	return true
}

/*
func (this *GoDBS) SetDBS(db *sql.DB, srv *http.Server) {
	this.db = db
	this.srv = srv
	this.initConf()
}
*/

func (this *GoDBS) dbs(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("CharacterEncoding", "utf-8")

	var buf bytes.Buffer
	var err error
	switch strings.ToLower(strDT) {
	case "xls":
		fallthrough
	case "xlsx":
		err = this.Query2Xlsx(120, &buf, strSql)
		if err == nil {
			w.Header().Set("Content-Type", "application/vnd.ms-excel")
			w.Header().Set("Content-Disposition", "attachment;filename="+r.FormValue("sn")+".xlsx")
		}
	default:
		err = this.Query2Json(120, &buf, strSql)
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

func Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, handler)
}

func (this *GoDBS) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (this *GoDBS) RunHttp() {
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

	this.initService()
	this.loadService()

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
