package extension

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/weaming/telehello/core"
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

type DeleteRSSSignal struct {
	ChatID, URL string
}
type ItemParseFunc func(int, *gofeed.Item) string

var deleteRssChan = make(chan DeleteRSSSignal, 100)

func parseFeed(url, chatID string, html bool, itemFunc ItemParseFunc) (string, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	//fmt.Printf("%#v", feed)

	if !core.NotifiedErr(err, chatID) {
		var itemTextArr []string
		// title
		itemTextArr = append(itemTextArr, feed.Title)
		itemTextArr = append(itemTextArr, feed.Updated)

		// check sent
		// prepare db bucket
		err := db.CreateBucketIfNotExists(chatID)
		core.PrintErr(err)

		rssUpdateKey := url + updatedKey
		updateTime, err := db.Get(chatID, rssUpdateKey)
		core.FatalErr(err)
		sent := feed.Updated == string(updateTime)
		defer func() {
			err = db.Set(chatID, rssUpdateKey, feed.Updated)
			core.FatalErr(err)
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
			log.Printf("crawled %v, but don't have any update\n", url)
		}
		return content, nil
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
	core.FatalErr(err)
	init_db()
}

func NotifyAdmin(text, chatID string) {
	if admin, ok := core.ChatsMap[core.AdminKey]; ok {
		if chatID != admin.Destination() {
			core.NotifyText(text, admin.Destination())
		}
	}
}

func ScanRSS(url, chatID string, delta time.Duration, itemFuc ItemParseFunc, daemon bool) {
outer:
	for {
		log.Printf("crawl rss, url:%v id:%v delta:%v daemon:%v\n", url, chatID, delta, daemon)
		content, err := parseFeed(url, chatID, false, itemFuc)
		if !core.NotifailedLog(err, chatID, "info") {
			// send rss content
			core.NotifyText(content, chatID)

			// log to admin
			if user, ok := core.ChatsMap[chatID]; ok {
				text := fmt.Sprintf("sent %v to %v", url, user.String())
				NotifyAdmin(text, chatID)
			}
		}

		if daemon {
			timer := time.NewTimer(delta)
		waitDelete:
			for {
				select {
				case pair := <-deleteRssChan:
					if pair.ChatID == chatID && pair.URL == url {
						defer func() { core.NotifyText(fmt.Sprintf("crawler for %v stopped", url), chatID) }()
						break outer
					}
					// else put signal back
					deleteRssChan <- pair
				case <-timer.C:
					// timeout, then crawl for next time
					break waitDelete
				default:
					// do nothing
				}
			}
		} else {
			// exit after crawling
			break
		}
	}
}

func CloseDB() {
	err := db.Close()
	core.PrintErr(err)
}

func GetOldURLs(userID string) ([]string, error) {
	return db.GetFieldsInDB(userID, rssKey)
}

func GetChatIDList() []string {
	chats, _ := db.GetFieldsInDB(globalKey, chatListKey)
	return chats
}

func AddUser(id string) {
	// add userID to list in DB
	_, err := db.AddFieldInDB(globalKey, chatListKey, id)
	if err != nil {
		NotifyAdmin(err.Error(), id)
	}
}

func AddRSS(userID, url string, delta time.Duration) error {
	AddUser(userID)

	urls, err := db.AddFieldInDB(userID, rssKey, url)
	core.NotifiedErr(err, userID)
	core.NotifyText(fmt.Sprintf("Current RSS list:\n%v", strings.Join(urls, "\n")), userID)

	// should send new notification to app
	go ScanRSS(url, userID, delta, ItemParseLink, true)
	return nil
}

func DeleteRSS(userID, url string) error {
	_, err := db.RemoveFieldInDB(userID, rssKey, url)
	core.NotifiedErr(err, userID)

	deleteRssChan <- DeleteRSSSignal{ChatID: userID, URL: url}
	return err
}

func StartRSSCrawlers(daemon bool, scanInterval int) {
	for _, chatID := range GetChatIDList() {
		CrawlForUser(chatID, daemon, scanInterval)
	}
}

func CrawlForUser(userID string, daemon bool, scanInterval int) {
	urls, err := GetOldURLs(userID)
	if !core.NotifiedErr(err, userID) {
		for _, url := range urls {
			go ScanRSS(url, userID, time.Minute*time.Duration(scanInterval), ItemParseLink, daemon)
		}
	}
}

func init_db() {
	db = NewDB(DBName)
	db.CreateBucketIfNotExists(globalKey)
}

func Start(interval int) {
	init_db()
	StartRSSCrawlers(true, interval)
}
