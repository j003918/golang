package main

import (
	"fmt"
	"freenovel"
	//	"math/bits"
	"godbs"
	"net/http"
)

func test_freenovel() {
	nd := freenovel.NewNovelDownloader()
	novelUrl := ""
	for {
		fmt.Print("Please input novel url: ")
		fmt.Scanln(&novelUrl)
		nd.Start(novelUrl)
	}
}

func aa(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("method aa"))
}

func main() {
	strDsn := `jhf:jhf@tcp(130.1.10.230:3306)/czzyy`
	//dbs := godbs.GoDBS.InitDBS()
	//dbs.InitDBS("mysql", strDsn, 3, 1, ":8080")

	//dbs.HandleFunc("/aa", aa)
	//dbs.Run(false)
	//godbs.
	godbs.InitDBS("mysql", strDsn)
	godbs.Run(":8080", false)
}
