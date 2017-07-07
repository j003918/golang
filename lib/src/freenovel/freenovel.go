// freenovel project freenovel.go
package freenovel

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-yaml/yaml"
)

type novel struct {
	proxy           string
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
		panic(err)
	}

	err = yaml.Unmarshal(data, &mapYaml)
	if err != nil {
		panic(err)
	}
}

func getBookInfo(bi *bookInfo, nl *novel, noveUrl, proxyUrl string) bool {
	hc := newNovelHttp(proxyUrl)

	buf := &bytes.Buffer{}
	getBodyByUrl(hc, noveUrl, nl.charset, buf)

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
			uu, _ := url.Parse(noveUrl)
			strPreUrl = uu.Scheme + "://" + uu.Host
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
	nitem.proxy = v["proxy"].(string)
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

	if !getBookInfo(&bi, nitem, noveUrl, nitem.proxy) {
		fmt.Println("parse website tag err")
		return false
	}

	f, err := os.Create(bi.name + ".txt")
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()
	hc := newNovelHttp(nitem.proxy)
	req := newNovelRequest("GET", "", "")
	req.Header.Set("Referer", noveUrl)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("DNT", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36")
	buf := &bytes.Buffer{}

	nChapter := len(bi.chtUrlList)
	for i := 0; i < nChapter; i++ {
		func(strTitle, strUrl string) {
			req.URL, _ = url.Parse(strUrl)
			getBodyByReq(hc, req, nitem.charset, buf)
			//getBodyByUrl(hc, strUrl, nitem.charset, buf)
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
