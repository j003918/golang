// test project main.go
package main

import (
	"bytes"
	"fmt"
	"lb"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func mssqlTest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := &bytes.Buffer{}
	mssqlInfo(buf)

	w.Header().Set("Content-type", "application/json;charset=utf-8")
	w.Write(buf.Bytes())
}

func mysqlTest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := &bytes.Buffer{}
	mysqlInfo(buf)

	w.Header().Set("Content-type", "application/json;charset=utf-8")
	w.Write(buf.Bytes())
}

func oracleTest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := &bytes.Buffer{}
	oracleInfo(buf)

	w.Header().Set("Content-type", "application/json;charset=utf-8")
	w.Write(buf.Bytes())
}

func proxy(w http.ResponseWriter, r *http.Request) {
	lbs.Proxy(w, r)
}

var lbs *lb.LoadBalance

func main() {
	lbs = lb.NewLB()

	lbs.Register("http", "130.1.10.230:8080", false)
	lbs.Register("http", "130.1.10.230", false)

	//	router := httprouter.New()
	//	router.GET("/mssql", mssqlTest)
	//	router.GET("/mysql", mysqlTest)
	//	router.GET("/oracle", oracleTest)
	//	router.GET("/lb", httpLB)
	//	fmt.Println(http.ListenAndServe(":8080", router))
	http.HandleFunc("/", proxy)
	fmt.Println(http.ListenAndServe(":8080", nil))

}
