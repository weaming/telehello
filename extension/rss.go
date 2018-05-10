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
	dbname     = "telebot.db"
	updatedKey = "updatedDate"
	rssKey     = "RSS"
	// store global meta information, such as users list, rather than user individual info
	globalKey   = "global"
	chatListKey = "chats_list"
	but         = "but don't have any update"
)

type RSSPool struct {
	db       *BoltConnection
	delCh    chan DeleteRSSSignal
	interval time.Duration
}

type DeleteRSSSignal struct {
	ChatID, URL string
}

type ItemParseFunc func(int, *gofeed.Item) string

func NewRSSPool(interval time.Duration, resetdb bool) *RSSPool {
	p := RSSPool{
		db:       NewDB(dbname),
		delCh:    make(chan DeleteRSSSignal, 100),
		interval: interval,
	}
	if resetdb {
		p.ClearCrawlStatus()
	}
	p.db.CreateBucketIfNotExists(globalKey)
	return &p
}

func (p *RSSPool) parseFeed(url, chatID string, html bool, itemFunc ItemParseFunc) (string, error) {
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
		err := p.db.CreateBucketIfNotExists(chatID)
		core.PrintErr(err)

		rssUpdateKey := url + updatedKey
		updateTime, err := p.db.Get(chatID, rssUpdateKey)
		core.FatalErr(err)
		sent := feed.Updated == string(updateTime)
		defer func() {
			err = p.db.Set(chatID, rssUpdateKey, feed.Updated)
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
			msg := fmt.Sprintf("crawled %v, %v\n", url, but)
			log.Println(msg)
			return content, errors.New(msg)
		}
		return content, nil
	}
	return "", errors.New(fmt.Sprintf("error during crawling %v: %v", url, err))
}

func (p *RSSPool) ClearCrawlStatus() {
	err := p.db.Clear()
	core.FatalErr(err)
}

func ItemParseLink(i int, item *gofeed.Item) string {
	return fmt.Sprintf("%d %v:\n%v", i+1, item.Title, item.Link)
}

func ItemParseDesc(i int, item *gofeed.Item) string {
	return fmt.Sprintf("%d %v:\n%v", i+1, item.Title, item.Description)
}

func (p *RSSPool) ScanRSS(url, chatID string, itemFuc ItemParseFunc, daemon bool) {
outer:
	for {
		log.Printf("crawl rss, url:%v id:%v delta:%v daemon:%v\n", url, chatID, p.interval, daemon)
		content, err := p.parseFeed(url, chatID, false, itemFuc)
		if err != nil {
			if !strings.Contains(err.Error(), but) {
				core.NotifiedLog(err, chatID, "info")
			}
		} else {
			// send rss content
			core.NotifyText(content, chatID)

			// log to admin
			if user, ok := core.ChatsMap[chatID]; ok {
				text := fmt.Sprintf("sent %v to %v", url, user.String())
				core.NotifyAdmin(text, chatID)
			}
		}

		if daemon {
			timer := time.NewTimer(p.interval)
		waitDelete:
			for {
				select {
				case pair := <-p.delCh:
					if pair.ChatID == chatID && pair.URL == url {
						defer func() { core.NotifyText(fmt.Sprintf("crawler for %v stopped", url), chatID) }()
						break outer
					}
					// else put signal back
					p.delCh <- pair
				case <-timer.C:
					// timeout, then crawl for next time
					break waitDelete
				}
			}
		} else {
			// exit after crawling
			break
		}
	}
}

func (p *RSSPool) CloseDB() {
	err := p.db.Close()
	core.PrintErr(err)
}

func (p *RSSPool) GetOldURLs(userID string) ([]string, error) {
	return p.db.GetFieldsInDB(userID, rssKey)
}

func (p *RSSPool) GetChatIDList() []string {
	chats, _ := p.db.GetFieldsInDB(globalKey, chatListKey)
	return chats
}

func (p *RSSPool) AddUser(id string) {
	// add userID to list in DB
	_, err := p.db.AddFieldInDB(globalKey, chatListKey, id)
	if err != nil {
		core.NotifyAdmin(err.Error(), id)
	}
}

func (p *RSSPool) AddRSS(userID, url string) error {
	p.AddUser(userID)

	urls, err := p.db.AddFieldInDB(userID, rssKey, url)
	core.NotifiedErr(err, userID)
	core.NotifyText(fmt.Sprintf("Current RSS list:\n%v", strings.Join(urls, "\n")), userID)

	// should send new notification to app
	go p.ScanRSS(url, userID, ItemParseLink, true)
	return nil
}

func (p *RSSPool) DeleteRSS(userID, url string) error {
	_, err := p.db.RemoveFieldInDB(userID, rssKey, url)
	core.NotifiedErr(err, userID)

	p.delCh <- DeleteRSSSignal{ChatID: userID, URL: url}
	return err
}

func (p *RSSPool) LoopOnExistedUsers(daemon bool) {
	for _, chatID := range p.GetChatIDList() {
		p.CrawlForUser(chatID, daemon)
	}
}

func (p *RSSPool) CrawlForUser(userID string, daemon bool) {
	urls, err := p.GetOldURLs(userID)
	if !core.NotifiedErr(err, userID) {
		for _, url := range urls {
			go p.ScanRSS(url, userID, ItemParseLink, daemon)
		}
	}
}

func (p *RSSPool) Start() {
	p.LoopOnExistedUsers(true)
}
