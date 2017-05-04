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

	http.HandleFunc("/info/service", listMethod)
	http.HandleFunc("/info/online", activeGuest)

	http.HandleFunc("/auth/login", guestlogin)

	http.HandleFunc("/cfg", setConfig)
	http.HandleFunc("/m/del", delMethod)

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
	pass := r.FormValue("pass")

	if !AddAuth(r.RemoteAddr, user, pass) {
		w.Write([]byte("User Name or Password does not exist"))
	} else {
		w.Write([]byte("ok"))
	}
}

func setConfig(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cfgKey := r.Form.Get("k")
	cfgVal := r.Form.Get("v")
	if _, ok := cmdArgs[cfgKey]; ok {
		cmdArgs[cfgKey] = cfgVal
	}
}

func delMethod(w http.ResponseWriter, r *http.Request) {
	//r.ParseForm()
	//delete(methodSql, r.Form.Get("m"))
}

func isLoop(k, v interface{}) bool {
	fmt.Println(k.(string))
	return false
}

func listMethod(w http.ResponseWriter, r *http.Request) {
	//signMap.Loop(isLoop)

	//	strTmp := ""
	//	for k, v := range methodSql {
	//		strTmp += k + "{" + string('\n') + v + string('\n') + "}" + strings.Repeat(string('\n'), 2)
	//	}
	//	w.Write([]byte(strTmp))
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	user := r.Form.Get("user")
	//user := r.Header.Get("x-auth-user")

	if !CheckAuth(r.RemoteAddr, user) {
		strJson := `{"result":"` + Code401 + `",` + `"msg":"` + Code401Msg + `", "data": null}`
		w.Write([]byte(strJson))
		return
	}

	timeout := 30 * time.Second
	strCmd := r.Form.Get("m")
	strSql := ""
	if MapMethod.Check(strCmd) {
		strSql = MapMethod.Get(strCmd).(*MethdContent).Content
	} else {
		strJson := `{"result":"` + Code400 + `",` + `"msg":"` + Code400Msg + `", "data": null}`
		w.Write([]byte(strJson))
		return
	}

	for k, _ := range r.Form {
		strSql = strings.Replace(strSql, "#"+k+"#", r.Form.Get(k), -1)
	}

	strRst := Code400
	strMsg := Code400Msg
	strVal := ""
	var err error
	var bufdata bytes.Buffer

	if strCmd == "" || strSql == "" || strings.ContainsAny(strSql, "#") {
		strRst = Code400
		strMsg = Code400Msg
		strVal = "null"
	} else {
		ctx, _ := context.WithTimeout(context.Background(), timeout)
		err = sql2json.GetJson(ctx, MapMethod.Get(strCmd).(*MethdContent).Mthdb, strSql, &bufdata)
		if nil != err {
			strRst = Code500
			strMsg = err.Error()
			strVal = "null"
		} else {
			strRst = Code200
			strMsg = Code200Msg
		}
	}

	var json_buf bytes.Buffer
	json_buf.WriteString(`{"result":` + strRst + ",")
	json_buf.WriteString(`"msg":"` + strMsg + `",`)
	json_buf.WriteString(`"data":`)
	if "" == strVal {
		json_buf.Write(bufdata.Bytes())
	} else {
		json_buf.WriteString(strVal)
	}
	json_buf.WriteString("}")

	//w.Header().Set("Connection", "close")
	//w.Header().Set("CharacterEncoding", "utf-8")
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//w.Header().Set("Pragma", "no-cache")
	//w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0")
	//w.Header().Set("Expires", "1L")

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && json_buf.Len() >= 1024 {
		var gzbuf bytes.Buffer
		gz := gzip.NewWriter(&gzbuf)
		_, err = gz.Write(json_buf.Bytes())
		gz.Close()
		if err == nil {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Length", strconv.Itoa(gzbuf.Len()))
			w.Write(gzbuf.Bytes())
		} else {
			fmt.Println(err.Error())
			w.Write(json_buf.Bytes())
			return
		}
	} else {
		w.Write(json_buf.Bytes())
	}
}
