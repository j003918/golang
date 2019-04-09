// dbtest project main.go
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	//"github.com/tealeg/xlsx"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-oci8"
)

/*

# native compiler windows amd64

GOROOT=D:\PortableSoftware\go
#GOBIN=
GOARCH=amd64
GOOS=windows
CGO_ENABLED=1

MINGW64=D:/PortableSoftware/mingw-w64/mingw64
instantclient=D:/instantclient_12_2
PKG_CONFIG_PATH=%instantclient%/pkg-config
TNS_ADMIN=%instantclient%/network/admin
#PATH=c:\mingw64\bin;%GOROOT%\bin;%PATH%
PATH=%PATH%;%MINGW64%/bin;%GOROOT%/bin;%instantclient%;%instantclient%/pkg-config

LITEIDE_GDB=gdb64
LITEIDE_MAKE=mingw32-make
LITEIDE_TERM=%COMSPEC%
LITEIDE_TERMARGS=
LITEIDE_EXEC=%COMSPEC%
LITEIDE_EXECOPT=/C


*/

func Rows2Json(rows *sql.Rows, out_buf *bytes.Buffer) bool {
	colKeys, err := rows.Columns()
	if err != nil {
		log.Println(err)
		return false
	}

	colVals := make([]sql.NullString, len(colKeys))
	colValsPtr := make([]interface{}, len(colKeys))
	var valBuf []byte

	for i, _ := range colKeys {
		colKeys[i] = strings.ToLower(colKeys[i])
		colValsPtr[i] = &colVals[i]
	}

	out_buf.WriteByte('[')
	for rows.Next() {
		err = rows.Scan(colValsPtr...)
		if err != nil {
			log.Println(err)
			return false
		}

		out_buf.WriteByte('{')
		for i, val := range colVals {
			valBuf, err = json.Marshal(&val.String)
			if err != nil {
				log.Println(err)
				return false
			}
			out_buf.WriteString(`"` + colKeys[i] + `":` + string(valBuf) + `,`)
		}
		out_buf.Bytes()[out_buf.Len()-1] = '}'
		out_buf.WriteByte(',')
	}

	if out_buf.Len() > 1 {
		out_buf.Bytes()[out_buf.Len()-1] = ']'
	} else {
		out_buf.WriteByte(']')
	}

	return true
}

/*
func Rows2Xlsx(rows *sql.Rows, out_buf *bytes.Buffer) bool {
	colKeys, err := rows.Columns()
	if err != nil {
		log.Println(err)
		return false
	}

	colVals := make([]sql.NullString, len(colKeys))
	colValsPtr := make([]interface{}, len(colKeys))

	tmp := xlsx.NewFile()
	sheet, err := tmp.AddSheet("sheet1")
	if err != nil {
		log.Println(err)
		return false
	}

	for i, _ := range colKeys {
		colKeys[i] = strings.ToLower(colKeys[i])
		colValsPtr[i] = &colVals[i]
	}

	sheet.AddRow().WriteSlice(&colKeys, -1)
	cells := make([]string, len(colKeys))
	for rows.Next() {
		err = rows.Scan(colValsPtr...)
		if err != nil {
			log.Println(err)
			return false
		}

		for i, val := range colVals {
			cells[i] = val.String
		}
		sheet.AddRow().WriteSlice(&cells, -1)
	}

	//tmp.Save("test.xlsx")
	tmp.Write(out_buf)

	return true
}
*/

func mssqlInfo(buf *bytes.Buffer) bool {
	strConn := `sqlserver://sa:www126.com@130.1.10.217:1433?database=rights&encrypt=disable`
	strSql := `select @@version,GETDATE() as sysdate`

	db, err := sql.Open("sqlserver", strConn)
	if err != nil {
		return false
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return false
	}

	rows, err := db.Query(strSql)
	if err != nil {
		return false
	}
	defer rows.Close()

	return Rows2Json(rows, buf)
}

func mysqlInfo(buf *bytes.Buffer) bool {
	strConn := `root:root@tcp(130.1.10.230:3306)/zyyoutdoor`
	strSql := `select version(), now()`

	db, err := sql.Open("mysql", strConn)
	if err != nil {
		return false
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return false
	}

	rows, err := db.Query(strSql)
	if err != nil {
		return false
	}
	defer rows.Close()

	return Rows2Json(rows, buf)
}

func oracleInfo(buf *bytes.Buffer) bool {
	strConn := `system/manager@//130.1.10.90:1521/orcl`
	strSql := `select VERSION,sysdate from v$instance`
	//strSql := `select USER_NAME,HRP_USER_NAME,NAME,JOB,CREATE_DATE from staff_dict order by USER_NAME`

	db, err := sql.Open("oci8", strConn)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {

		fmt.Println(err.Error(), "2222")
		return false
	}

	rows, err := db.Query(strSql)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer rows.Close()

	return Rows2Json(rows, buf)
}

func dbinfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var rst bool
	buf := &bytes.Buffer{}
	switch strings.ToLower(r.URL.Path) {
	case "/mysql":
		rst = mysqlInfo(buf)
	case "/mssql":
		rst = mssqlInfo(buf)
	case "/oracle":
		rst = oracleInfo(buf)
	default:
		http.NotFound(w, r)
		return
	}

	if rst {
		w.Header().Set("Content-type", "application/json;charset=utf-8")
		w.Write(buf.Bytes())
	} else {
		w.Write([]byte("error"))
	}
}

func main() {
	router := httprouter.New()
	router.GET("/mssql", dbinfo)
	router.GET("/mysql", dbinfo)
	router.GET("/oracle", dbinfo)

	fmt.Println(http.ListenAndServe(":8090", router))
}
