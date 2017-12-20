package main

import (
	"fmt"
	"net/http"
	"route"
	"time"
)

func user(rw http.ResponseWriter, r *http.Request) {
	dd1 := r.URL.Query().Get("dts")
	dd2 := r.URL.Query().Get("dte")
	rw.Write([]byte(dd1 + ":" + dd2))
}

func bb(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("b"))
}

func logOut(rw http.ResponseWriter, r *http.Request) {
	fmt.Println(time.Now().Format(time.StampNano), r.Method, r.RequestURI)
}

func main() {
	mu := route.New()

	mu.Get("/dbs/{dts}/{dte:([0-9]+)}", user)
	mu.Get("/st/a/b", bb)
	mu.Static("/", "./static/")
	mu.Filter(logOut)

	http.ListenAndServe(":8080", mu)
}
