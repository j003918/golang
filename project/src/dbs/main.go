// dbs project main.go
package main

import (
	"bytes"
	"log"
	"net/http"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

func foo(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("foo---->")
		w.Write([]byte("foo("))
		next(w, r)
		w.Write([]byte(")"))
		log.Println("<----foo")
	}
}

func bar(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("bar---->")
		w.Write([]byte("bar("))
		next(w, r)
		w.Write([]byte(")"))
		log.Println("<----bar")
	}
}

func test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"result":"ok"}`))
}

func main() {
	loadDBConn()
	loadSN()
	http.HandleFunc("/login", login)
	http.HandleFunc("/dbs", NewChain(mwValidateToken).ThenFunc(dbs))
	log.Println(http.ListenAndServe(":8090", nil))
}

func dbs(w http.ResponseWriter, r *http.Request) {
	log.Println("dbs in")
	r.ParseForm()
	strSN := r.Form.Get("sn")
	v, ok := snMap.Load(strSN)

	if !ok {
		log.Println("load SN [", strSN, "] error")
		w.WriteHeader(404)
		w.Write([]byte(http.StatusText(404)))
		return
	}

	si := v.(*serviceInfo)
	strQuery := si.serviceQuery

	for k, v := range r.URL.Query() {
		strQuery = strings.ReplaceAll(strQuery, "#"+k+"#", []string(v)[0])
		log.Println(k, v, strQuery)
	}

	if strings.Contains(strQuery, "#") {
		w.WriteHeader(404)
		w.Write([]byte(http.StatusText(404)))
		return
	}

	rows, _ := si.dbConn.Query(strQuery)

	buf := &bytes.Buffer{}
	rows2Json(rows, buf)
	w.Write(buf.Bytes())
	log.Println("dbs exit")
}
