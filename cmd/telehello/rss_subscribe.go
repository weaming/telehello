package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/weaming/telehello/extension"
)

const (
	RSS_WANQU      = "https://wanqu.co/feed"
	RSS_HACKMIND   = "http://mindhacks.cn/feed/"
	RSS_RUANYIFENG = "http://www.ruanyifeng.com/blog/atom.xml"
)

func ItemParseWanquDaily(i int, item *gofeed.Item) string {
	htmlText := strings.TrimSpace(item.Description)
	var regURL = regexp.MustCompile(`^<img .+><a .*href="(http.+?)\?.+">`)

	result := regURL.FindAllStringSubmatch(htmlText, 1)
	var url string
	if len(result) > 0 && len(result[0]) > 1 {
		url = result[0][1]
	}

	if url == "" {
		var regURL1 = regexp.MustCompile(`^<img .+>.*<a .*href="(http.+)">`)
		result := regURL1.FindAllStringSubmatch(htmlText, 1)
		if len(result) > 0 && len(result[0]) > 1 {
			url = result[0][1]
		}
	}
	return fmt.Sprintf("%d %v:\n%v", i+1, item.Title, url)
}

func GoBuiltinRSS(id string) {
	go extension.ScanRSS(RSS_WANQU, id, time.Minute*time.Duration(scanMinutes), ItemParseWanquDaily, true)
	go extension.ScanRSS(RSS_HACKMIND, id, time.Minute*time.Duration(scanMinutes), extension.ItemParseLink, true)
	//go ScanRSS(RSS_RUANYIFENG, time.Minute*time.Duration(scanMinutes), ItemParseLink)
}
