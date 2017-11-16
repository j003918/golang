// gons3918 project gons3918.go
package gons3918

import (
	"net/http"
	"strings"
)

func init() {
	http.HandleFunc("/getip", getip)
}

//https://gons3918.appspot.com/
func getip(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	strIP := r.RemoteAddr
	index := strings.LastIndexAny(strIP, ":")
	if index > 0 {
		strIP = string([]rune(strIP)[0:index])
	}

	w.Write([]byte(strIP))
}
