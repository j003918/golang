// gdi project main.go
package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"tinydb"
)

var (
	mydb       *sql.DB
	myerr      error
	httpPort   string
	mRPT       sync.Map
	strRptCols = `
    select column_name,column_comment
		from information_schema.columns 
		where 1=1
		and table_schema ='czzyy' 
		and table_name = 'jhf_ffxlc' 
        and column_comment <> "";
	`

	strRptFxxlc = `
	select
BRID as "病人ID"
,JZNO as "就诊NO"
,XM as "姓名"
,NL as "年龄"
,SJH as "手机号"
,XL as "学历"
,XBS as "现病史"
,JWS as "既往史"
,SYS_ZGHY as "生育史-总共怀孕"
,SYS_ZYC as "生育史-足月产"
,SYS_ZC as "生育史-早产"
,SYS_RL as "生育史-人流（药流）"
,SYS_NMLC as "生育史-难免流产（胎停）"
,SYS_SHRS as "生育史-生化妊娠"
,SYS_GWY as "生育史-宫外孕"
,LCS_CS as "流产史-次数"
,LCS_LCSJ as "流产史-流产时间"
,LCS_HYTS as "流产史-怀孕天数"
,LCS_YN as "流产史-孕囊"
,LCS_PY as "流产史-胚芽"
,LCS_TX as "流产史-胎心"
,LCS_ZRLC as "流产史-自然流产"
,LCS_SFQG as "流产史-是否清宫"
,LCS_TERST as "流产史-胎儿染色体"
,YJS_ZQ as "月经史-周期"
,YJS_TJ as "月经史-痛经（±）"
,FQSFRSTQK as "夫妻双方染色体情况"
,NFJYCG as "男方精液常规"
,BRMXJB as "本人慢性疾病"
,FMMXJB as "父母慢性疾病"
,MQSFYHY as "目前是/否已怀孕"
,MQSFYY as "目前是/否用药"
,MQYY as "目前用药"
,JCXJS_FSH as "检查性激素-FSH"
,JCXJS_LH as "检查性激素-LH"
,JCXJS_E as "检查性激素-E"
,JCXJS_P as "检查性激素-P"
,JCXJS_T as "检查性激素-T"
,JCXJS_PRL as "检查性激素-PRL"
,YDS_0 as "胰岛素-0"
,YDS_60 as "胰岛素-60"
,YDS_120 as "胰岛素-120"
,YDS_180 as "胰岛素-180"
,PTT_0 as "葡萄糖-0"
,PTT_60 as "葡萄糖-60"
,PTT_120 as "葡萄糖-120"
,PTT_180 as "葡萄糖-180"
,JZXGN_K_TPO as "甲状腺功能-抗-TPO"
,JZXGN_TSH as "甲状腺功能-TSH"
,JZXGN_YLT3 as "甲状腺功能-游离T3"
,JZXGN_YLT4 as "甲状腺功能-游离T4"
,JZXGN_JZXQDB as "甲状腺功能-甲状腺球蛋白"
,JZXGN_CJZXSSTKT as "甲状腺功能-促甲状腺素受体抗体"
,JZXGN_KJZXQDBKT as "甲状腺功能-抗甲状腺球蛋白抗体"
,LZZHZ_KXLZKTIgG as "磷脂综合征-抗心磷脂抗体IgG"
,LZZHZ_KXLZKTIgM as "磷脂综合征-抗心磷脂抗体IgM"
,LZZHZ_K2TDBIKT as "磷脂综合征-抗β2糖蛋白I抗体"
,TXBZAS as "同型半胱氨酸"
,XXBJJL_ZDJJL_ADP as "血小板聚集率-最大聚集率(ADP)"
,XXBJJL_ZDJJL_AA as "血小板聚集率-最大聚集率(AA)"
,D_EJT as "D-二聚体"
,XCG_PT as "血常规-PT"
,XCG_PT_INR as "血常规-PT-INR"
,XCG_FIB as "血常规-FIB"
,XCG_APTT as "血常规-APTT"
,XCG_TT as "血常规-TT"
,XCG_AT as "血常规-AT-Ш"
,XCG_PC as "血常规-PC"
,XCG_WBC as "血常规-WBC"
,XCG_RBC as "血常规-RBC"
,XCG_HGB as "血常规-HGB"
,XCG_PLT as "血常规-PLT"
,XCG_PCT as "血常规-PCT"
,ENA_KHKTDL as "ENA-抗核抗体定量"
,ENA_KSLDNADL as "ENA-抗双链DNA定量"
,ENA_KSSAKTDL as "ENA-抗SSA抗体定量"
,ENA_KSSBKTDL as "ENA-抗SSB抗体定量"
,ENA_KJO_1KTDL as "ENA-抗JO-1抗体定量"
,ENA_KSmKTDL as "ENA-抗Sm抗体定量"
,ENA_KnRNPKTDL as "ENA-抗nRNP抗体定量"
,ENA_KScL_70KTDL as "ENA-抗ScL-70抗体定量"
,ENA_KZSDKTDL as "ENA-抗着丝点抗体定量"
,ENA_ZDBKTDL as "ENA-组蛋白抗体定量"
,ZGDMZL_ZCPSV as "子宫动脉阻力-左侧PSV"
,ZGDMZL_ZCEDV as "子宫动脉阻力-左侧EDV"
,ZGDMZL_ZCSD as "子宫动脉阻力-左侧S/D"
,ZGDMZL_ZCRI as "子宫动脉阻力-左侧RI"
,ZGDMZL_YCPSV as "子宫动脉阻力-右侧PSV"
,ZGDMZL_YCEDV as "子宫动脉阻力-右侧EDV"
,ZGDMZL_YCSD as "子宫动脉阻力-右侧S/D"
,ZGDMZL_YCRI as "子宫动脉阻力-右侧RI"
,BMI as "BMI(kg/m*m)"
,DM as "多毛（±）"
,HJP as "黑棘皮（±）"
,ST as "舌苔"
,MX as "脉象"
,MKZH as "目眶黯黑"
,MS as "面色"
,YS as "腰酸"
,SPFL as "神疲乏力"
,TYEM as "头晕耳鸣"
,YXK as "有血块"
from jhf_ffxlc
	`
	strTabFfxlc = `
	CREATE TABLE IF NOT EXISTS jhf_ffxlc(
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

	mRPT.Store("ffxlc", strRptFxxlc)
	mRPT.Store("cols", strRptCols)

	httpPort = ":" + strPort
	if strDSN == "" {
		strDSN = `jhf:jhf@tcp(130.1.10.230:3306)/czzyy`
	}

	mydb, myerr = tinydb.OpenDb(10, "mysql", strDSN, 3, 1)
	if myerr != nil {
		fmt.Println(myerr)
		panic(myerr)
	}
	tinydb.ModifyTab(5, mydb, strTabFfxlc)
}

func gdi(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	strCols := "_PRIV_"
	strVals := "1"
	strSql := "insert into jhf_ffxlc"
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
		rowCount, _ := tinydb.ModifyTab(30, mydb, strSql)

		if rowCount != 1 {
			strRspHtml = "error"
			goto RST
		}

		if rowCount == 1 {
			strRspHtml = "添加成功!"
			goto RST
		}
	}

RST:
	strRspHtml = `<html><body> <div style="text-align:center;"><br><br><br><br><br><br>` + strRspHtml + `<br><a href="/">返回</a></div></body></html>`
	w.Write([]byte(strRspHtml))
}

func expXlsx(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	strSql, ok := mRPT.Load(strings.ToLower(r.FormValue("rpt")))

	if !ok {
		w.Write([]byte("id error"))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.ms-excel; charset=utf-8") //application/vnd.ms-excel or application/x-xls
	w.Header().Set("Content-Disposition", "attachment;filename=report.xlsx")
	w.Header().Set("Connection", "close")
	w.Header().Set("CharacterEncoding", "utf-8")

	tinydb.Sql2Xlsx(30, mydb, strSql.(string), "", w)
}

func js(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	strSql, ok := mRPT.Load(strings.ToLower(r.FormValue("rpt")))

	if !ok {
		w.Write([]byte("id error"))
		return
	}

	var bufdata bytes.Buffer
	tinydb.SQL2Json(10, mydb, strSql.(string), &bufdata)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Connection", "close")
	w.Header().Set("CharacterEncoding", "utf-8")
	w.Write(bufdata.Bytes())
}

func rptlist(w http.ResponseWriter, r *http.Request) {
	mRPT.Range(func(k, v interface{}) bool {
		w.Write([]byte(k.(string) + ";"))
		return true
	})
}

func rptadd(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	strRpt := strings.ToLower(r.FormValue("rpt"))
	strSql := r.FormValue("sql")

	if strRpt != "" && strSql != "" {
		mRPT.Store(strRpt, strSql)
	}
}

func rptdel(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	strRpt := strings.ToLower(r.FormValue("rpt"))
	if strRpt != "" {
		mRPT.Delete(strRpt)
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./html/")))
	http.HandleFunc("/gdi", gdi)
	http.HandleFunc("/js", js)
	http.HandleFunc("/exp", expXlsx)
	http.HandleFunc("/rptlist", rptlist)
	http.HandleFunc("/rptadd", rptadd)
	http.HandleFunc("/rptdel", rptdel)
	fmt.Println(http.ListenAndServe(httpPort, nil))
}
