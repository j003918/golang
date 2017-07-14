// novel_http
package freenovel

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/axgle/mahonia"
)

func newNovelHttp(proxyUrl string) *http.Client {
	transport := &http.Transport{}
	if v, _ := url.Parse(proxyUrl); v.Host != "" {
		transport.Proxy = http.ProxyURL(v)
	}

	hc := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	return hc
}

func newNovelRequest(method, reqUrl, param string) *http.Request {
	req, err := http.NewRequest(method, reqUrl, strings.NewReader(param))
	if err != nil {
		fmt.Println(reqUrl, err)
		return nil
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func getBodyByReq(hc *http.Client, req *http.Request, charset string, outBuf *bytes.Buffer) bool {
	nTry := 1
RETRYGET:
	func() {
		outBuf.Reset()
		rsp, err := hc.Do(req)
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

	if outBuf.Len() == 0 && nTry < 3 {
		nTry += 1
		goto RETRYGET
	}

	return !(outBuf.Len() == 0)
}
