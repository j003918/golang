// ffxlc project main.go
package main

import (
	"flag"
	"godbs"
	"net/http"
	"strings"
)

var dbs *godbs.GoDBS

func gdi(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	strCols := "_PRIV_"
	strVals := "1"
	strSql := "insert into ffxlc"
	strVal := ""
	cnt := 0
	strRspHtml := "添加失败"
	for k, v := range r.Form {
		strVal = v[0]
		if strVal != "" {
			cnt += 1
			strVal = strings.Replace(strVal, "'", "\\'", -1)
			strCols += "," + k
			strVals += ",'" + strVal + "'"
		}
	}

	if cnt > 0 {
		strSql += "(" + strCols + ") values(" + strVals + ")"
		rowCount, _ := dbs.Exec(30, strSql)

		if rowCount != 1 {
			strRspHtml = "error"
			goto RST
		}

		if rowCount == 1 {
			strRspHtml = "添加成功!"
			goto RST
		}
	}

RST:
	strRspHtml = `<html><body> <div style="text-align:center;"><br><br><br><br><br><br>` + strRspHtml + `<br><a href="/">返回</a></div></body></html>`
	w.Write([]byte(strRspHtml))
}

func main() {
	srvAddr := *(flag.String("port", "8080", ""))
	strDsn := *(flag.String("dsn", "", ""))
	intMaxOpen := *(flag.Int("maxopen", 3, ""))
	intMaxIdle := *(flag.Int("maxidle", 1, ""))
	flag.Parse()

	srvAddr = ":" + srvAddr
	if strDsn == "" {
		strDsn = `jhf:jhf@tcp(130.1.10.230:3306)/czzyy`
	}

	dbs = godbs.NewGoDBS()
	dbs.InitDBS("mysql", strDsn, intMaxOpen, intMaxIdle, srvAddr)
	dbs.HandleFunc("/gdi", gdi)
	dbs.Run(false)
}
