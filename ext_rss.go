package main

import (
	"errors"
	"fmt"
	"github.com/mmcdole/gofeed"
	"strings"
	"time"
)

const (
	DBName     = "telebot.db"
	updatedKey = "updatedDate"
	rssKey     = "RSS"
	// store global meta information, such as users list, rather than user individual info
	globalKey   = "global"
	chatListKey = "chats_list"
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
		err := db.CreateBucketIfNotExists(chatID)
		printErr(err)

		rssUpdateKey := url + updatedKey
		updatedValue, err := db.Get(chatID, rssUpdateKey)
		fatalErr(err)
		sent := feed.Updated == string(updatedValue)
		defer func() {
			err = db.Set(chatID, rssUpdateKey, feed.Updated)
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

func ScanRSS(url, chatID string, delta time.Duration, itemFuc func(int, *gofeed.Item) string, daemon bool) {
	for {
		content, err := parseFeed(url, chatID, false, itemFuc)
		if err != nil && content != "" {
			// have not updatedKey

		} else if !NotifyErr(err, chatID) {
			// send rss content
			NotifyText(content, chatID)

			// log to admin
			if admin, ok := ChatsMap[AdminKey]; ok {
				if chatID != admin.Destination() {
					NotifyText(fmt.Sprintf("sent %v to %v", url, ChatsMap[chatID].String()),
						admin.Destination())
				}
			}
		}

		if daemon {
			timer := time.NewTimer(delta)
			<-timer.C
		} else {
			break
		}
	}
}

func CloseDB() {
	err := db.Close()
	printErr(err)
}

func getFieldsInDB(bucket, key string) ([]string, error) {
	db.CreateBucketIfNotExists(bucket)
	old, err := db.Get(bucket, key)
	if err != nil {
		return []string{}, err
	}
	return strings.Fields(string(old)), nil
}

func GetOldURLs(userID string) ([]string, error) {
	return getFieldsInDB(userID, rssKey)
}

func getChatIDList() []string {
	chats, _ := getFieldsInDB(globalKey, chatListKey)
	return chats
}

func AddRSS(userID, url string, delta time.Duration) error {
	urls, err := GetOldURLs(userID)
	NotifyErr(err, userID)
	urls = append(urls, url)

	err = db.Set(userID, rssKey, strings.Join(urls, " "))
	if err != nil {
		return err
	}

	// should send new notification to app
	go ScanRSS(url, userID, delta, ItemParseLink, true)
	return nil
}

func DeleteRSS(userID, url string) error {
	urls, err := GetOldURLs(userID)
	NotifyErr(err, userID)

	var newURLs []string
	for _, u := range urls {
		if u != url {
			newURLs = append(newURLs, u)
		}
	}
	err = db.Set(userID, rssKey, strings.Join(newURLs, " "))
	return err
}

func StartRSSCrawlers(daemon bool) {
	for _, chatID := range getChatIDList() {
		CrawlerForUser(chatID, daemon)
	}
}

func CrawlerForUser(userID string, daemon bool) {
	urls, err := GetOldURLs(userID)
	if !NotifyErr(err, userID) {
		for _, url := range urls {
			go ScanRSS(url, userID, time.Minute*time.Duration(scanMinutes), ItemParseLink, daemon)
		}
	}
}

func init_db() {
	db = NewDB(DBName)
	db.CreateBucketIfNotExists(globalKey)
}

func init() {
	init_db()
	StartRSSCrawlers(true)
}
