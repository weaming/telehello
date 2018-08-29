package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tucnak/telebot"
	"github.com/weaming/telehello/core"
	"github.com/weaming/telehello/extension"
)

const (
	TURING_KEY  = "e8151bef6a9f9deaf641a7c71b5cb0bc"
	TURING_NAME = "小Q"
)

var turing *extension.TuringBot
var listen string
var period int64 = 30
var resetdb bool
var scanMinutes int = 60 * 24
var doubanScore float64
var adminTelegramID = "weaming"

var rss *extension.RSSPool

func init() {
	fmt.Printf("一个Telegram消息机器人\n\nFeatures:\n\t1. RSS抓取\n\t2. HTTP接口接收消息\n\t3. 图灵聊天机器人\n\n")
	// parse args
	flag.StringVar(&adminTelegramID, "telegramID", adminTelegramID, "your telegram ID without @")
	flag.StringVar(&listen, "l", ":1234", "[host]:port http hook to receive message")
	flag.Int64Var(&period, "t", 30, "telegram bot long poll timeout in seconds")
	flag.BoolVar(&resetdb, "x", false, "delete bot status KV database before start")
	flag.IntVar(&scanMinutes, "rss", 60*24, "period of crawling wanqu.co RSS in minutes")
	flag.Float64Var(&doubanScore, "douban", 8, "douban movie min score")
	flag.Parse()

	interval := time.Minute * time.Duration(scanMinutes)

	// turing robot
	turing = extension.NewTuringBot(TURING_KEY, TURING_NAME)
	// RSS
	rss = extension.NewRSSPool(interval, resetdb)
	rss.Start()
	// douban host movie
	go ScanDoubanMovie(doubanScore, time.Duration(60*24))
}

func main() {
	// telegram robot
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		token = "337645430:AAFQcjIk1bBffl5x1O1T-A9ZvAliCOreTCo"
	}
	bot, err := telebot.NewBot(token)
	core.FatalErr(err)
	log.Printf("running with token: %v\n", token)

	messages := make(chan telebot.Message, 100)
	bot.Listen(messages, time.Duration(period)*time.Second)
	core.Start(bot, listen)

	// handler received msg from app
	go func() {
		for message := range messages {
			// run in separate goroutine
			go func(message telebot.Message) {
				core.LogMessage(message)

				// prepare common
				userID := strconv.Itoa(message.Origin().ID)
				userName := message.Origin().Username
				text := message.Text

				// check if new user first
				if _, ok := core.ChatsMap[userID]; !ok {
					rss.AddUser(userID)

					if root, ok2 := core.ChatsMap[core.AdminKey]; ok2 {
						// send log to admin
						core.NotifyText(fmt.Sprintf("New user %v(%v)", userName, userID), root.ID)
					}
					if userName == adminTelegramID {
						// crawl defaults RSSes for weaming
						GoBuiltinRSS(rss, userID)
						// update chat id with myself
						core.ChatsMap[core.AdminKey] = &core.ChatUser{TeleName: userName, ID: userID}
					}
				}

				// register/update user
				core.ChatsMap[userID] = &core.ChatUser{TeleName: userName, ID: userID}

				// process text
				var responseText string

				if len(text) == 0 {
					goto photos
				}

				if text[0] == '/' {
					responseText = extension.ProcessCommand(text, userID, rss, turing)
				} else {
					if message.Text == "hi" {
						responseText = "Hello, " + message.Sender.FirstName + "!"
					} else if strings.HasPrefix(text, "debug") {
						responseText = text
					} else {
						responseText = turing.Answer(text, userID)
					}
				}

			photos:
				if len(message.Photo) > 0 {
					var logs []string
					// only the largest file
					thumbnail, err := core.MaxFileSize(message.Photo)

					// for loop on photos
					if err != nil {
						responseText = err.Error()
					} else {
						url, _ := bot.GetFileDirectURL(thumbnail.FileID)
						filePath, err := core.DownloadTeleFile(url, thumbnail.FileID+".jpg")
						if err != nil {
							log.Println(err)
							logs = append(logs, err.Error())
						} else {
							logText := fmt.Sprintf("downloaded at %v", filePath)
							log.Println(logText)
							logs = append(logs, logText)
						}
						responseText = strings.Join(logs, "\n\n")
					}
				}

				// response text
				bot.SendMessage(message.Chat, responseText, nil)
			}(message)
		}
	}()

	fmt.Println("Started")
	exit := make(chan bool)
	<-exit
}