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
}
