// test project main.go
package main

import (
	"bytes"
	"fmt"
	"loadbalance"
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

var lb *loadbalance.LB

func main() {
	lb = loadbalance.NewLB()

	lb.Register("http", "130.1.10.230:8080")
	lb.Register("http", "130.1.10.230")

	//	router := httprouter.New()
	//	router.GET("/mssql", mssqlTest)
	//	router.GET("/mysql", mysqlTest)
	//	router.GET("/oracle", oracleTest)
	//	router.GET("/lb", httpLB)
	//	fmt.Println(http.ListenAndServe(":8080", router))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { lb.Proxy(w, r) })
	fmt.Println(http.ListenAndServe(":8080", nil))
}
