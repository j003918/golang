// auth
package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func redirectUrl(uri string) string {
	return `
	<script language="javascript" type="text/javascript">
	window.location.href="` + uri + `";
	</script> 
	`
}

func validAuth(f httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		queryForm, err := url.ParseQuery(r.URL.RawQuery)
		if err == nil && len(queryForm["sid"]) > 0 {
			si := smgr.Get(queryForm["sid"][0])
			if si != nil && si.Get("fid").(string) == strings.Split(r.Host, ":")[0]+r.UserAgent() {
				f(w, r, ps)
			} else {
				log.Println(r.Method, r.RequestURI, "sessino invalid")
				w.Write([]byte(redirectUrl("/home/")))
			}
		} else {
			log.Println(r.Method, r.RequestURI, "not login")
			w.Write([]byte(redirectUrl("/home/")))
		}
	}
}

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	strUrl := "/home/"
	if his.login(r.Form.Get("usr"), r.Form.Get("pwd")) {
		si := smgr.NewSessino()
		si.Set("fid", strings.Split(r.Host, ":")[0]+r.UserAgent())
		strUrl = "/home/xjys.html?sid=" + si.SID()
	}

	w.Write([]byte(redirectUrl(strUrl)))
}

func logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err == nil && len(queryForm["sid"]) > 0 {
		smgr.Del(queryForm["sid"][0])
	}

	w.Write([]byte(redirectUrl("/home/")))
}
