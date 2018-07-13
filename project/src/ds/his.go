// his
package main

import (
	"bytes"
	"database/sql"
	"dbs"
	"log"
	"strings"

	_ "github.com/mattn/go-oci8"
)

//var HIS_DB *sql.DB

type HIS struct {
	conn *sql.DB
}

func NewHIS() *HIS {
	strDSN := `system/manager@//130.1.10.90:1521/orcl`
	db, err := sql.Open("oci8", strDSN)
	db.Ping()
	if err != nil {
		return nil
	}
	return &HIS{conn: db}
}

func (h *HIS) staffInfo(buf *bytes.Buffer, dt, st, et string) bool {
	strSql := `select * from OUTPBILL.OUTP_BILL_ITEMS where VISIT_DATE >= date':1' and VISIT_DATE < date':2'`

	rows, err := h.conn.Query(strSql, st, et)
	if err != nil {
		log.Println(err)
		return false
	}
	defer rows.Close()
	if dt == "json" {
		dbs.Rows2Json(rows, buf)
	} else {
		dbs.Rows2Xlsx(rows, buf)
	}

	return true
}

func (h *HIS) patientInfo(buf *bytes.Buffer, pid string) bool {
	strSql := `select NAME,nvl(SEX,'-'),nvl(PHONE_NUMBER_HOME,'-'),
				case when DATE_OF_BIRTH is not null then  to_char(sysdate,'yyyy')-to_char(DATE_OF_BIRTH,'yyyy') else -1 end as age
	 		  from MEDREC.PAT_MASTER_INDEX where PATIENT_ID = :1`
	strName, strSex, strTel, strAge := "", "", "", ""

	err := h.conn.QueryRow(strSql, pid).Scan(&strName, &strSex, &strTel, &strAge)
	if err != nil {
		log.Println("error:", err)
		buf.WriteString(`{"name":"","sex":"","tel":"","age":""}`)
		return false
	}
	buf.WriteString(`{"name":"` + strName + `","sex":"` + strSex + `","tel":"` + strTel + `","age":"` + strAge + `"}`)
	return true
}

func (h *HIS) docotors(buf *bytes.Buffer) bool {
	strSql := `select USER_NAME as username,NAME as realname from COMM.STAFF_DICT where DEPT_CODE like '01%' order by name`
	rows, err := h.conn.Query(strSql)
	if err != nil {
		log.Println(err)
		return false
	}
	defer rows.Close()

	return dbs.Rows2Json(rows, buf)
}

/*
func (h *HIS) diagnosis(buf *bytes.Buffer) bool {
	strSql := `select DIAGNOSIS_CODE as diagcode,DIAGNOSIS_NAME as diagname from COMM.DIAGNOSIS_DICT order by DIAGNOSIS_NAME`
	rows, err := h.conn.Query(strSql)
	if err != nil {
		log.Println(err)
		return false
	}
	defer rows.Close()

	return dbs.Rows2Json(rows, buf)
}
*/
func (h *HIS) diagnosis(buf *bytes.Buffer, key string) bool {
	strSql := `select DIAGNOSIS_CODE as diagcode,DIAGNOSIS_NAME as diagname from COMM.DIAGNOSIS_DICT where DIAGNOSIS_NAME like '%` + key + `%' order by DIAGNOSIS_NAME`
	rows, err := h.conn.Query(strSql)
	if err != nil {
		log.Println(err)
		return false
	}
	defer rows.Close()

	return dbs.Rows2Json(rows, buf)
}

func (h *HIS) login(user, pwd string) bool {
	strSql := `select comm.F_DESCRIPT(password) as pwd from comm.staff_dict  where user_name = :1`
	strPwd := ""
	err := h.conn.QueryRow(strSql, user).Scan(&strPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	if strings.ToLower(strPwd) == strings.ToLower(pwd) {
		return true
	}

	return false
}
