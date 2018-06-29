// ds project main.go
package main

import (
	"bytes"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func do(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := &bytes.Buffer{}
	his.staffInfo(buf)
	w.Write(buf.Bytes())
}

var his *HIS

func main() {
	his = NewHIS()
	router := httprouter.New()
	router.ServeFiles("/web/*filepath", http.Dir("html"))

	router.GET("/test", do)
	log.Println(http.ListenAndServe(":8080", router))
}
