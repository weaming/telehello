package main

import (
	"github.com/tucnak/telebot"
	"sync"
	"time"
)

var TelegramNotificationBox = make(chan Boxer, 1000)
var ChatsMap map[string]ChatUser

type ChatUser struct {
	TeleName string
	ID       string
	sync.RWMutex
}

func (p ChatUser) Destination() string {
	return p.ID
}

func (p *ChatUser) UpdateID(new string) {
	if new != p.ID {
		p.Lock()
		defer p.Unlock()
		p.ID = new
	}
}

func notifyText(bot *telebot.Bot, content string, recipient ChatUser) (err error) {
	return bot.SendMessage(recipient, content, &telebot.SendOptions{DisableWebPagePreview: true})
}

func notifyHTML(bot *telebot.Bot, content string, recipient ChatUser) (err error) {
	return bot.SendMessage(recipient, content, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

type Boxer interface {
	Message() string
	Type() string
	Destination() string
}

func PollInbox(bot *telebot.Bot, inbox chan Boxer) {
	var err error
	for msg := range inbox {
		charID := msg.Destination()
		if msg.Type() == "HTML" {
			err = notifyHTML(bot, msg.Message(), ChatUser{ID: charID})
		} else {
			err = notifyText(bot, msg.Message(), ChatUser{ID: charID})
		}
		printErr(err)
	}
}

func NotifyText(text, chatID string) {
	TelegramNotificationBox <- &Notification{text, chatID, time.Now(), ""}
}

func NotifyHTML(text, chatID string) {
	TelegramNotificationBox <- &Notification{text, chatID, time.Now(), "HTML"}
}

// if is error, return true
func NotifyErr(err error, chatID string) bool {
	if err != nil {
		NotifyText("error: "+err.Error(), chatID)
		return true
	}
	return false
}
