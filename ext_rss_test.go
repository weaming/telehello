package main

import (
	"fmt"
	"testing"
)

const (
	WANQU = "https://wanqu.co/feed"
)

func TestRSS(t *testing.T) {
	content, err := parseFeed(WANQU, true, ItemParseLink)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(content)
	}
}

func TestWanqu(t *testing.T) {
	content, err := parseFeed(WANQU, true, ItemParseWanquDaily)
	fmt.Println(content)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(content)
	}
}
