// service
package main

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var dbMap, snMap sync.Map
var sysdb *sql.DB

func init() {
	var err error
	sysdb, err = sql.Open("sqlite3", "file:bdtoor")
	if err != nil {
		log.Println(err)
		panic(err)
	}

	createSysTable := `
	CREATE TABLE IF NOT EXISTS account (
        accout_no    VARCHAR (128) PRIMARY KEY UNIQUE NOT NULL,
        password     VARCHAR (128) NOT NULL,
        user_id      VARCHAR (128) NOT NULL DEFAULT USERS,
        del_flag     INT DEFAULT 0 NOT NULL,
        created_time DATETIME NOT NULL DEFAULT (datetime('now', 'localtime'))
    );
    
	CREATE TABLE IF NOT EXISTS driver(
        id           INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
        driver_name  VARCHAR(64) NOT NULL UNIQUE,
        notes        VARCHAR(256),
        del_flag     INT DEFAULT 0 NOT NULL,
        created_time DATETIME NOT NULL DEFAULT (datetime('now', 'localtime'))
    );
    
    CREATE TABLE IF NOT EXISTS dsn(
        id           INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
        driver_id    INT NOT NULL,
	    data_source  VARCHAR(1024) NOT NULL,
        del_flag     INT DEFAULT 0 NOT NULL,
        created_time DATETIME NOT NULL DEFAULT (datetime('now', 'localtime'))
    );
    
    CREATE TABLE IF NOT EXISTS service(
        id           INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
        dsn_id       INT NOT NULL,
        service_name VARCHAR(128) NOT NULL UNIQUE,
	    query_sql    VARCHAR(3072) NOT NULL,
        del_flag     INT DEFAULT 0 NOT NULL,
        modify_time  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        created_time DATETIME NOT NULL DEFAULT (datetime('now', 'localtime'))
    );
    
    CREATE VIEW IF NOT EXISTS v_dsn AS
    SELECT a.id as dsn_id, 
        b.driver_name,
        a.data_source 
    FROM dsn a
        JOIN driver b on a.driver_id=b.id and b.del_flag = 0
    WHERE a.del_flag = 0;
    
    CREATE VIEW IF NOT EXISTS v_service AS
    SELECT a.service_name,
        a.dsn_id,
        --a.service_name,
        a.query_sql,
        a.del_flag
        --b.data_source,
        --c.driver_name
    FROM service a
        JOIN dsn b ON a.dsn_id = b.id and b.del_flag = 0
        JOIN driver c ON b.driver_id = c.id and c.del_flag = 0
    --WHERE a.del_flag = 0;
	`

	initTable := `
	insert into account(accout_no,password) values('root','toor');
    
    insert into driver(driver_name,notes) values('mysql','user:password@tcp(host:port)/dbname');
    insert into driver(driver_name,notes) values('mssql','SQL Server 2005 or newer;  sqlserver://sa:mypass@localhost:1234?database=master&connection+timeout=30');
    insert into driver(driver_name,notes) values('oci8','system/manager@//130.1.10.90:1521/orcl');
    
    --insert into dsn(driver_id,data_source) values(1,'jhf:jhf@tcp(192.168.0.244:3306)/itop');

    --insert into service(dsn_id,service_name,query_sql) values(1,'mysql_info','select version() as ver, now() as cur_time;');
    --insert into service(dsn_id,service_name,query_sql) values(1,'itop','select * from contact where id=#id#;');
	`

	_, err = sysdb.Exec(createSysTable)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	sysdb.Exec(initTable)
	go refreshService()
}

func refreshService() {
	timer1 := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-timer1.C:
			loadDBConn()
			loadSN()
		}
	}
}

func openDBConn(driver, dsn string) *sql.DB {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Println("sql.Open error", err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		log.Println("db.Ping error", err)
		return nil
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(100)
	return db
}

func loadDBConn() {
	strSql := `select dsn_id,driver_name,data_source from v_dsn`
	rows, err := sysdb.Query(strSql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	dsn_id := 0
	driver_name, data_source := "", ""
	for rows.Next() {
		err = rows.Scan(&dsn_id, &driver_name, &data_source)
		if err != nil {
			log.Println(err)
			continue
		}

		if _, ok := dbMap.Load(dsn_id); ok {
			continue
		}

		db := openDBConn(driver_name, data_source)
		if db == nil {
			continue
		}

		//if db != nil {
		dbMap.Store(dsn_id, db)
		//log.Println("load db", driver_name, data_source)
		//}
	}
	// err = rows.Err()
	// if err != nil {
	// 	log.Println(err)
	// }
}

type serviceInfo struct {
	dsn_id       int
	dbConn       *sql.DB
	serviceName  string
	serviceQuery string
}

func loadSN() {
	strSql := `select dsn_id,service_name,query_sql,del_flag from v_service`
	rows, err := sysdb.Query(strSql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	dsn_id, del_flag := 0, 0
	snName, snQuery := "", ""
	for rows.Next() {
		err = rows.Scan(&dsn_id, &snName, &snQuery, &del_flag)
		if err != nil {
			log.Println(err)
			continue
		}

		if del_flag == 1 {
			snMap.Delete(snName)
			continue
		}

		db, ok := dbMap.Load(dsn_id)
		if !ok {
			log.Println("get db conn error ", err)
			continue
		}

		if _, ok := snMap.Load(snName); ok {
			//continue or delete?
			snMap.Delete(snName)
		}

		si := &serviceInfo{
			dsn_id:       dsn_id,
			dbConn:       db.(*sql.DB),
			serviceName:  snName,
			serviceQuery: snQuery,
		}

		snMap.Store(snName, si)
		//log.Println("load service:", snName)
	}
}
