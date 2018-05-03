package core

import (
	"github.com/tucnak/telebot"
)

func Start(bot *telebot.Bot, listen string) {
	// notify from TelegramNotificationBox
	go RunInboxService(listen)
	go func() {
		PollInbox(bot, TelegramNotificationBox)
	}()
}

