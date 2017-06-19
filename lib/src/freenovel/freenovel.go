// freenovel project freenovel.go
package freenovel

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/crufter/goquery"
)

type novel struct {
	wetsite        string
	searchUrl      string
	charset        string
	novelName      string
	novelMenu      string
	chapterTitle   string
	chapterContent string
	chapterPreUrl  string
}

type bookInfo struct {
	name        string
	chtNameList []string
	chtUrlList  []string
}

var mapNovels map[string]*novel = make(map[string]*novel)
var chtReplacer = strings.NewReplacer("<br>", "\r\n", "<br/>", "\r\n", "<br />", "\r\n")

func init() {
	mapNovels["www.xxbiquge.com"] = &novel{
		wetsite:        "www.xxbiquge.com",
		searchUrl:      "http://zhannei.baidu.com/cse/search?s=8823758711381329060&ie=utf-8&q=",
		charset:        "utf-8",
		novelName:      "#info h1",
		novelMenu:      "#list dl dd a",
		chapterTitle:   "div.bookname h1",
		chapterContent: "#content",
		chapterPreUrl:  "http://www.xxbiquge.com",
	}

	mapNovels["www.zwdu.com"] = &novel{
		wetsite:        "www.zwdu.com",
		searchUrl:      "http://zhannei.baidu.com/cse/search?s=9974397986872341910&ie=gbk&q=",
		charset:        "gbk",
		novelName:      "#info h1",
		novelMenu:      "#list dl dd a",
		chapterTitle:   "div.bookname h1",
		chapterContent: "#content",
		chapterPreUrl:  "http://www.zwdu.com",
	}

	mapNovels["www.23us.com"] = &novel{
		wetsite:        "www.23us.com",
		searchUrl:      "http://zhannei.baidu.com/cse/search?s=9974397986872341910&ie=gbk&q",
		charset:        "gbk",
		novelName:      "div.bdsub dl dd h1",
		novelMenu:      "#at tbody tr td a",
		chapterTitle:   "div.bdsub dl dd",
		chapterContent: "#contents",
		chapterPreUrl:  "",
	}

}

func getPageHtml(strUrl, charset string, hc *http.Client) string {
	rsp, err := hc.Get(strUrl)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer rsp.Body.Close()

	p, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	strBody := string(p)
	switch charset {
	case "gbk":
		strBody = mahonia.NewDecoder("gbk").ConvertString(strBody)
	default:
	}

	return strBody
}

func getBookInfo(bi *bookInfo, nl *novel, noveUrl string) bool {
	strHtml := getPageHtml(noveUrl, nl.charset, &http.Client{})
	if strHtml == "" {
		return false
	}
	dom, err := goquery.ParseString(strHtml)
	if err != nil {
		return false
	}

	bi.name = dom.Find(nl.novelName).Text()
	nodes := dom.Find(nl.novelMenu)
	for i := 0; i < nodes.Length(); i++ {
		v := nodes.Eq(i)
		if v.Text() != "" {
			if nl.chapterPreUrl != "" {
				bi.chtUrlList = append(bi.chtUrlList, nl.chapterPreUrl+v.Attr("href"))
			} else {
				bi.chtUrlList = append(bi.chtUrlList, noveUrl+v.Attr("href"))
			}
			bi.chtNameList = append(bi.chtNameList, v.Text())
		}
	}

	return true
}

func NovelDownload(noveUrl string) bool {
	u, err := url.Parse(noveUrl)
	if err != nil {
		fmt.Println(err)
		return false
	}

	nitem, ok := mapNovels[u.Host]
	if !ok {
		fmt.Println("not supported website:", noveUrl)
		return false
	}

	bi := bookInfo{}
	if !getBookInfo(&bi, nitem, noveUrl) {
		fmt.Println("parse website tag err")
		return false
	}

	hc := &http.Client{}
	f, err := os.Create(bi.name + ".txt")
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()

	for i := 0; i < len(bi.chtUrlList); i++ {
		func(strTitle, strUrl string) {
			str := getPageHtml(strUrl, nitem.charset, hc)
			dom, err := goquery.ParseString(str)
			if err != nil {
				fmt.Println(err)
				return
			}

			strContent := chtReplacer.Replace(dom.Find(nitem.chapterContent).Html())

			if strContent == "" {
				fmt.Println("get charpter error:", strTitle, strUrl)
			}

			f.WriteString(strTitle + "\r\n\r\n")
			f.WriteString(strContent)
			f.WriteString("\r\n")
			fmt.Println(strTitle, strUrl)
			f.Sync()
		}(bi.chtNameList[i], bi.chtUrlList[i])
	}
	return true
}
