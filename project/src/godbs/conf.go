// config.go
package main

import (
	"fmt"
	"strings"
	"tinydb"
)

func initConf() {
	strSql := `		
		CREATE TABLE IF NOT EXISTS godbs 
		(
			sn			VARCHAR(64)		PRIMARY KEY NOT NULL,     
    		content		VARCHAR(4096) 	NOT NULL, 
			name		VARCHAR(128) 	DEFAULT NULL, 
    		create_time	TIMESTAMP		NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`

	tinydb.ModifyTab(5, dbConn, strSql)
}

func loadService() {
	strSql := "select sn,content from godbs"
	rows, err := dbConn.Query(strSql)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	strSN, strContent := "", ""
	for rows.Next() {
		err = rows.Scan(&strSN, &strContent)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		mapService.Store(strings.ToLower(strSN), strContent)
	}
}
