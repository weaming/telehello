package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/weaming/telehello/extension"
	"regexp"
	"strings"
)

const (
	RSS_WANQU    = "https://wanqu.co/feed"
	RSS_HACKMIND = "http://mindhacks.cn/feed/"
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

func GoBuiltinRSS(rss *extension.RSSPool, id string) {
	go rss.ScanRSS(RSS_WANQU, id, ItemParseWanquDaily, true)
	go rss.ScanRSS(RSS_HACKMIND, id, extension.ItemParseLink, true)
}
