// freenovel project freenovel.go
package freenovel

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/go-yaml/yaml"
)

type novel struct {
	wetsite         string
	charset         string
	menuRefer       string
	novelName       string
	menuList        string
	chtRefer        string
	chtTitle        string
	chtContent      string
	chtContentStrip string
}

type bookInfo struct {
	name        string
	chtNameList []string
	chtUrlList  []string
}

var chtReplacer = strings.NewReplacer("<br>", "\r\n", "<br/>", "\r\n", "<br />", "\r\n")
var mapYaml map[interface{}]interface{} = make(map[interface{}]interface{})

func init() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(data, &mapYaml)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
}

func viewSource(strUrl, charset string, outBuf *bytes.Buffer, hc *http.Client, tryCount int) {
	outBuf.Reset()
	nTry := 0
	if tryCount < 1 {
		tryCount = 1
	}
RETRYGET:
	func() {
		rsp, err := hc.Get(strUrl)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer rsp.Body.Close()

		p, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch charset {
		case "gb2312":
			outBuf.WriteString(mahonia.NewDecoder("gbk").ConvertByte(p))
		case "gbk":
			outBuf.WriteString(mahonia.NewDecoder("gbk").ConvertByte(p))
		case "gb18030":
			outBuf.WriteString(mahonia.NewDecoder("gb18030").ConvertByte(p))
		case "utf-16":
			outBuf.WriteString(mahonia.NewDecoder("utf-16").ConvertByte(p))
		default:
			outBuf.Write(p)
		}
	}()

	if outBuf.Len() == 0 && nTry < tryCount {
		nTry += 1
		goto RETRYGET
	}
}

func getBookInfo(bi *bookInfo, nl *novel, noveUrl string) bool {
	hc := &http.Client{}
	buf := &bytes.Buffer{}
	viewSource(noveUrl, nl.charset, buf, hc, 3)

	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		fmt.Println(err)
		return false
	}

	bi.name = doc.Find(nl.novelName).Text()
	nodes := doc.Find(nl.menuList)

	itemCount := nodes.Length()
	if itemCount <= 0 {
		return false
	}

	strPreUrl := ""
	strItemLink := "href"
	if strUrl, ok := nodes.Eq(0).Attr(strItemLink); ok {
		if strUrl[0] == '/' {
			strPreUrl = "http://" + nl.wetsite
		} else {
			urlIdx := strings.LastIndex(noveUrl, "/")
			strPreUrl = noveUrl[0 : urlIdx+1]
		}
	}

	for i := 0; i < itemCount; i++ {
		v := nodes.Eq(i)
		strTitle := v.Text()
		strUrl, _ := v.Attr(strItemLink)
		if strTitle != "" {
			bi.chtUrlList = append(bi.chtUrlList, strPreUrl+strUrl)
			bi.chtNameList = append(bi.chtNameList, strTitle)
		}
	}

	return true
}

func WebsiteList() {
	for k, _ := range mapYaml {
		fmt.Println(k)
	}
}

func NovelDownload(noveUrl string) bool {
	u, err := url.Parse(noveUrl)
	if err != nil {
		fmt.Println(err)
		return false
	}

	v, ok := mapYaml[u.Host].(map[interface{}]interface{})
	if !ok {
		fmt.Println("not supported website:", noveUrl)
		return false
	}

	nitem := &novel{}
	nitem.wetsite = v["wetsite"].(string)
	nitem.charset = v["charset"].(string)
	nitem.menuRefer = v["menuRefer"].(string)
	nitem.novelName = v["novelName"].(string)
	nitem.menuList = v["menuList"].(string)
	nitem.chtRefer = v["chtRefer"].(string)
	nitem.chtTitle = v["chtTitle"].(string)
	nitem.chtContent = v["chtContent"].(string)
	nitem.chtContentStrip = v["chtContentStrip"].(string)

	bi := bookInfo{}

	if !getBookInfo(&bi, nitem, noveUrl) {
		fmt.Println("parse website tag err")
		return false
	}

	hc := &http.Client{}
	buf := &bytes.Buffer{}

	f, err := os.Create(bi.name + ".txt")
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()

	nChapter := len(bi.chtUrlList)
	for i := 0; i < nChapter; i++ {
		func(strTitle, strUrl string) {
			viewSource(strUrl, nitem.charset, buf, hc, 3)
			doc, err := goquery.NewDocumentFromReader(buf)
			if err != nil {
				fmt.Println(err)
				return
			}

			strContentHtml, _ := doc.Find(nitem.chtContent).Html()
			strContent := chtReplacer.Replace(strContentHtml)
			if nitem.chtContentStrip != "" {
				strContent = strings.Replace(strContent, nitem.chtContentStrip, "", -1)
			}

			if strContent == "" {
				fmt.Println("get charpter error:", strTitle, strUrl)
			}

			f.WriteString(strTitle + "\r\n\r\n")
			f.WriteString(strContent)
			f.WriteString("\r\n")
			fmt.Println(i+1, "/", nChapter, strTitle, strUrl)
			f.Sync()
		}(bi.chtNameList[i], bi.chtUrlList[i])
	}

	return true
}
