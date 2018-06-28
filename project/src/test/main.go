// test project main.go
package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"os"
	"time"
	//	"crypto/md5"
	//	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"
	"syscall"
	"unsafe"
	//"loadbalance"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func mssqlTest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := &bytes.Buffer{}
	mssqlInfo(buf)

	w.Header().Set("Content-type", "application/json;charset=utf-8")
	w.Write(buf.Bytes())
}

func mysqlTest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := &bytes.Buffer{}
	mysqlInfo(buf)

	w.Header().Set("Content-type", "application/json;charset=utf-8")
	w.Write(buf.Bytes())
}

func oracleTest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buf := &bytes.Buffer{}
	err := oracleInfo(buf)
	if err == nil {
		w.Header().Set("Content-type", "application/json;charset=utf-8")
		w.Write(buf.Bytes())
	} else {
		w.Write([]byte(err.Error()))
	}

}

//var lb *loadbalance.LB

func testdll() {
	Handle := syscall.NewLazyDLL("netdll.dll")
	md5_encode := Handle.NewProc("md5_encode")
	str := `bizid=Q2MrhC&cmdid=scancodepay&nonce=0266e33d3f546cb5436a10798e657d97&req={"recordId":"2018051033420","businessType":"2","bizOutTradeNo":"2018051033420001","buyerId":"2018051033420001","body":"","fee":"14.2","goodsName":"","hosFlag":"XMCS2","goodsId":"2018051033420001","notifyUrl":"sss"}&bizkey=BPsnsnY4uM87WYV9YyHEkUX0bY4ZGOBA`
	var a [33]byte

	r, _, _ := md5_encode.Call(uintptr(unsafe.Pointer(syscall.StringBytePtr(str))), uintptr(unsafe.Pointer(&a[0])))

	//syscall.StringBytePtr()

	fmt.Println(r, "OK", a)
}

func enc_gbk(str string) {
	fmt.Println("enc_gbk:", str)

	data, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(str)), simplifiedchinese.GBK.NewEncoder()))
	//	fmt.Printf("utf8: %02X\n", str)
	//	fmt.Printf("GBK: %02X\n", data)

	for _, v := range []byte(str) {
		fmt.Printf("0x%02X,", v)
	}
	fmt.Printf("\n")

	for _, v := range data {
		fmt.Printf("0x%02X,", v)
	}
	fmt.Printf("\n")

}

func zhybdz(date string) string {

	strUrl := "http://10.76.24.121:8114/agent/reciveFromHos"
	strReq := `{"operDate":"`
	strReq += date
	strReq += `","hospitalID":"0010","operationID":"9101"}`

	//fmt.Println(strUrl, strReq)

	resp, err := http.Post(strUrl, "application/json", strings.NewReader(strReq))

	if err != nil {
		fmt.Println(err)
		return "NA"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "NA"
	}

	//value := gjson.Get(string(body), "name.last")

	//jsonparser.

	//strTest := `{"ownCost":"494.56","appmsg":"null","positiveCount":"47","totCost":"2508.73","appcode":0,"negativeCount":"18"}`
	strRst, strTmp := "", ""
	intRst, _ := jsonparser.GetInt(body, "appcode")

	if intRst != 0 {
		return "NONE"
	}

	strTmp, _ = jsonparser.GetString(body, "totCost")
	strRst += strTmp + " "

	strTmp, _ = jsonparser.GetString(body, "ownCost")
	strRst += strTmp + " "

	strTmp, _ = jsonparser.GetString(body, "positiveCount")
	strRst += strTmp + " "

	strTmp, _ = jsonparser.GetString(body, "negativeCount")
	strRst += strTmp
	return strRst
}

func parsOrders() {

	fi, err := os.Open("C:/Users/b/Desktop/orders.json")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	//update t_clinic_fee_order_temp set bill_no = '' where order_no = '';
	strac, strfid, strbill, strSql := "", "", "", ""
	for {
		strac, strfid, strbill, strSql = "", "", "", ""

		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}

		a1, _, c1 := br.ReadLine()
		if c1 == io.EOF {
			break
		}

		strac, _ = jsonparser.GetString(a1, "appcode")

		if strac == "-1" {
			continue
		}

		strfid, _ = jsonparser.GetString(a, "feeId")
		strbill, _ = jsonparser.GetString(a, "billNo")
		//strbal, _ = jsonparser.GetString(a1, "balId")

		strSql = `update t_clinic_fee_order_temp set bill_no = '` + strbill + `' where order_no = '` + strfid + `';`

		fmt.Println(strSql)
	}
}

func main() {

	parsOrders()

	st := flag.String("st", time.Now().Format("2006-01-02"), "")
	et := flag.String("et", time.Now().Format("2006-01-02"), "")
	flag.Parse()

	//fmt.Println(*st, *et)

	//	toBeCharge := "2015-01-01"           //待转化为时间戳的字符串 注意 这里的小时和分钟还要秒必须写 因为是跟着模板走的 修改模板的话也可以不写
	timeLayout := "2006-01-02"           //转化所需模板
	loc, _ := time.LoadLocation("Local") //重要：获取时区
	tSt, _ := time.ParseInLocation(timeLayout, *st, loc)
	tEt, _ := time.ParseInLocation(timeLayout, *et, loc)

	dd, _ := time.ParseDuration("24h")
	strRQ := ""
	for tSt.Unix() <= tEt.Unix() {
		strRQ = tSt.Format(timeLayout)
		fmt.Println(strRQ, zhybdz(strRQ))
		tSt = tSt.Add(dd)
		//fmt.Println()
	}

	//	fmt.Println(theTime)

	//fmt.Println(time.Now().Format("2006-01-02"))

	/*
		var testaa = [10]uint16{
			//3: 10,
			//5: 6,
			3,
			5,
		}

		fmt.Println(testaa)

		str := `bizid=Q2MrhC&cmdid=scancodepay&nonce=0266e33d3f546cb5436a10798e657d97&req={"recordId":"2018051033420","businessType":"2","bizOutTradeNo":"2018051033420001","buyerId":"2018051033420001","body":"","fee":"14.2","goodsName":"","hosFlag":"XMCS2","goodsId":"2018051033420001","notifyUrl":"sss"}&bizkey=BPsnsnY4uM87WYV9YyHEkUX0bY4ZGOBA`
		//str := `admin`
		h := md5.New()
		h.Write([]byte(str)) // 需要加密的字符串为 123456
		cipherStr := h.Sum(nil)
		fmt.Println(cipherStr)
		fmt.Printf("%s\n", hex.EncodeToString(cipherStr)) // 输出加密结果
		//testdll()

		//fmt.Println("test rst:", []byte("是αabcdβ我"))//中文GBK显示

		strEnc := "是αabcdβ我" //yte(strEnc)) //

		enc_gbk(strEnc)

		//	lb = loadbalance.NewLB()
		//	lb.Register("http", "130.1.10.230:8080")
		//  lb.Register("http", "130.1.10.230")

		router := httprouter.New()
		router.GET("/mssql", mssqlTest)
		router.GET("/mysql", mysqlTest)
		router.GET("/oracle", oracleTest)
		router.GET("/user", oracleTest)

		fmt.Println(http.ListenAndServe(":8080", router))
		//	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { lb.Proxy(w, r) })
		//	fmt.Println(http.ListenAndServe(":8080", nil))
	*/
}
