// ds project main.go
package main

import (
	"bytes"
	"log"
	"net/http"
	"session"

	"github.com/julienschmidt/httprouter"
)

func xjtf(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	for k, v := range r.Form {
		log.Println(k, v)
	}

	strMsg := "添加成功!"
	strRspHtml := `<html><body> <div style="text-align:center;"><br><br><br><br><br><br>` + strMsg + `<br><a href="home/xjys.html?"` + r.URL.RawQuery + `>返回</a></div></body></html>`
	w.Write([]byte(strRspHtml))
}

func getPatientInfo(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	w.Header().Add("Content-Type", "application/json")
	pid := ps.ByName("pid")
	buf := &bytes.Buffer{}
	his.patientInfo(buf, pid)
	w.Write(buf.Bytes())
}

func getdoctor(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "application/json")
	buf := &bytes.Buffer{}
	his.docotors(buf)
	w.Write(buf.Bytes())
}

func getdiags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Add("Content-Type", "application/json")
	key := ps.ByName("key")
	buf := &bytes.Buffer{}
	his.diagnosis(buf, key)
	w.Write(buf.Bytes())
}

var (
	his  *HIS
	smgr *session.SessionMgr
)

func init() {
	his = NewHIS()
	smgr = session.NewSessionMgr()
}

func main() {
	router := httprouter.New()
	router.ServeFiles("/home/*filepath", http.Dir("web"))

	router.GET("/his/pid/:pid", validAuth(getPatientInfo))
	router.GET("/his/doctor", validAuth(getdoctor))
	router.GET("/his/diag/:key", validAuth(getdiags))
	router.GET("/logout", validAuth(logout))
	router.POST("/login", login)
	router.POST("/tf", validAuth(xjtf))
	log.Println(http.ListenAndServe(":8080", router))
}
