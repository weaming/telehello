package reminder

import (
	"database/sql"
	"github.com/jasonlvhit/gocron"
	"log"
	"strconv"
	s "strings"
	"time"
)

type ReminderContext struct {
	dbFunc       func() *sql.DB
	DB           *sql.DB
	commands     Commands
	timeLocation *time.Location
	Inbox        chan TelegramMsg
	startedCron  map[int]bool
}

type TelegramMsg struct {
	ChatID, Text string
}

type Reminder struct {
	Id      int       `sql:id`
	Content string    `sql:content`
	Created time.Time `sql:created`
	DueDt   time.Time `sql:due_dt`
	ChatId  int       `sql:chat_id`
}

func NewReminderContext(fn func() *sql.DB, cmds Commands) ReminderContext {
	sh, _ := time.LoadLocation("Asia/Shanghai")
	ctx := ReminderContext{dbFunc: fn, commands: cmds, timeLocation: sh}
	ctx.initStruct()
	return ctx
}

func (rc *ReminderContext) initStruct() {
	rc.Inbox = make(chan TelegramMsg, 100)
	rc.startedCron = map[int]bool{}
	log.Println("inited context")
}

func (rc *ReminderContext) HandleCommandText(text string, chatId int) bool {
	rc.DB = rc.dbFunc()
	defer rc.DB.Close()

	// create connection
	cmd, txt, ddt := rc.commands.Extract(text)
	log.Println(cmd, txt, ddt)

	switch s.ToLower(cmd) {
	case "remind":
		rc.save(txt, ddt, chatId)
	case "check due":
		rc.CheckDue(chatId, false)
	case "list":
		rc.list(chatId)
	case "renum":
		rc.renum(chatId)
	case "clear":
		i, _ := strconv.Atoi(txt)
		rc.clear(i, chatId)
	case "clearall":
		rc.clearall(chatId)
	default:
		return false
	}

	if _, started := rc.startedCron[chatId]; !started {
		go func() {
			rc.startedCron[chatId] = true
			gocron.Every(40).Minutes().Do(rc.CheckDue, chatId, true)
			log.Printf("Starting reminder for %v", chatId)
			<-gocron.Start()
		}()
	}

	return true
}

func (rc *ReminderContext) save(txt string, ddt time.Time, chatId int) {
	now := time.Now().Format(time.RFC3339)

	_, err := rc.DB.Exec(
		`INSERT INTO reminders(content, created, chat_id, due_dt) VALUES ($1, $2, $3, $4)`,
		txt,
		now,
		chatId,
		ddt.Format(time.RFC3339))

	panicErr(err)
	rc.SendText(chatId, "Araseo~ remember liao!")
}

func (rc *ReminderContext) clear(id int, chatId int) {
	_, err := rc.DB.Exec(`DELETE FROM reminders WHERE chat_id=$1 AND id=$2`, chatId, id)
	panicErr(err)
	// "&#127881;"
	rc.SendText(chatId, "Pew!")
}

func (rc *ReminderContext) clearall(chatId int) {
	_, err := rc.DB.Exec(`DELETE FROM reminders WHERE chat_id=$1`, chatId)
	panicErr(err)
	rc.SendText(chatId, "Pew Pew Pew!")
}

func (rc *ReminderContext) list(chatId int) {
	rows, err := rc.DB.Query(`SELECT id, content, due_dt FROM reminders WHERE chat_id=$1`, chatId)
	panicErr(err)
	defer rows.Close()

	var arr []string
	var i int
	var c string
	var dt time.Time

	for rows.Next() {
		_ = rows.Scan(&i, &c, &dt)
		line := "• " + c + " (`" + strconv.Itoa(int(i)) + "`)"
		if !dt.IsZero() {
			line = line + " - due " + dt.In(rc.timeLocation).Format("2 Jan 3:04PM")
		}
		arr = append(arr, line)
	}
	text := s.Join(arr, "\n")

	if len(text) < 5 {
		text = "No current reminders, hiak~"
	}

	rc.SendText(chatId, text)
}

func timeSinceLabel(d time.Time) string {
	var duration = time.Since(d)
	var durationNum int
	var unit string

	if int(duration.Hours()) == 0 {
		durationNum = int(duration.Minutes())
		unit = "min"
	} else if duration.Hours() < 24 {
		durationNum = int(duration.Hours())
		unit = "hour"
	} else {
		durationNum = int(duration.Hours()) / 24
		unit = "day"
	}

	if durationNum > 1 {
		unit = unit + "s"
	}

	return " `" + strconv.Itoa(int(durationNum)) + " " + unit + "`"
}

// This resets numbers for everyone!
func (rc *ReminderContext) renum(chatId int) {
	rows, err := rc.DB.Query(`SELECT content, due_dt, created, chat_id FROM reminders`)
	panicErr(err)
	defer rows.Close()

	var arr []Reminder
	var c string
	var dt time.Time
	var ct time.Time
	var cid int

	for rows.Next() {
		_ = rows.Scan(&c, &dt, &ct, &cid)
		arr = append(arr, Reminder{Content: c, DueDt: dt, Created: ct, ChatId: cid})
	}

	_, err = rc.DB.Exec(`DELETE FROM reminders`)
	panicErr(err)

	_, err = rc.DB.Exec(`DELETE FROM sqlite_sequence WHERE name='reminders';`)
	panicErr(err)

	for _, r := range arr {
		_, err := rc.DB.Exec(`INSERT INTO reminders(content, due_dt, created, chat_id) VALUES ($1, $2, $3, $4)`, r.Content, r.DueDt, r.Created, r.ChatId)
		panicErr(err)
	}

	rc.list(chatId)
}

func (rc *ReminderContext) CheckDue(chatId int, timedCheck bool) {
	rc.DB = rc.dbFunc()
	defer rc.DB.Close()

	rows, err := rc.DB.Query(
		`SELECT id, content, due_dt FROM reminders WHERE chat_id=$1 and due_dt<=$2 and due_dt!=$3`,
		chatId,
		time.Now().Format(time.RFC3339),
		"0001-01-01T00:00:00Z",
	)
	panicErr(err)
	defer rows.Close()

	var arr []string
	var i int
	var c string
	var dt time.Time

	arr = append(arr, "Overdue:")
	for rows.Next() {
		_ = rows.Scan(&i, &c, &dt)
		line := "• " + c + " (`" + strconv.Itoa(int(i)) + "`)"
		if !dt.IsZero() {
			line = line + " - due " + dt.In(rc.timeLocation).Format("2 Jan 3:04PM")
		}
		arr = append(arr, line)
	}
	text := s.Join(arr, "\n")
	log.Println(text)

	if len(text) < 10 {
		text = "No overdues, keke~"
		if !timedCheck {
			rc.SendText(chatId, text)
		}
	} else {
		rc.SendText(chatId, text)
	}
}

func (rc *ReminderContext) SendText(chatId int, text string) {
	rc.Inbox <- TelegramMsg{ChatID: strconv.Itoa(int(chatId)), Text: text}
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
