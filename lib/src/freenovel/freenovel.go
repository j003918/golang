// novel
package freenovel

import (
	"bytes"
	"container/list"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-yaml/yaml"
)

var chtReplacer = strings.NewReplacer("<br>", "\r\n", "<br/>", "\r\n", "<br />", "\r\n")
var mapYaml map[interface{}]interface{} = make(map[interface{}]interface{})

type website struct {
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

type chapter struct {
	idx        int
	url        string
	menuTitle  string
	chaptTitle string
	conent     string
	stats      int
}

type novelDownloader struct {
	wg       *sync.WaitGroup
	hc       *http.Client
	ws       *website
	chapters *list.List
	lock     *sync.Mutex
	url      string
	name     string
}

func (nd *novelDownloader) Start(url string) {
	wi := getWebsite(url)
	if wi == nil {
		return
	}

	nd.ws = wi
	nd.url = url
	nd.hc = newNovelHttp(wi.proxy)
	nd.requestMenu()
	nd.worker()
	nd.wg.Wait()
	nd.save2File()
	fmt.Println("saved to file", nd.name+".txt")
}

func (nd *novelDownloader) requestMenu() {
	buf := &bytes.Buffer{}
	req := newNovelRequest("GET", nd.url, "")
	getBodyByReq(nd.hc, req, nd.ws.charset, buf)

	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	nd.name = doc.Find(nd.ws.novelName).Text()
	nodes := doc.Find(nd.ws.menuList)

	itemCount := nodes.Length()
	if itemCount <= 0 {
		return
	}

	strPreUrl := ""
	strItemLink := "href"
	if strUrl, ok := nodes.Eq(0).Attr(strItemLink); ok {
		if strUrl[0] == '/' {
			uu, _ := url.Parse(nd.url)
			strPreUrl = uu.Scheme + "://" + uu.Host
		} else {
			urlIdx := strings.LastIndex(nd.url, "/")
			strPreUrl = nd.url[0 : urlIdx+1]
		}
	}

	for i := 0; i < itemCount; i++ {
		v := nodes.Eq(i)
		strTitle := v.Text()
		strUrl, _ := v.Attr(strItemLink)
		if strTitle != "" {
			cht := &chapter{
				idx:       i,
				url:       strPreUrl + strUrl,
				menuTitle: strTitle,
			}
			nd.chapters.PushBack(cht)
			fmt.Println("add chapter", strTitle, strPreUrl+strUrl)
		}
	}
}

func (nd *novelDownloader) getDownload() *chapter {
	nd.lock.Lock()
	defer nd.lock.Unlock()

	for e := nd.chapters.Front(); e != nil; e = e.Next() {
		cht := e.Value.(*chapter)
		if cht.stats == 0 {
			cht.stats = 1
			return cht
		}
	}

	return nil
}

func (nd *novelDownloader) requestChapter() {
	defer nd.wg.Done()
	buf := &bytes.Buffer{}
	for {
		cht := nd.getDownload()

		if cht != nil {

			req := newNovelRequest("GET", cht.url, "")
			req.Header.Set("Referer", nd.url)
			req.Header.Set("Connection", "keep-alive")
			req.Header.Set("DNT", "1")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36")

			getBodyByReq(nd.hc, req, nd.ws.charset, buf)
			doc, err := goquery.NewDocumentFromReader(buf)
			if err != nil {
				fmt.Println(err)
				continue
			}

			strContentHtml, _ := doc.Find(nd.ws.chtContent).Html()
			strContent := chtReplacer.Replace(strContentHtml)
			strTitle := doc.Find(nd.ws.chtTitle).Text()
			if nd.ws.chtContentStrip != "" {
				strContent = strings.Replace(strContent, nd.ws.chtContentStrip, "", -1)
			}

			if strContent == "" {
				fmt.Println("get charpter error:", strTitle, cht.url)
			}

			cht.chaptTitle = strTitle
			cht.conent = strContent
			fmt.Println("download chapter", cht.menuTitle, cht.url)
		} else {
			break
		}
	}
}

func (nd *novelDownloader) worker() {
	for i := 0; i < 10; i++ {
		nd.wg.Add(1)
		go nd.requestChapter()
	}
}

func (nd *novelDownloader) save2File() {
	f, err := os.Create(nd.name + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	var n *list.Element
	for e := nd.chapters.Front(); e != nil; e = n {
		n = e.Next()
		cht := e.Value.(*chapter)
		f.WriteString("\r\n\r\n")
		f.WriteString(cht.chaptTitle)
		f.WriteString("\r\n\r\n")
		f.WriteString(cht.conent)
		nd.chapters.Remove(e)
	}
}

func init() {
	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &mapYaml)
	if err != nil {
		panic(err)
	}
}

func NewNovelDownloader() *novelDownloader {
	return &novelDownloader{
		wg: &sync.WaitGroup{},

		chapters: list.New(),
		lock:     &sync.Mutex{},
	}
}

func getWebsite(novelUrl string) *website {
	u, err := url.Parse(novelUrl)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	ws, ok := mapYaml[u.Host].(map[interface{}]interface{})
	if !ok {
		fmt.Println("not supported website:", novelUrl)
		return nil
	}

	wi := &website{}
	wi.proxy = ws["proxy"].(string)
	wi.wetsite = ws["wetsite"].(string)
	wi.charset = ws["charset"].(string)
	wi.menuRefer = ws["menuRefer"].(string)
	wi.novelName = ws["novelName"].(string)
	wi.menuList = ws["menuList"].(string)
	wi.chtRefer = ws["chtRefer"].(string)
	wi.chtTitle = ws["chtTitle"].(string)
	wi.chtContent = ws["chtContent"].(string)
	wi.chtContentStrip = ws["chtContentStrip"].(string)

	return wi
}
