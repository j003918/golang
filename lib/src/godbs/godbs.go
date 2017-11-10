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

var (
	dm = &dbsManager{}
)

type dbs struct {
	sn     string
	sql    string
	dbConn *sql.DB
}

type dbsManager struct {
	sysdb     *sql.DB
	mapDSN    sync.Map
	mapServie sync.Map
}

func (this *dbsManager) initDB(driver, dsn string, maxOpen, maxIdle int) error {
	db, err := dbOpen(driver, dsn, maxOpen, maxIdle)
	if err != nil {
		return err
	}

	this.mapDSN.Store(-1, db)
	this.sysdb = db

	db.Exec(sql_godbs_user)
	db.Exec(sql_godbs_dsn)
	db.Exec(sql_godbs_service)

	db.Exec(sql_godbs_service_test)
	return nil
}

func (this *dbsManager) delService(sn string) {
	this.mapServie.Delete(strings.ToLower(sn))
}

func (this *dbsManager) addService(sn, strSql string, dsnid int) bool {
	obj, ok := this.mapDSN.Load(dsnid)
	if !ok {
		fmt.Println("load", sn, "failure")
		return false
	}

	this.mapServie.Store(strings.ToLower(sn), &dbs{
		sn:     strings.ToLower(sn),
		sql:    strSql,
		dbConn: obj.(*sql.DB),
	})

	fmt.Println("load service:", sn)
	return true
}

func (this *dbsManager) getService(sn string) *dbs {
	obj, ok := this.mapServie.Load(strings.ToLower(sn))
	if !ok {
		return nil
	}
	return obj.(*dbs)
}

func (this *dbsManager) loadDSN() {
	strsql := "select id,driver,dsn,info from godbs_dsn"
	rows, err := this.sysdb.Query(strsql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		strDriver, strDSN, dsnid, info := "", "", 0, ""
		err := rows.Scan(&dsnid, &strDriver, &strDSN, &info)
		if err != nil {
			panic(err)
		}
		fmt.Println("find dsn", info)
		_, ok := this.mapDSN.Load(dsnid)
		if ok {
			continue
		}

		db, err := dbOpen(strDriver, strDSN, 0, 0)
		if err == nil {
			this.mapDSN.Store(dsnid, db)
			fmt.Println("load DSN", dsnid, info)
		} else {
			fmt.Println(err)
		}
	}
}

func (this *dbsManager) loadService() {
	strsql := "select sn,content,dsn_id from godbs_service"
	rows, err := this.sysdb.Query(strsql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		strSN, strContent, dsnid := "", "", 0
		err := rows.Scan(&strSN, &strContent, &dsnid)
		if err != nil {
			panic(err)
		}
		this.addService(strSN, strContent, dsnid)
	}
}

func (this *dbsManager) closeDB() {
	this.mapDSN.Range(func(k, v interface{}) bool {
		db := v.(*sql.DB)
		fmt.Println("close db", k.(int))
		err := db.Close()
		if err != nil {
			fmt.Printf(err.Error())
		}
		return true
	})
}

/******************************************************************************/
func InitDBS(driver, dsn string) error { return dm.initDB(driver, dsn, 0, 0) }

func Query(query string, args ...interface{}) (*sql.Rows, error) {
	return dm.sysdb.Query(query, args...)
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	return dm.sysdb.Exec(query, args...)
}

func service(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	strDT := strings.ToLower(r.FormValue("dt"))
	strSN := r.FormValue("sn")

	mydbs := dm.getService(strSN)
	if mydbs == nil {
		w.WriteHeader(404)
		w.Write([]byte(http.StatusText(404)))
		return
	}

	strSql := mydbs.sql
	db := mydbs.dbConn

	for k, _ := range r.Form {
		strSql = strings.Replace(strSql, "#"+k+"#", r.Form.Get(k), -1)
	}

	w.Header().Set("Connection", "close")
	w.Header().Set("CharacterEncoding", "utf-8")

	var buf bytes.Buffer
	var err error
	switch strings.ToLower(strDT) {
	case "xlsx", "xls":
		err = Query2Xlsx(&buf, db, strSql)
		if err == nil {
			w.Header().Set("Content-Type", "application/vnd.ms-excel")
			w.Header().Set("Content-Disposition", "attachment;filename="+r.FormValue("sn")+".xlsx")
		}
	default:
		err = Query2Json(&buf, db, strSql)
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

/*
func Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, handler)
}
*/

func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func Run(addr string, withTLS bool) {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:           addr,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    120 * time.Second,
		WriteTimeout:   120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	http.Handle("/", http.FileServer(http.Dir("./static/")))
	http.HandleFunc("/dbs", service)

	http.HandleFunc("/dbs/sys/reload", func(w http.ResponseWriter, r *http.Request) {
		dm.loadDSN()
		dm.loadService()
	})

	http.HandleFunc("/dbs/sys/list", func(w http.ResponseWriter, r *http.Request) {
		dm.mapServie.Range(func(k, v interface{}) bool {
			w.Write([]byte(k.(string) + ": " + v.(*dbs).sql + "\r\n"))
			return true
		})
	})

	http.HandleFunc("/dbs/sys/add", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		strSN := r.FormValue("sn")
		strDsnid := r.FormValue("sn")
		strContent := r.FormValue("content")
		dsnid, err := strconv.Atoi(strDsnid)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if dm.addService(strSN, strContent, dsnid) {
			w.Write([]byte("ok"))
		} else {
			w.Write([]byte("error"))
		}
	})

	http.HandleFunc("/dbs/sys/del", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		dm.delService(r.FormValue("sn"))
	})

	go func() {
		dm.loadDSN()
		dm.loadService()

		if withTLS {
			fmt.Println(srv.ListenAndServeTLS("./ca/ca.crt", "./ca/ca.key"))
		} else {
			fmt.Println(srv.ListenAndServe())
		}
	}()

	<-stopChan
	fmt.Println("Shutting down server...")

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	srv.Shutdown(ctx)

	if nil != ctx.Err() {
		fmt.Println(ctx.Err().Error())
	}

	dm.closeDB()
	fmt.Println("Exit OK!")
}
