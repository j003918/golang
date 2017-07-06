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
)

type novel struct {
	wetsite         string
	charset         string
	menuRefer       string
	noveName        string
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

var mapNovel map[string]*novel = make(map[string]*novel)
var chtReplacer = strings.NewReplacer("<br>", "\r\n", "<br/>", "\r\n", "<br />", "\r\n")

func init() {
	mapNovel["www.xxbiquge.com"] = &novel{
		wetsite:         "www.xxbiquge.com",
		charset:         "utf-8",
		menuRefer:       "",
		noveName:        "#info h1",
		menuList:        "#list dl dd a",
		chtRefer:        "",
		chtTitle:        "div.bookname h1",
		chtContent:      "#content",
		chtContentStrip: "",
	}

	mapNovel["www.zwdu.com"] = &novel{
		wetsite:         "www.zwdu.com",
		charset:         "gbk",
		menuRefer:       "",
		noveName:        "#info h1",
		menuList:        "#list dl dd a",
		chtRefer:        "",
		chtTitle:        "div.bookname h1",
		chtContent:      "#content",
		chtContentStrip: "",
	}

	mapNovel["www.23us.com"] = &novel{
		wetsite:         "www.23us.com",
		charset:         "gbk",
		menuRefer:       "",
		noveName:        "div.bdsub dl dd h1",
		menuList:        "#at tbody tr td a",
		chtRefer:        "",
		chtTitle:        "div.bdsub dl dd",
		chtContent:      "#contents",
		chtContentStrip: "顶点小说 ２３ＵＳ．ＣＯＭ更新最快",
	}

	mapNovel["www.88dushu.com"] = &novel{
		wetsite:         "www.88dushu.com",
		charset:         "gbk",
		menuRefer:       "",
		noveName:        "div.rt h1",
		menuList:        "div.mulu ul li a",
		chtRefer:        "",
		chtTitle:        "div.novel h1",
		chtContent:      "div.yd_text2",
		chtContentStrip: "",
	}

	mapNovel["www.qu.la"] = &novel{
		wetsite:         "www.qu.la",
		charset:         "utf-8",
		menuRefer:       "",
		noveName:        "#info h1",
		menuList:        "#list dl dd a",
		chtRefer:        "",
		chtTitle:        "div.bookname h1",
		chtContent:      "#content",
		chtContentStrip: "<script>chaptererror();</script>",
	}

	mapNovel["www.biqudao.com"] = &novel{
		wetsite:         "www.biqudao.com",
		charset:         "utf-8",
		menuRefer:       "",
		noveName:        "#info h1",
		menuList:        "#list dl dd a",
		chtRefer:        "",
		chtTitle:        "div.bookname h1",
		chtContent:      "#content",
		chtContentStrip: "",
	}

	mapNovel["www.shoujikanshu.org"] = &novel{
		wetsite:         "www.shoujikanshu.org",
		charset:         "gb2312",
		menuRefer:       "",
		noveName:        "div.box-artic h1",
		menuList:        "div.list li a",
		chtRefer:        "",
		chtTitle:        "div.subNav h1",
		chtContent:      "div.content",
		chtContentStrip: "",
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

	bi.name = doc.Find(nl.noveName).Text()
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

func NovelDownload(noveUrl string) bool {
	u, err := url.Parse(noveUrl)
	if err != nil {
		fmt.Println(err)
		return false
	}

	nitem, ok := mapNovel[u.Host]
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
