package main

import (
	"errors"
	"fmt"
	"github.com/mmcdole/gofeed"
	"strings"
	"time"
)

const (
	updated = "updated"
	DB_NAME = "telebot.db"
)

var db *BoltConnection

func parseFeed(url, chatID string, html bool, itemFunc func(int, *gofeed.Item) string) (string, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	//fmt.Printf("%#v", feed)

	if !NotifyErr(err, chatID) {
		var itemTextArr []string
		// title
		itemTextArr = append(itemTextArr, feed.Title)
		itemTextArr = append(itemTextArr, feed.Updated)

		// check sent
		// prepare db bucket
		err := db.CreateBucketIfNotExists(url)
		printErr(err)

		updatedValue, err := db.Get(url, updated)
		fatalErr(err)
		sent := feed.Updated == string(updatedValue)
		defer func() {
			err = db.Set(url, updated, feed.Updated)
			fatalErr(err)
		}()

		// items
		for i, item := range feed.Items {
			itemText := itemFunc(i, item)
			itemTextArr = append(itemTextArr, itemText)
		}

		// join them
		content := strings.Join(itemTextArr, "\n\n")
		//fmt.Println(content)

		if sent {
			return content, errors.New(fmt.Sprintf("crawled %v, but don't have any update", url))
		} else {
			return content, nil
		}
	}
	return "", errors.New(fmt.Sprintf("error during crawling %v: %v", url, err))
}

func ItemParseLink(i int, item *gofeed.Item) string {
	return fmt.Sprintf("%d %v:\n%v", i+1, item.Title, item.Link)
}

func ItemParseDesc(i int, item *gofeed.Item) string {
	return fmt.Sprintf("%d %v:\n%v", i+1, item.Title, item.Description)
}

func ClearCrawlStatus() {
	err := db.Clear()
	fatalErr(err)
	init_db()
}

func ScanRSS(url, chatID string, delta time.Duration, itemFuc func(int, *gofeed.Item) string) {
	for {
		content, err := parseFeed(url, chatID, false, itemFuc)
		if err != nil && content != "" {
			// have not updated

		} else if !NotifyErr(err, chatID) {
			NotifyText(content, chatID)
		}
		timer := time.NewTimer(delta)
		<-timer.C
	}
}

func CloseDB() {
	err := db.Close()
	printErr(err)
}

func GetOldURLs(bucket string) ([]string, error) {
	old, err := db.Get(bucket, "sources")
	if err != nil {
		return []string{}, err
	}
	return strings.Fields(string(old)), nil
}

func AddRSS(userID, url string, delta time.Duration) error {
	urls, err := GetOldURLs(userID)
	NotifyErr(err, userID)
	urls = append(urls, url)

	db.CreateBucketIfNotExists(userID)
	err = db.Set(userID, "sources", strings.Join(urls, " "))
	if err != nil {
		return err
	}

	// should send new notification to app
	go ScanRSS(url, userID, delta, ItemParseLink)
	return nil
}

func DeleteRSS(userID, url string) error {
	urls, err := GetOldURLs(userID)
	NotifyErr(err, userID)
	urls = append(urls, url)

	var newURLs []string
	for _, u := range urls {
		if u != url {
			newURLs = append(newURLs, u)
		}
	}
	err = db.Set(userID, "sources", strings.Join(urls, " "))
	return err
}

func StartRSSCrawlers() {
	for userID, user := range ChatsMap {
		if user.TeleName == adminTelegramID {
			urls, err := GetOldURLs(userID)
			if !NotifyErr(err, userID) {
				for _, url := range urls {
					go ScanRSS(url, userID, time.Minute*time.Duration(scanMinutes), ItemParseLink)
				}
			}
		}
	}
}

func init_db() {
	db = NewDB(DB_NAME)
}

func init() {
	init_db()
	StartRSSCrawlers()
}
