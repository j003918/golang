// ds project main.go
package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func do(w http.ResponseWriter, r *http.Request) {
	//	str := `[{"name":"姓名","url":"网址"},{"name":"Google","url":"http://www.google.com"},{"name":"Baidu","url":"http://www.baidu.com"},{"name":"SoSo","url":"http://www.SoSo.com"}]`
	//	fmt.Fprint(w, str)
	//	fmt.Println("req in")

	buf := &bytes.Buffer{}
	his.staffInfo(buf)
	w.Write(buf.Bytes())

	//fmt.Fprint(w, str)
}

var his *HIS

func main() {

	his = NewHIS()
	//http://130.1.10.230:8080/dbs?sn=ffxlc&dt=json
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("html"))))
	http.HandleFunc("/test", do)
	fmt.Println(http.ListenAndServe(":8080", nil))
}
