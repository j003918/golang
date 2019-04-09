// testnovel project main.go
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type novelMenu struct {
	Url          string
	charset      string
	chapterTitle []string
	chapterUrl   []string
}

type novelChapter struct {
	Url            string
	menuTitle      string
	chapterTitle   string
	chapterContent string
}

var mynovel novelMenu

func getMenu(url, tag, charset string) {
	mynovel.Url = url
	mynovel.charset = charset //"utf-8"

	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s\n", res.StatusCode, res.Status)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	doc.Find(tag).Each(func(i int, s *goquery.Selection) {
		strUrl := s.AttrOr("href", "")
		strTitle := s.Text()

		if mynovel.charset != "utf-8" {
			data, _ := ioutil.ReadAll(
				transform.NewReader(bytes.NewReader([]byte(s.Text())),
					simplifiedchinese.GBK.NewDecoder()))
			strTitle = string(data)
		}

		strUrl = res.Request.URL.Scheme + "://" + res.Request.Host + strUrl
		mynovel.chapterTitle = append(mynovel.chapterTitle, strTitle)
		mynovel.chapterUrl = append(mynovel.chapterUrl, strUrl)
	})

	fmt.Println(mynovel.chapterTitle)
	fmt.Println(mynovel.chapterUrl)
}

func ExampleScrape() {
	// Request the HTML page.
	//res, err := http.Get("http://metalsucks.net")
	res, err := http.Get("https://www.xxbiquge.com/9_9208/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("#list dl dd a").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		//s.Text()

		dd := s.AttrOr("href", "")

		data, _ := ioutil.ReadAll(
			transform.NewReader(bytes.NewReader([]byte(s.Text())),
				simplifiedchinese.GBK.NewDecoder()))
		fmt.Println(dd, s.Text(), string(data))
	})

	//fmt.Println(doc.Find("head meta").Text())
	doc.Find("head meta").Each(func(i int, s *goquery.Selection) {

		ddd, err := s.Html()
		fmt.Println(ddd, err)
	})

}

func main() {
	//ExampleScrape()
	getMenu("https://www.xxbiquge.com/9_9208/", "#list dl dd a", "utf-8")
	//getItem1("http://127.0.0.1:8080/home/s2.html", "body tr", "utf-8")
}
