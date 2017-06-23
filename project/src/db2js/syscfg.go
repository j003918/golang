// syscfg
package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"datastruct/safemap"
)

var (
	MapDbDriver *safemap.SafeMap
	MapMethod   *safemap.SafeMap

	SysTabCreate_sys_user = `		
		CREATE TABLE IF NOT EXISTS sys_user 
		(
			id 			VARCHAR(64) PRIMARY KEY NOT NULL,     
    		pass		VARCHAR(128) NOT NULL, 
			sign		VARCHAR(32) NOT NULL DEFAULT "", 
    		trustzone 	VARCHAR(512) NOT NULL DEFAULT "*", 
    		status		INTEGER NOT NULL DEFAULT 0, 
    		login_time	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    		create_time	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`

	SysTabCreate_sys_dsn = `		
		CREATE TABLE IF NOT EXISTS sys_dsn 
		(
			id		INTEGER PRIMARY KEY AUTO_INCREMENT,
    		driver	VARCHAR(64) NOT NULL, 
    		dsn		VARCHAR(2048) NOT NULL,
			status 	INTEGER NOT NULL DEFAULT 0,
			info	VARCHAR(128) NOT NULL DEFAULT "",
			update_time	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    		create_time	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`

	SysTabCreate_sys_method = `		
		CREATE TABLE IF NOT EXISTS sys_method 
		(
			method		VARCHAR(128) PRIMARY KEY NOT NULL, 
    		content		VARCHAR(1024) NOT NULL, 
    		dsn_id		INTEGER NOT NULL,
			status 		INTEGER NOT NULL DEFAULT 0, 
    		update_time	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    		create_time	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`
)

type MethdContent struct {
	Method  string
	Content string
	DBConn  *sql.DB
}

type Query2JS struct {
}

func init() {
	//mydb, err := OpenDb(30, "mysql", `jhf:jhf@tcp(130.1.11.60:3306)/test?charset=utf8`, 80, 5)
	mydb, err := OpenDb(30, cmdArgs["driver"], cmdArgs["dsn"], 80, 5)
	if err != nil {
		panic(err)
	}

	MapDbDriver = safemap.NewSafeMap()
	MapMethod = safemap.NewSafeMap()
	MapDbDriver.Set(-1, mydb)

	setupSysTab(mydb)

	go reloadDriver(30)
	go reloadMethod(15)
}

func loopMethod(k, v interface{}) bool {
	return true
}

func loopDb(k, v interface{}) bool {
	v.(*sql.DB).Close()
	return true
}

func CloseAll() {
	MapMethod.LoopDel(loopMethod)
	MapDbDriver.LoopDel(loopDb)
}

func reloadDriver(sec time.Duration) {
	setupDriver()
	timerDriver := time.NewTicker(sec * time.Second)
	for {
		select {
		case <-timerDriver.C:
			setupDriver()
		}
	}
}

func reloadMethod(sec time.Duration) {
	setupMethod()
	timerMethod := time.NewTicker(sec * time.Second)
	for {
		select {
		case <-timerMethod.C:
			setupMethod()
		}
	}
}

func setupDriver() {
	if !MapDbDriver.Check(-1) {
		return
	}

	rows, err := MapDbDriver.Get(-1).(*sql.DB).Query(`select id,driver,dsn from sys_dsn where status=0`)
	if err != nil {
		return
	}
	defer rows.Close()

	id, driver, dsn := -1, "", ""

	for rows.Next() {
		id, driver, dsn = -1, "", ""
		err = rows.Scan(&id, &driver, &dsn)
		if err != nil {
			fmt.Println("setupDriver error:", err.Error())
			return
		}

		if !MapDbDriver.Check(id) {
			newdb, err := OpenDb(30, driver, dsn, 80, 5)
			if err == nil {
				MapDbDriver.Set(id, newdb)
				//fmt.Println("load db driver ", id, driver, dsn)
			}
		}
	}
}

func setupMethod() {
	if MapDbDriver.Get(1) == nil {
		return
	}

	rows, err := MapDbDriver.Get(-1).(*sql.DB).Query(`select method,content,dsn_id from sys_method where status=0`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rows.Close()

	dsn_id, method, content := -1, "", ""
	for rows.Next() {
		dsn_id, method, content = -1, "", ""
		err = rows.Scan(&method, &content, &dsn_id)
		if err != nil {
			fmt.Println("setupMethod error:", err.Error())
			return
		}

		if !MapDbDriver.Check(dsn_id) {
			continue
		}

		if !MapMethod.Check(method) {
			mc := MethdContent{Method: "", Content: "", DBConn: nil}
			mc.Method = method
			mc.Content = content
			mc.DBConn = MapDbDriver.Get(dsn_id).(*sql.DB)
			MapMethod.Set(method, &mc)
			fmt.Println("load method ", method)
		}
	}
}

func setupSysTab(mydb *sql.DB) {
	ModifyTab(15, mydb, `drop table if EXISTS sys_user `)
	ModifyTab(15, mydb, `drop table if EXISTS sys_dsn `)
	ModifyTab(15, mydb, `drop table if EXISTS sys_method `)
	//create sysconfig table
	ModifyTab(15, mydb, SysTabCreate_sys_user)
	ModifyTab(15, mydb, SysTabCreate_sys_dsn)
	ModifyTab(15, mydb, SysTabCreate_sys_method)

	//add init user
	SCUAdd("admin", "czzyy_123")
	SCUAdd("jhf", "jhf")
	SCUAdd("tf", "tf")

	//add test dsn
	SCDAdd("oci8", `system/manager@his`, "his")
	SCDAdd("mysql", `root:root@tcp(172.25.125.101:3306)/oa0618?charset=utf8`, "oa")
	SCDAdd("mysql", `root:root@tcp(130.1.10.230:3306)/zyyoutdoor?charset=utf8`, "zyyoutdoor")

	//add test method
	ModifyTab(15, mydb, `insert into sys_method(method,content,dsn_id) values(?,?,?)`, "inpi", `select DEPT_CODE,DEPT_NAME,CHARGES from COMM.V_JHF_INCOME_OUTP order by DEPT_NAME`, 1)
	ModifyTab(15, mydb, `insert into sys_method(method,content,dsn_id) values(?,?,?)`, "outpi", `select DEPT_CODE,DEPT_NAME,CHARGES from COMM.V_JHF_INCOME_INP order by DEPT_NAME`, 1)
}

func str2md5(str string) string {
	sum := md5.Sum([]byte(str))
	return hex.EncodeToString(sum[:])
}

//SCUAdd table sys_user add new user
func SCUAdd(id, pass string) bool {
	mydb := MapDbDriver.Get(-1).(*sql.DB)
	rowCount, ok := ModifyTab(15, mydb, `insert into sys_user(id,pass) values(?,?)`, id, str2md5(id+pass))
	return rowCount == 1 && ok
}

//SCUCheck table sys_user user login chek
func SCUCheck(id, pass string) bool {
	mydb := MapDbDriver.Get(-1).(*sql.DB)
	rowCount, ok := ModifyTab(15, mydb, `update sys_user set login_time=now() where id=? and pass=? and status=0`, id, str2md5(id+pass))
	return (rowCount == 1) && ok
}

//SCUChangePass table sys_user change password
func SCUChangePass(id, pass, newPass string) bool {
	mydb := MapDbDriver.Get(-1).(*sql.DB)
	rowCount, ok := ModifyTab(15, mydb, `update sys_user set pass=? where id=? and pass=?`, str2md5(id+newPass), id, str2md5(id+pass))
	return rowCount == 1 && ok
}

//SCDAdd table sys_dsn add new dsn
func SCDAdd(driver, dsn, info string) bool {
	mydb := MapDbDriver.Get(-1).(*sql.DB)
	rowCount, ok := ModifyTab(15, mydb, `insert into sys_dsn(driver,dsn,info) values(?,?,?)`, driver, dsn, info)
	return rowCount == 1 && ok
}

//SCDSetDriver table sys_dsn set driver
func SCDSetDriver(driver, id string) bool {
	mydb := MapDbDriver.Get(-1).(*sql.DB)
	_, ok := ModifyTab(15, mydb, `update sys_dsn set driver=? where id=?`, driver, id)
	return ok
}

//SCDSetDSN table sys_dsn set dsn
func SCDSetDSN(dsn, id string) bool {
	mydb := MapDbDriver.Get(-1).(*sql.DB)
	rowCount, ok := ModifyTab(15, mydb, `update sys_dsn set dsn=? where id=?`, id)
	return rowCount == 1 && ok
}
