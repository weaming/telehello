package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/weaming/telehello/reminder"
	"log"
)

func startReminder(ctx reminder.ReminderContext) {
	go func() {
		pollMsg(&ctx)
		defer ctx.DB.Close()
	}()
}

func newContext() reminder.ReminderContext {
	fn := func() *sql.DB {
		db, err := sql.Open("sqlite3", sqlite3dbDir+"/reminders.db")
		fatalErr(err)
		return db
	}
	rmd := reminder.NewReminderContext(fn, reminder.NewCommandList())
	return rmd
}

func pollMsg(ctx *reminder.ReminderContext) {
	log.Println("reminder poll started")
	defer log.Println("reminder poll exited")
	for msg := range ctx.Inbox {
		NotifyText(msg.Text, msg.ChatID)
	}
}
