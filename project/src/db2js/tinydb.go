// tinydb
package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-oci8"
	//_ "github.com/mattn/go-sqlite3"
)

func OpenDb(timeout time.Duration, driver, dsn string, maxOpen, maxIdle int) (*sql.DB, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	mydb, err := _openDb(driver, dsn, maxOpen, maxIdle)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return mydb, err
}

func _openDb(driver, dsn string, maxOpen, maxIdle int) (*sql.DB, error) {
	mydb, err := sql.Open(driver, dsn)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	mydb.SetMaxOpenConns(maxOpen)
	mydb.SetMaxIdleConns(maxIdle)

	err = mydb.Ping()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return mydb, err
}

//for insert update delete
func ModifyTab(timeout time.Duration, mydb *sql.DB, strsql string, args ...interface{}) (RowsAffected int64, ok bool) {
	ctx, _ := context.WithTimeout(context.Background(), timeout*time.Second)
	rowCount, ok := _modifyTab(mydb, strsql, args...)

	select {
	case <-ctx.Done():
		return -1, false
	default:
	}
	return rowCount, ok
}

func _modifyTab(mydb *sql.DB, strsql string, args ...interface{}) (RowsAffected int64, ok bool) {
	rst, err := mydb.Exec(strsql, args...)
	if err != nil {
		return 0, false
	}

	rowCount, _ := rst.RowsAffected()
	return rowCount, true
}
