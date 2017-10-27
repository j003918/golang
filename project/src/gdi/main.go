// gdi project main.go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var (
	mydb        *sql.DB
	myerr       error
	httpPort    string
	strHtml_TAG = `
    select #column_name,column_comment,
		case when column_name not in('XBS','JWS')  then
			CONCAT('<dl><dt>',column_comment,'</dt><dd><input type="text" name="',column_name,'" value="" /></dd></dl>')
		else 
			CONCAT('<dl><dt>',column_comment,'</dt><dd><textarea  name="',column_name,'" rows="5" cols=60 value="" ></textarea></dd></dl>')
		end as HTML_TAG
		from information_schema.columns where table_schema ='oa0618' and table_name = 'JHF_GDI' and column_comment <> "";
	`
	strTab_GDI = `
	CREATE TABLE IF NOT EXISTS JHF_GDI(
		ID INT UNSIGNED NOT NULL AUTO_INCREMENT,
		#-------
		BRID VARCHAR(50) COMMENT '病人ID',
		JZNO VARCHAR(50) COMMENT '就诊NO',
		XM VARCHAR(50) COMMENT '姓名',
		NL VARCHAR(50) COMMENT '年龄',
		SJH VARCHAR(50) COMMENT '手机号',
		XL VARCHAR(50) COMMENT '学历',
		XBS VARCHAR(1000) COMMENT '现病史',
		JWS VARCHAR(1000) COMMENT '既往史',
		SYS_ZGHY VARCHAR(50) COMMENT '生育史-总共怀孕',
		SYS_ZYC VARCHAR(50) COMMENT '生育史-足月产',
		SYS_ZC VARCHAR(50) COMMENT '生育史-早产',
		SYS_RL VARCHAR(50) COMMENT '生育史-人流（药流）',
		SYS_NMLC VARCHAR(50) COMMENT '生育史-难免流产（胎停）',
		SYS_SHRS VARCHAR(50) COMMENT '生育史-生化妊娠',
		SYS_GWY VARCHAR(50) COMMENT '生育史-宫外孕',
		LCS_CS VARCHAR(50) COMMENT '流产史-次数',
		LCS_LCSJ VARCHAR(50) COMMENT '流产史-流产时间',
		LCS_HYTS VARCHAR(50) COMMENT '流产史-怀孕天数',
		LCS_YN VARCHAR(50) COMMENT '流产史-孕囊',
		LCS_PY VARCHAR(50) COMMENT '流产史-胚芽',
		LCS_TX VARCHAR(50) COMMENT '流产史-胎心',
		LCS_ZRLC VARCHAR(50) COMMENT '流产史-自然流产',
		LCS_SFQG VARCHAR(50) COMMENT '流产史-是否清宫',
		LCS_TERST VARCHAR(50) COMMENT '流产史-胎儿染色体',
		YJS_ZQ VARCHAR(50) COMMENT '月经史-周期',
		YJS_TJ VARCHAR(50) COMMENT '月经史-痛经（±）',
		FQSFRSTQK VARCHAR(50) COMMENT '夫妻双方染色体情况',
		NFJYCG VARCHAR(50) COMMENT '男方精液常规',
		BRMXJB VARCHAR(50) COMMENT '本人慢性疾病',
		FMMXJB VARCHAR(50) COMMENT '父母慢性疾病',
		MQSFYHY VARCHAR(50) COMMENT '目前是/否已怀孕',
		MQSFYY VARCHAR(50) COMMENT '目前是/否用药',
		MQYY VARCHAR(50) COMMENT '目前用药',
		JCXJS_FSH VARCHAR(50) COMMENT '检查性激素-FSH',
		JCXJS_LH VARCHAR(50) COMMENT '检查性激素-LH',
		JCXJS_E VARCHAR(50) COMMENT '检查性激素-E',
		JCXJS_P VARCHAR(50) COMMENT '检查性激素-P',
		JCXJS_T VARCHAR(50) COMMENT '检查性激素-T',
		JCXJS_PRL VARCHAR(50) COMMENT '检查性激素-PRL',
		YDS_0 VARCHAR(50) COMMENT '胰岛素-0',
		YDS_60 VARCHAR(50) COMMENT '胰岛素-60',
		YDS_120 VARCHAR(50) COMMENT '胰岛素-120',
		YDS_180 VARCHAR(50) COMMENT '胰岛素-180',
		PTT_0 VARCHAR(50) COMMENT '葡萄糖-0',
		PTT_60 VARCHAR(50) COMMENT '葡萄糖-60',
		PTT_120 VARCHAR(50) COMMENT '葡萄糖-120',
		PTT_180 VARCHAR(50) COMMENT '葡萄糖-180',
		JZXGN_K_TPO VARCHAR(50) COMMENT '甲状腺功能-抗-TPO',
		JZXGN_TSH VARCHAR(50) COMMENT '甲状腺功能-TSH',
		JZXGN_YLT3 VARCHAR(50) COMMENT '甲状腺功能-游离T3',
		JZXGN_YLT4 VARCHAR(50) COMMENT '甲状腺功能-游离T4',
		JZXGN_JZXQDB VARCHAR(50) COMMENT '甲状腺功能-甲状腺球蛋白',
		JZXGN_CJZXSSTKT VARCHAR(50) COMMENT '甲状腺功能-促甲状腺素受体抗体',
		JZXGN_KJZXQDBKT VARCHAR(50) COMMENT '甲状腺功能-抗甲状腺球蛋白抗体',
		LZZHZ_KXLZKTIgG VARCHAR(50) COMMENT '磷脂综合征-抗心磷脂抗体IgG',
		LZZHZ_KXLZKTIgM VARCHAR(50) COMMENT '磷脂综合征-抗心磷脂抗体IgM',
		LZZHZ_K2TDBIKT VARCHAR(50) COMMENT '磷脂综合征-抗β2糖蛋白I抗体',
		TXBZAS VARCHAR(50) COMMENT '同型半胱氨酸',
		XXBJJL_ZDJJL_ADP VARCHAR(50) COMMENT '血小板聚集率-最大聚集率(ADP)',
		XXBJJL_ZDJJL_AA VARCHAR(50) COMMENT '血小板聚集率-最大聚集率(AA)',
		D_EJT VARCHAR(50) COMMENT 'D-二聚体',
		XCG_PT VARCHAR(50) COMMENT '血常规-PT',
		XCG_PT_INR VARCHAR(50) COMMENT '血常规-PT-INR',
		XCG_FIB VARCHAR(50) COMMENT '血常规-FIB',
		XCG_APTT VARCHAR(50) COMMENT '血常规-APTT',
		XCG_TT VARCHAR(50) COMMENT '血常规-TT',
		XCG_AT VARCHAR(50) COMMENT '血常规-AT-Ш',
		XCG_PC VARCHAR(50) COMMENT '血常规-PC',
		XCG_WBC VARCHAR(50) COMMENT '血常规-WBC',
		XCG_RBC VARCHAR(50) COMMENT '血常规-RBC',
		XCG_HGB VARCHAR(50) COMMENT '血常规-HGB',
		XCG_PLT VARCHAR(50) COMMENT '血常规-PLT',
		XCG_PCT VARCHAR(50) COMMENT '血常规-PCT',
		ENA_KHKTDL VARCHAR(50) COMMENT 'ENA-抗核抗体定量',
		ENA_KSLDNADL VARCHAR(50) COMMENT 'ENA-抗双链DNA定量',
		ENA_KSSAKTDL VARCHAR(50) COMMENT 'ENA-抗SSA抗体定量',
		ENA_KSSBKTDL VARCHAR(50) COMMENT 'ENA-抗SSB抗体定量',
		ENA_KJO_1KTDL VARCHAR(50) COMMENT 'ENA-抗JO-1抗体定量',
		ENA_KSmKTDL VARCHAR(50) COMMENT 'ENA-抗Sm抗体定量',
		ENA_KnRNPKTDL VARCHAR(50) COMMENT 'ENA-抗nRNP抗体定量',
		ENA_KScL_70KTDL VARCHAR(50) COMMENT 'ENA-抗ScL-70抗体定量',
		ENA_KZSDKTDL VARCHAR(50) COMMENT 'ENA-抗着丝点抗体定量',
		ENA_ZDBKTDL VARCHAR(50) COMMENT 'ENA-组蛋白抗体定量',
		ZGDMZL_ZCPSV VARCHAR(50) COMMENT '子宫动脉阻力-左侧PSV',
		ZGDMZL_ZCEDV VARCHAR(50) COMMENT '子宫动脉阻力-左侧EDV',
		ZGDMZL_ZCSD VARCHAR(50) COMMENT '子宫动脉阻力-左侧S/D',
		ZGDMZL_ZCRI VARCHAR(50) COMMENT '子宫动脉阻力-左侧RI',
		ZGDMZL_YCPSV VARCHAR(50) COMMENT '子宫动脉阻力-右侧PSV',
		ZGDMZL_YCEDV VARCHAR(50) COMMENT '子宫动脉阻力-右侧EDV',
		ZGDMZL_YCSD VARCHAR(50) COMMENT '子宫动脉阻力-右侧S/D',
		ZGDMZL_YCRI VARCHAR(50) COMMENT '子宫动脉阻力-右侧RI',
		BMI VARCHAR(50) COMMENT 'BMI(kg/m*m)',
		DM VARCHAR(50) COMMENT '多毛（±）',
		HJP VARCHAR(50) COMMENT '黑棘皮（±）',
		ST VARCHAR(50) COMMENT '舌苔',
		MX VARCHAR(50) COMMENT '脉象',
		MKZH VARCHAR(50) COMMENT '目眶黯黑',
		MS VARCHAR(50) COMMENT '面色',
		YS VARCHAR(50) COMMENT '腰酸',
		SPFL VARCHAR(50) COMMENT '神疲乏力',
		TYEM VARCHAR(50) COMMENT '头晕耳鸣',
		YXK VARCHAR(50) COMMENT '有血块',
		#-------
		CREATE_TIME TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
		#UPDATE_TIME TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		_PRIV_ INT,
		PRIMARY KEY (ID)
	);
	`
)

func init() {

	strPort := *(flag.String("port", "8080", ""))
	strDSN := *(flag.String("dsn", "", ""))
	flag.Parse()

	httpPort = ":" + strPort
	if strDSN == "" {
		strDSN = `admin:admin@tcp(172.25.125.101:3306)/oa0618`
	}

	mydb, myerr = sql.Open("mysql", strDSN)
	if myerr != nil {
		panic(myerr)
	}

	mydb.SetMaxOpenConns(3)
	mydb.SetMaxIdleConns(1)

	myerr = mydb.Ping()
	if myerr != nil {
		panic(myerr)
	}

	ctTable()
}

func ctTable() {
	_, err := mydb.Exec(strTab_GDI)
	if err != nil {
		panic(err)
	}
}

func gdi(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	strCols := "_PRIV_"
	strVals := "1"
	strSql := "insert into JHF_GDI"
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
		rst, err := mydb.Exec(strSql)

		if err != nil {
			//panic(err)
			strRspHtml = err.Error()
			//fmt.Println(err)
			goto RST
		}

		rowCount, _ := rst.RowsAffected()
		if rowCount == 1 {
			strRspHtml = "添加成功!"
			//w.Write([]byte("add ok"))
			//return
			goto RST
		}
	}
RST:
	strRspHtml = `<html><body> <div style="text-align:center;"><br><br><br><br><br><br>` + strRspHtml + `<br><a href="/">返回</a></div></body></html>`
	w.Write([]byte(strRspHtml))
}

func prt_html() {
	rows, err := mydb.Query(strHtml_TAG)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	//var arrHTML [6]string
	strHTML := ""
	//cnt := 0

	for rows.Next() {
		err = rows.Scan(&strHTML)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		//arrHTML[cnt%6] += strHTML
		//cnt += 1

		fmt.Println(strHTML)
	}
	/*
			arrHTML[0] = `<div class="leftbar"><dl>` + arrHTML[0] + `</dl></div>`
			arrHTML[1] = `<div class="leftbar"><dl>` + arrHTML[1] + `</dl></div>`
			arrHTML[2] = `<div class="leftbar"><dl>` + arrHTML[2] + `</dl></div>`
			arrHTML[3] = `<div class="leftbar"><dl>` + arrHTML[3] + `</dl></div>`
			arrHTML[4] = `<div class="leftbar"><dl>` + arrHTML[4] + `</dl></div>`
			arrHTML[5] = `<div class="leftbar"><dl>` + arrHTML[5] + `</dl></div>`

		fmt.Println(arrHTML[0])
		fmt.Println(arrHTML[1])
		fmt.Println(arrHTML[2])
		fmt.Println(arrHTML[3])
		fmt.Println(arrHTML[4])
		fmt.Println(arrHTML[5])
	*/
}

func main() {
	prt_html()
	http.Handle("/", http.FileServer(http.Dir("./html/")))
	http.HandleFunc("/gdi", gdi)
	fmt.Println(http.ListenAndServe(httpPort, nil))
}
