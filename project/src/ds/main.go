// ds project main.go
package main

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"session"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func dojson(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dt := ps.ByName("dt")
	st := ps.ByName("st")
	et := ps.ByName("et")

	buf := &bytes.Buffer{}
	his.staffInfo(buf, dt, st, et)

	if dt == "josn" {
		w.Header().Add("Content-Type", "application/json")
	} else {
		w.Header().Add("Content-Disposition", "attachment")
		//w.Header().Add("Content-Type", "application/vnd.ms-excel")
		w.Header().Add("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	}

	w.Write(buf.Bytes())
}

func redirectUrl(uri string) string {
	return `
	<script language="javascript" type="text/javascript">
	window.location.href="` + uri + `";
	</script> 
	`
}

func xjtf(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	for k, v := range r.Form {
		log.Println(k, v)
	}

	strMsg := "添加成功!"
	strRspHtml := `<html><body> <div style="text-align:center;"><br><br><br><br><br><br>` + strMsg + `<br><a href="home/xjys.html">返回</a></div></body></html>`
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

func login(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	strUrl := "/home/"
	if his.login(r.Form.Get("usr"), r.Form.Get("pwd")) {
		si := smgr.NewSessino()
		si.Set("fid", strings.Split(r.Host, ":")[0]+r.UserAgent())
		//si.Set("useragent", r.UserAgent())
		strUrl = "/home/xjys.html?sid=" + si.SID()
	}

	//	strHtml := `
	//	<script language="javascript" type="text/javascript">
	//	window.location.href="` + strUrl + `";
	//	</script>
	//	`
	w.Write([]byte(redirectUrl(strUrl)))

}

func validAuth(f httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		log.Print(r.Method, r.RequestURI)
		queryForm, err := url.ParseQuery(r.URL.RawQuery)
		if err == nil && len(queryForm["sid"]) > 0 {
			si := smgr.Get(queryForm["sid"][0])
			if si != nil && si.Get("fid").(string) == strings.Split(r.Host, ":")[0]+r.UserAgent() {
				f(w, r, ps)
			} else {
				log.Println("sessino invalid")
				w.WriteHeader(403)
				//w.Write([]byte(redirectUrl("/home/")))
			}
		} else {
			log.Println("not login")
			w.WriteHeader(403)
			//w.Write([]byte(redirectUrl("/home/")))
		}
	}
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

	router.GET("/test/:dt/:st/:et", dojson)
	router.GET("/his/pid/:pid", validAuth(getPatientInfo))
	router.GET("/his/doctor", validAuth(getdoctor))
	router.GET("/his/diag/:key", validAuth(getdiags))
	router.POST("/tf", validAuth(xjtf))
	router.POST("/login", login)
	log.Println(http.ListenAndServe(":8080", router))
}
