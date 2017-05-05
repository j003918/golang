/* liteide *.env add ENV support oci8
MINGW64=D:/mingw-w64/mingw64
instantclient=D:/instantclient_12_2
PKG_CONFIG_PATH=%instantclient%/pkg-config
TNS_ADMIN=%instantclient%/network/admin
PATH=%PATH%;%MINGW64%/bin;%GOROOT%/bin;%instantclient%;%instantclient%/pkg-config
*/

// db2js project main.go
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"sql2json"
)

var (
	cmdArgs    map[string]string
	curReqNum  = 0
	curReqLock sync.Mutex
)

func init() {
	os.Setenv("NLS_LANG", "AMERICAN_AMERICA.AL32UTF8")
	cmdArgs = make(map[string]string)
	tls := flag.Int("tls", 0, "0:disable 1:enable")
	port := flag.Int("port", 80, "http:80 https:443")
	strdbdriver := flag.String("driver", "mysql", "")
	strdsn := flag.String("dsn", `jhf:jhf@tcp(130.1.11.60:3306)/test?charset=utf8`, "")
	flag.Parse()

	cmdArgs["tls"] = strconv.Itoa(*tls)
	cmdArgs["port"] = strconv.Itoa(*port)
	cmdArgs["driver"] = *strdbdriver
	cmdArgs["dsn"] = *strdsn
}

func main() {
	listen_addr := ":"
	if "80" == cmdArgs["port"] && "1" == cmdArgs["tls"] {
		listen_addr += "443"
	} else {
		listen_addr += cmdArgs["port"]
	}

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	http.Handle("/", http.FileServer(http.Dir("./html/")))
	http.HandleFunc("/get", worker)
	http.HandleFunc("/info/online", activeGuest)
	http.HandleFunc("/auth/login", guestlogin)

	srv := &http.Server{
		Addr:           listen_addr,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if "1" == cmdArgs["tls"] {
			fmt.Println(srv.ListenAndServeTLS("./ca/ca.crt", "./ca/ca.key"))
		} else {
			fmt.Println(srv.ListenAndServe())
		}
	}()

	<-stopChan
	fmt.Println("Shutting down server...")

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	srv.Shutdown(ctx)
	CloseAll()

	if nil != ctx.Err() {
		fmt.Println(ctx.Err().Error())
	}
}

func guestlogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user := r.FormValue("user")
	//user := r.Header.Get("x-auth-user")
	pass := r.FormValue("pass")

	strIP := r.RemoteAddr
	index := strings.LastIndexAny(strIP, ":")
	if index > 0 {
		strIP = string([]rune(strIP)[0:index])
	}

	if !AddAuth(strIP, user, pass) {
		w.WriteHeader(401)
		w.Write([]byte(http.StatusText(401)))
	} else {
		w.Write([]byte(http.StatusText(200)))
	}
}

func activeGuest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(curReqNum)))
}

func incGuest() {
	curReqLock.Lock()
	curReqNum++
	curReqLock.Unlock()
}

func decGuest() {
	curReqLock.Lock()
	curReqNum--
	curReqLock.Unlock()
}

//http://127.0.0.1/get?m=fee&param=w
func worker(w http.ResponseWriter, r *http.Request) {
	incGuest()
	defer decGuest()
	r.ParseForm()

	user := r.Form.Get("user")
	//user := r.Header.Get("x-auth-user")

	strIP := r.RemoteAddr
	index := strings.LastIndexAny(strIP, ":")
	if index > 0 {
		strIP = string([]rune(strIP)[0:index])
	}

	if !CheckAuth(strIP, user) {
		w.WriteHeader(401)
		w.Write([]byte(http.StatusText(401)))
		return
	}

	timeout := 30 * time.Second
	strCmd := r.Form.Get("m")
	strSql := ""
	if MapMethod.Check(strCmd) {
		strSql = MapMethod.Get(strCmd).(*MethdContent).Content
	} else {
		w.WriteHeader(404)
		w.Write([]byte(http.StatusText(404)))
		return
	}

	for k, _ := range r.Form {
		strSql = strings.Replace(strSql, "#"+k+"#", r.Form.Get(k), -1)
	}

	var err error
	var bufdata bytes.Buffer

	if strings.ContainsAny(strSql, "#") {
		w.WriteHeader(400)
		w.Write([]byte(http.StatusText(400)))
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	err = sql2json.GetJson(ctx, MapMethod.Get(strCmd).(*MethdContent).DBConn, strSql, &bufdata)
	if nil != err {
		w.WriteHeader(500)
		w.Write([]byte(http.StatusText(500)))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Connection", "close")
	w.Header().Set("CharacterEncoding", "utf-8")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0")

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && bufdata.Len() >= 1024*10 {
		var gzbuf bytes.Buffer
		gz := gzip.NewWriter(&gzbuf)
		_, err = gz.Write(bufdata.Bytes())
		gz.Close()
		if err == nil {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Length", strconv.Itoa(gzbuf.Len()))
			w.Write(gzbuf.Bytes())
		} else {
			fmt.Println(err.Error())
			w.Write(bufdata.Bytes())
		}
	} else {
		w.Write(bufdata.Bytes())
	}
}
