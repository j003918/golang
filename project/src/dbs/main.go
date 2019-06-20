// dbs project main.go
package main

import (
	"bytes"
	"log"
	"net/http"
	"strings"
)

func dbs(w http.ResponseWriter, r *http.Request) {
	//log.Println("dbs in")
	r.ParseForm()
	strSN := r.Form.Get("sn")
	v, ok := snMap.Load(strSN)

	if !ok {
		log.Println("not found service", strSN)
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
}

func main() {
	loadDBConn()
	loadSN()
	http.HandleFunc("/login", login)
	http.HandleFunc("/dbs", dbs)
	http.HandleFunc("/jws/dbs", NewChain(mwValidateToken).ThenFunc(dbs))
	log.Println(http.ListenAndServe(":8080", nil))
}
