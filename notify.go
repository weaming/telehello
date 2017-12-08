package main

import (
	"github.com/tucnak/telebot"
	"sync"
	"time"
)

var myChatID = "142664361"
var TelegramNotificationBox = make(chan Boxer, 1000)
var MeRecipient Receiver = &Myself{ID: myChatID}

type Receiver interface {
	telebot.Recipient
	UpdateID(string)
}

type Myself struct {
	ID string
	L  sync.RWMutex
}

func (p *Myself) Destination() string {
	p.L.RLock()
	defer p.L.RUnlock()
	return p.ID
}

func (p *Myself) UpdateID(new string) {
	if new != myChatID {
		p.L.Lock()
		defer p.L.Unlock()
		p.ID = new
	}
}

func notifyMeText(bot *telebot.Bot, content string) (err error) {
	return bot.SendMessage(MeRecipient, content, &telebot.SendOptions{DisableWebPagePreview: true})
}

func notifyMeHTML(bot *telebot.Bot, content string) (err error) {
	return bot.SendMessage(MeRecipient, content, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

type Boxer interface {
	Message() string
	Type() string
}

func PollInbox(bot *telebot.Bot, inbox chan Boxer) {
	var err error
	for msg := range inbox {
		if msg.Type() == "HTML" {
			err = notifyMeHTML(bot, msg.Message())
		} else {
			err = notifyMeText(bot, msg.Message())
		}
		printErr(err)
	}
}

func NotifyText(s string) {
	TelegramNotificationBox <- &Notification{s, time.Now(), ""}
}

func NotifyHTML(s string) {
	TelegramNotificationBox <- &Notification{s, time.Now(), "HTML"}
}

// if is error, return true
func NotifyErr(err error) bool {
	if err != nil {
		NotifyText("error: " + err.Error())
		return true
	}
	return false
}
