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

type websiteInfo struct {
	wetsite     string
	charset     string
	searchUrl   string
	searchTag   string
	seaTitleTag string
	seaUrlTag   string
}

type novelMenuTag struct {
	refererUrl string
	name       string
	list       string
	isVol      bool
	volume     string
	item       string
	preUrl     string
	itemLink   string
}

type novelChapterTag struct {
	refererUrl  string
	title       string
	content     string
	nextChapter string
	strip       string
}

type Noveler struct {
	wsi websiteInfo
	nmt novelMenuTag
	nct novelChapterTag
}

type bookInfo struct {
	name        string
	chtNameList []string
	chtUrlList  []string
}

var mapNoveler map[string]*Noveler = make(map[string]*Noveler)
var chtReplacer = strings.NewReplacer("<br>", "\r\n", "<br/>", "\r\n", "<br />", "\r\n")

func init() {
	mapNoveler["www.xxbiquge.com"] = &Noveler{
		websiteInfo{
			wetsite:     "www.xxbiquge.com",
			charset:     "utf-8",
			searchUrl:   "http://zhannei.baidu.com/cse/search?s=8823758711381329060&q=",
			searchTag:   "div.result-game-item-detail h3 a",
			seaTitleTag: "title",
			seaUrlTag:   "href",
		},
		novelMenuTag{
			refererUrl: "",
			name:       "#info h1",
			list:       "#list dl",
			isVol:      false,
			volume:     "",
			item:       "dd a",
			preUrl:     "http://www.xxbiquge.com",
			itemLink:   "href",
		},
		novelChapterTag{
			refererUrl:  "",
			title:       "div.bookname h1",
			content:     "#content",
			nextChapter: "",
			strip:       "",
		},
	}

	mapNoveler["www.zwdu.com"] = &Noveler{
		websiteInfo{
			wetsite:     "www.zwdu.com",
			charset:     "gbk",
			searchUrl:   "http://zhannei.baidu.com/cse/search?s=9974397986872341910&q=",
			searchTag:   "div.result-game-item-detail h3 a",
			seaTitleTag: "title",
			seaUrlTag:   "href",
		},
		novelMenuTag{
			refererUrl: "",
			name:       "#info h1",
			list:       "#list dl",
			isVol:      false,
			volume:     "",
			item:       "dd a",
			preUrl:     "http://www.zwdu.com",
			itemLink:   "href",
		},
		novelChapterTag{
			refererUrl:  "",
			title:       "div.bookname h1",
			content:     "#content",
			nextChapter: "",
			strip:       "",
		},
	}

	mapNoveler["www.23us.com"] = &Noveler{
		websiteInfo{
			wetsite:     "www.23us.com",
			charset:     "gbk",
			searchUrl:   "http://zhannei.baidu.com/cse/search?s=8253726671271885340&q=",
			searchTag:   "div.result-game-item-detail h3 a",
			seaTitleTag: "title",
			seaUrlTag:   "href",
		},
		novelMenuTag{
			refererUrl: "",
			name:       "div.bdsub dl dd h1",
			list:       "#at tbody tr ",
			isVol:      false,
			volume:     "",
			item:       "td a",
			preUrl:     "",
			itemLink:   "href",
		},
		novelChapterTag{
			refererUrl:  "",
			title:       "div.bdsub dl dd",
			content:     "#contents",
			nextChapter: "",
			strip:       "顶点小说 ２３ＵＳ．ＣＯＭ更新最快",
		},
	}

	mapNoveler["www.88dushu.com"] = &Noveler{
		websiteInfo{
			wetsite:     "www.88dushu.com",
			charset:     "gbk",
			searchUrl:   "http://zn.88dushu.com/cse/search?s=2308740887988514756&q=",
			searchTag:   "div.result-game-item-detail h3 a",
			seaTitleTag: "title",
			seaUrlTag:   "href",
		},
		novelMenuTag{
			refererUrl: "",
			name:       "div.rt h1",
			list:       "div.mulu ul",
			isVol:      false,
			volume:     "",
			item:       "li a",
			preUrl:     "",
			itemLink:   "href",
		},
		novelChapterTag{
			refererUrl:  "",
			title:       "div.novel h1",
			content:     "div.yd_text2",
			nextChapter: "",
			strip:       "",
		},
	}

	mapNoveler["www.qu.la"] = &Noveler{
		websiteInfo{
			wetsite:     "www.qu.la",
			charset:     "utf-8",
			searchUrl:   "http://zhannei.baidu.com/cse/search?s=920895234054625192&q=",
			searchTag:   "div.result-game-item-detail h3 a",
			seaTitleTag: "title",
			seaUrlTag:   "href",
		},
		novelMenuTag{
			refererUrl: "",
			name:       "#info h1",
			list:       "#list dl",
			isVol:      false,
			volume:     "",
			item:       "dd a",
			preUrl:     "http://www.qu.la",
			itemLink:   "href",
		},
		novelChapterTag{
			refererUrl:  "",
			title:       "div.bookname h1",
			content:     "#content",
			nextChapter: "",
			strip:       "<script>chaptererror();</script>",
		},
	}

	mapNoveler["www.biqudao.com"] = &Noveler{
		websiteInfo{
			wetsite:     "www.biqudao.com",
			charset:     "utf-8",
			searchUrl:   "http://zhannei.baidu.com/cse/search?s=3654077655350271938&entry=1&q=",
			searchTag:   "div.result-game-item-detail h3 a",
			seaTitleTag: "title",
			seaUrlTag:   "href",
		},
		novelMenuTag{
			refererUrl: "",
			name:       "#info h1",
			list:       "#list dl",
			isVol:      false,
			volume:     "",
			item:       "dd a",
			preUrl:     "http://www.biqudao.com",
			itemLink:   "href",
		},
		novelChapterTag{
			refererUrl:  "",
			title:       "div.bookname h1",
			content:     "#content",
			nextChapter: "",
			strip:       "",
		},
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

func getBookInfo(bi *bookInfo, nl *Noveler, noveUrl string) bool {
	hc := &http.Client{}
	buf := &bytes.Buffer{}
	viewSource(noveUrl, nl.wsi.charset, buf, hc, 3)

	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		fmt.Println(err)
		return false
	}

	bi.name = doc.Find(nl.nmt.name).Text()
	nodes := doc.Find(nl.nmt.list).Find(nl.nmt.item)

	if nl.nmt.preUrl == "" {
		nl.nmt.preUrl = noveUrl
	}

	strItemLink := nl.nmt.itemLink

	for i := 0; i < nodes.Length(); i++ {
		v := nodes.Eq(i)
		strTitle := v.Text()
		strUrl, _ := v.Attr(strItemLink)
		if strTitle != "" {
			bi.chtUrlList = append(bi.chtUrlList, nl.nmt.preUrl+strUrl)
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

	nitem, ok := mapNoveler[u.Host]
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

	for i := 0; i < len(bi.chtUrlList); i++ {
		func(strTitle, strUrl string) {
			viewSource(strUrl, nitem.wsi.charset, buf, hc, 3)
			doc, err := goquery.NewDocumentFromReader(buf)
			if err != nil {
				fmt.Println(err)
				return
			}

			strContentHtml, _ := doc.Find(nitem.nct.content).Html()
			strContent := chtReplacer.Replace(strContentHtml)
			if nitem.nct.strip != "" {
				strContent = strings.Replace(strContent, nitem.nct.strip, "", -1)
			}

			if strContent == "" {
				fmt.Println("get charpter error:", strTitle, strUrl)
			}

			f.WriteString(strTitle + "\r\n\r\n")
			f.WriteString(strContent)
			f.WriteString("\r\n")
			fmt.Println(strTitle, strUrl)
			//fmt.Printf("%8q", i)
			f.Sync()
		}(bi.chtNameList[i], bi.chtUrlList[i])
	}

	return true
}

//-----search------

func printSearchRst(nl *Noveler, strKeyWord string) {
	searchUrl := nl.wsi.searchUrl + strKeyWord
	doc, err := goquery.NewDocument(searchUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

	sea := doc.Find(nl.wsi.searchTag)
	for i := 0; i < sea.Length(); i++ {
		strTitle, ok := sea.Eq(i).Attr(nl.wsi.seaTitleTag)
		if !ok || strTitle != strKeyWord {
			continue
		}
		strHref, _ := sea.Eq(i).Attr(nl.wsi.seaUrlTag)
		fmt.Println(strTitle, strHref)
	}
}

func NovelSearch(strKeyWord string) {
	for _, v := range mapNoveler {
		if v.wsi.searchUrl != "" {
			//fmt.Println("Host:", v.wsi.wetsite)
			printSearchRst(v, strKeyWord)
		}
	}
}
