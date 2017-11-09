// sp
package godbs

import (
	"database/sql"
	"fmt"
	"log"
)

var (
	mysql_SP_EXEC_SP_DROP   = `DROP PROCEDURE IF EXISTS SP_JHF_TEST_GOMYSQL;`
	mysql_SP_EXEC_SP_CREATE = `CREATE PROCEDURE SP_JHF_TEST_GOMYSQL (n1 INT, n2 INT, OUT out1 INT,OUT out2 INT)
	BEGIN 
		SET out1 = n1 + n2;
		SET out2 = out1*out1;
	END;`
	mysql_SP_EXEC = `call SP_JHF_TEST_GOMYSQL(?,?,@out1,@out2);`
)

var (
	oracle_SP_CREATE = `create or replace procedure SP_JHF_TEST_GOOCI8
	(
	p1 in number
	) is
	begin
		--insert into jhf_test_tmp (id) values (p1);
		DBMS_OUTPUT.PUT_LINE(p1);
	end;`
	oracle_SP_DROP = `DROP procedure SP_JHF_TEST_GOOCI8`
	oracle_SP_EXEC = `call SP_JHF_TEST_GOOCI8(:in1)`
)

func test_mysql_sp(db *sql.DB) {
	_, err := db.Exec(mysql_SP_EXEC_SP_DROP)
	if err != nil {
		panic(err.Error())
	}
	defer db.Exec(mysql_SP_EXEC_SP_DROP)

	rows, err := db.Query(mysql_SP_EXEC_SP_CREATE)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	stmt, err := db.Prepare(mysql_SP_EXEC)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()

	n1, n2 := 3, 4
	_, err = stmt.Exec(n1, n2)
	if err != nil {
		panic(err.Error())
	}

	var sql string = "SELECT @out1 as out1,@out2 as out2"
	selectInstance, err := db.Prepare(sql)
	if err != nil {
		panic(err.Error())
	}
	defer selectInstance.Close()

	var out1, out2 int
	err = selectInstance.QueryRow().Scan(&out1, &out2)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(out1, out2)
}

func test_oci8_sp(db *sql.DB) {
	db.Exec(oracle_SP_CREATE)
	defer db.Exec(oracle_SP_DROP)

	stmt, err := db.Prepare(oracle_SP_EXEC)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer stmt.Close()

	p1 := 10
	_, err = stmt.Exec(p1)

	if err != nil {
		log.Fatal(err.Error())
	}
}
