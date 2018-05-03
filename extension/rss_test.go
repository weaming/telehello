package extension

import (
	"fmt"
	"testing"
	"time"
)

const (
	WANQU = "https://wanqu.co/feed"
)

var rss = NewRSSPool(time.Duration(10000), false)

func TestRSS(t *testing.T) {
	content, err := rss.parseFeed(WANQU, "123", true, ItemParseLink)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(content)
	}
}
