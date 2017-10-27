// portmap project main.go

package main

import (
	"datastruct/safemap"
	"fmt"
	"io"
	"net"
	"net/http"
)

var portMap *safemap.SafeMap
var isStop = false

func main() {
	portMap = safemap.NewSafeMap()
	http.HandleFunc("/pm/add", addPortMap)
	http.HandleFunc("/pm/del", delPortMap)
	http.HandleFunc("/pm/list", listPortMap)
	http.ListenAndServe(":9655", nil)
}

func addPortMap(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	localport := r.Form.Get("lp")
	remoteaddr := r.Form.Get("ra")
	if !portMap.Check(localport) {
		portMap.Set(localport, remoteaddr)
		go setupPM(localport, remoteaddr)
		w.Write([]byte("ok"))
	}
}

func delPortMap(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	localport := r.Form.Get("lp")
	portMap.Del(localport)
}

func listPortMap(w http.ResponseWriter, r *http.Request) {
	//portMap.Println()
}

func handle(sconn net.Conn, ra string) {
	dconn, err := net.Dial("tcp", ra)
	if err != nil {
		return
	}

	fmt.Println(sconn.RemoteAddr(), dconn.RemoteAddr())
	go io.Copy(sconn, dconn)
	go io.Copy(dconn, sconn)
}

func setupPM(lp, ra string) {
	l, err := net.Listen("tcp", ":"+lp)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for !isStop {
		sconn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handle(sconn, ra)
	}
}
