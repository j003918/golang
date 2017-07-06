package main

import (
	"fmt"
	"freenovel"
)

func main() {
	freenovel.WebsiteList()

	bu := ""
	for {
		fmt.Print("Please input novel url: ")
		fmt.Scanln(&bu)
		freenovel.NovelDownload(bu)
	}

	//freenovel.NovelSearch("弄潮")
	//freenovel.NovelDownload("http://www.shoujikanshu.org/xiaoshuo/index_21793.html")
	//freenovel.NovelDownload("http://www.xxbiquge.com/65_65338/")
	//freenovel.NovelDownload("http://www.23us.com/html/67/67612/")
	//freenovel.NovelDownload("http://www.88dushu.com/xiaoshuo/71/71885/")
	//freenovel.NovelDownload("http://www.qu.la/book/4336/")
	//freenovel.NovelDownload("http://www.biqudao.com/bqge29882/")
}
