package main

import (
	"fmt"
	"freenovel"
)

func main() {
	nd := freenovel.NewNovelDownloader()
	novelUrl := ""
	for {
		fmt.Print("Please input novel url: ")
		fmt.Scanln(&novelUrl)
		nd.Start(novelUrl)
	}
}
