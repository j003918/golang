// godbs project godbs.go
package godbs

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
		ReadTimeout:    120 * time.Second,
		WriteTimeout:   120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	this.initService()

	return true
}

/*
func (this *GoDBS) SetDBS(db *sql.DB, srv *http.Server) {
	this.db = db
	this.srv = srv
	this.initService()
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

	for k, _ := range r.Form {
		strSql = strings.Replace(strSql, "#"+k+"#", r.Form.Get(k), -1)
	}

	w.Header().Set("Connection", "close")
	w.Header().Set("CharacterEncoding", "utf-8")

	var buf bytes.Buffer
	var err error
	switch strings.ToLower(strDT) {
	case "xlsx", "xls":
		err = this.Query2Xlsx(120, &buf, strSql)
		if err == nil {
			w.Header().Set("Content-Type", "application/vnd.ms-excel")
			w.Header().Set("Content-Disposition", "attachment;filename="+r.FormValue("sn")+".xlsx")
		}
	default:
		err = this.Query2Json(120, &buf, strSql)
		if err == nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0")
		}
	}

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && buf.Len() >= 1024*50 {
		var gzbuf bytes.Buffer
		gz := gzip.NewWriter(&gzbuf)
		_, err = gz.Write(buf.Bytes())
		gz.Close()
		if err == nil {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Length", strconv.Itoa(gzbuf.Len()))
			w.Write(gzbuf.Bytes())
		} else {
			fmt.Println(err.Error())
			w.Write(buf.Bytes())
		}
	} else {
		w.Write(buf.Bytes())
	}
}

func Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, handler)
}

func (this *GoDBS) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (this *GoDBS) Run(withTLS bool) {
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

	go func() {
		this.loadService()
		if withTLS {
			fmt.Println(this.srv.ListenAndServeTLS("./ca/ca.crt", "./ca/ca.key"))
		} else {
			fmt.Println(this.srv.ListenAndServe())
		}
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
