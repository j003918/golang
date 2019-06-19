// service
package main

import (
	"database/sql"
	"log"
	"sync"

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
        pk_id        INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
        driver_name  VARCHAR(64) NOT NULL UNIQUE,
        notes        VARCHAR(256),
        del_flag     INT DEFAULT 0 NOT NULL,
        created_time DATETIME NOT NULL DEFAULT (datetime('now', 'localtime'))
    );
    
    CREATE TABLE IF NOT EXISTS dsn(
        pk_id        INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
        driver_id    INT NOT NULL,
	    data_source  VARCHAR(1024) NOT NULL,
        del_flag     INT DEFAULT 0 NOT NULL,
        created_time DATETIME NOT NULL DEFAULT (datetime('now', 'localtime'))
    );
    
    CREATE TABLE IF NOT EXISTS service(
        pk_id        INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
        dsn_id       INT NOT NULL,
        service_name VARCHAR(128) NOT NULL UNIQUE,
	    query_sql    VARCHAR(3072) NOT NULL,
        del_flag     INT DEFAULT 0 NOT NULL,
        created_time DATETIME NOT NULL DEFAULT (datetime('now', 'localtime'))
    );
    
    CREATE VIEW IF NOT EXISTS v_dsn AS
    SELECT a.pk_id as dsn_id, 
        b.driver_name,
        a.data_source 
    FROM dsn a
        JOIN driver b on a.driver_id=b.pk_id and b.del_flag = 0
    WHERE a.del_flag = 0;
    
    CREATE VIEW IF NOT EXISTS v_service AS
    SELECT a.pk_id,
        a.dsn_id,
        a.service_name,
        a.query_sql
        --b.data_source,
        --c.driver_name
    FROM service a
        JOIN dsn b ON a.dsn_id = b.pk_id and b.del_flag = 0
        JOIN driver c ON b.driver_id = c.pk_id and c.del_flag = 0
    WHERE a.del_flag = 0;
	`

	initTable := `
	insert into account(accout_no,password) values('root','toor');
    
    insert into driver(driver_name,notes) values('mysql','user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true');
    insert into driver(driver_name,notes) values('mssql','SQL Server 2005 or newer;  sqlserver://sa:mypass@localhost:1234?database=master&connection+timeout=30');
    insert into driver(driver_name,notes) values('oci8','system/manager@//130.1.10.90:1521/orcl');
    
    insert into dsn(driver_id,data_source) values(1,'jhf:jhf@tcp(192.168.0.244:3306)/itop');

    insert into service(dsn_id,service_name,query_sql) values(1,'mysql_info','select version() as ver, now() as cur_time;');
    insert into service(dsn_id,service_name,query_sql) values(1,'itop','select * from contact where id=#id#;');
	`

	_, err = sysdb.Exec(createSysTable)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	sysdb.Exec(initTable)
}

func openDBConn(driver, dsn string) *sql.DB {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Println("openDBConn", driver, dsn, err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		log.Println("openDBConn", driver, dsn, err)
		return nil
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(50)
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

		db := openDBConn(driver_name, data_source)
		if db != nil {
			dbMap.Store(dsn_id, db)
			//log.Println("load db", driver_name, data_source)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
}

type serviceInfo struct {
	dsn_id       int
	dbConn       *sql.DB
	serviceName  string
	serviceQuery string
}

func loadSN() {
	strSql := `select dsn_id,service_name,query_sql from v_service`
	rows, err := sysdb.Query(strSql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	dsn_id := 0
	snName, snQuery := "", ""
	for rows.Next() {
		err = rows.Scan(&dsn_id, &snName, &snQuery)
		if err != nil {
			log.Println(err)
			continue
		}

		db, ok := dbMap.Load(dsn_id)
		if !ok {
			log.Println("get db conn error ", dsn_id, err)
			continue
		}
		si := &serviceInfo{
			dsn_id:       dsn_id,
			dbConn:       db.(*sql.DB),
			serviceName:  snName,
			serviceQuery: snQuery,
		}

		snMap.Store(snName, si)
		log.Println("load service:", snName)
	}
}
