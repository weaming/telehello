package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/tucnak/telebot"
)

var TelegramNotificationBox = make(chan InboxMessage, 1000)
var ChatsMap = make(map[string]*ChatUser)
var AdminKey = "admin"

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

func (p *ChatUser) String() string {
	return fmt.Sprintf("%v(%v)", p.TeleName, p.ID)
}

type Notifier func(bot *telebot.Bot, content string, recipient ChatUser) (err error)

func notifyText(bot *telebot.Bot, content string, recipient ChatUser) (err error) {
	return bot.SendMessage(recipient, content, &telebot.SendOptions{DisableWebPagePreview: true})
}

func notifyHTML(bot *telebot.Bot, content string, recipient ChatUser) (err error) {
	return bot.SendMessage(recipient, content, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}
func notifyPhoto(bot *telebot.Bot, photo *telebot.Photo, recipient ChatUser) (err error) {
	return bot.SendPhoto(recipient, photo, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

var notifyFuncMap = map[string]Notifier{
	"default": notifyText,
	"HTML":    notifyHTML,
}

type InboxMessage interface {
	Message() string
	Photo() *telebot.Photo
	PhotoClean()
	Type() string
	Destination() string
}

func PollInbox(bot *telebot.Bot, inbox chan InboxMessage) {
	var err error
	for msg := range inbox {
		charID := msg.Destination()
		if msg.Type() == "PHOTO" {
			photo := msg.Photo()
			if photo != nil {
				err = notifyPhoto(bot, photo, ChatUser{ID: charID})
			}
			defer msg.PhotoClean()
		} else {
			if fn, exist := notifyFuncMap[msg.Type()]; exist {
				err = fn(bot, msg.Message(), ChatUser{ID: charID})
			} else {
				fn := notifyFuncMap["default"]
				err = fn(bot, msg.Message(), ChatUser{ID: charID})
			}
		}
		PrintErr(err)
	}
}

func NotifyText(text, chatID string) {
	TelegramNotificationBox <- &Notification{text, chatID, time.Now(), "", []byte{}, ""}
}
func NotifyHTML(text, chatID string) {
	TelegramNotificationBox <- &Notification{text, chatID, time.Now(), "HTML", []byte{}, ""}
}
func NotifyPhoto(text, chatID string, bin []byte) {
	TelegramNotificationBox <- &Notification{text, chatID, time.Now(), "PHOTO", bin, ""}
}

func NotifiedLog(err error, chatID, level string) bool {
	if err != nil {
		NotifyText(fmt.Sprintf("%v: %v", level, err.Error()), chatID)
		// if is error, return true
		return true
	}
	return false
}
func NotifiedErr(err error, chatID string) bool {
	return NotifiedLog(err, chatID, "error")
}

func NotifyAdmin(text, chatID string) {
	if admin, ok := ChatsMap[AdminKey]; ok {
		if chatID != admin.Destination() {
			NotifyText(text, admin.Destination())
		}
	}
}
