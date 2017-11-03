// godbs project main.go
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"tinydb"
)

var (
	mapService sync.Map
	srvAddr    string
	dbConn     *sql.DB
)

func init() {
	srvAddr = *(flag.String("port", "8080", ""))
	strDsn := *(flag.String("dsn", "", ""))
	intMaxOpen := *(flag.Int("maxopen", 3, ""))
	intMaxIdle := *(flag.Int("maxidle", 1, ""))
	flag.Parse()

	srvAddr = ":" + srvAddr
	if strDsn == "" {
		strDsn = `jhf:jhf@tcp(130.1.10.230:3306)/czzyy`
	}

	_mydb, err := tinydb.OpenDb(30, "mysql", strDsn, intMaxOpen, intMaxIdle)
	if err != nil {
		panic(err)
	}
	dbConn = _mydb
	initConf()

	init_gdi()

	loadService()
}

func dbs(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	strDT := r.FormValue("dt")
	strSqlObj, ok := mapService.Load(strings.ToLower(r.FormValue("sn")))
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

	switch strings.ToLower(strDT) {
	case "xls":
		fallthrough
	case "xlsx":
		w.Header().Set("Content-Type", "application/vnd.ms-excel") //application/vnd.ms-excel or application/x-xls
		w.Header().Set("Content-Disposition", "attachment;filename="+r.FormValue("sn")+".xlsx")
	default:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	tinydb.Sql2Writer(30, dbConn, strSql, w, strDT)
}

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	http.Handle("/", http.FileServer(http.Dir("./static/")))
	http.HandleFunc("/dbs", dbs)

	http.HandleFunc("/dbs/sys/reload", func(w http.ResponseWriter, r *http.Request) {
		loadService()
	})

	http.HandleFunc("/dbs/sys/list", func(w http.ResponseWriter, r *http.Request) {
		mapService.Range(func(k, v interface{}) bool {
			w.Write([]byte(k.(string) + ": " + v.(string) + "\r\n"))
			return true
		})
	})

	http.HandleFunc("/dbs/sys/add", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		strSN := strings.ToLower(r.FormValue("sn"))
		strContent := r.FormValue("content")
		_, ok := mapService.LoadOrStore(strSN, strContent)

		if !ok {
			//w.Write([]byte(http.StatusText(200)))
			w.Write([]byte(http.StatusText(200)))
		} else {
			w.WriteHeader(500)
			w.Write([]byte(http.StatusText(500)))
		}
	})

	http.HandleFunc("/dbs/sys/del", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		strSN := strings.ToLower(r.FormValue("sn"))
		mapService.Delete(strSN)
	})

	// add ffxlc insert function
	http.HandleFunc("/gdi", gdi)
	//

	srv := &http.Server{
		Addr:           srvAddr,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		fmt.Println(srv.ListenAndServe())
	}()

	<-stopChan
	fmt.Println("Shutting down server...")

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	srv.Shutdown(ctx)

	if nil != ctx.Err() {
		fmt.Println(ctx.Err().Error())
	}

	fmt.Println("Exit OK!")
}
