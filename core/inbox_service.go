package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tucnak/telebot"
)

const (
	GET  = "GET"
	POST = "POST"
)

type Notification struct {
	Content     string
	CharID      string
	ReceiveTime time.Time
	ContentType string
	PhotoBin    []byte
	photoPath   string
}

func (p *Notification) Message() string {
	return fmt.Sprintf("%v\n\n%v", prettyTime(p.ReceiveTime), p.Content)
}

func TempFile(content []byte) (tfname string, err error) {
	tmpfile, err := ioutil.TempFile("", "telehello.*.png")
	tfname = tmpfile.Name()
	if err != nil {
		return
	}
	if _, err = tmpfile.Write(content); err != nil {
		tmpfile.Close()
		return
	}
	if err = tmpfile.Close(); err != nil {
		return
	}
	return
}

func (p *Notification) Photo() *telebot.Photo {
	if len(p.PhotoBin) == 0 {
		return nil
	}
	tempfile, err := TempFile(p.PhotoBin)
	if err != nil {
		log.Println(err)
		return nil
	}
	file, err := telebot.NewFile(tempfile)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &telebot.Photo{File: file}
}

func (p *Notification) PhotoClean() {
	exist, err := ExistFile(p.photoPath)
	if err != nil {
		log.Println(err)
	}
	if exist {
		defer func() {
			if r := recover(); r == nil {
				log.Println(r)
			}
		}()
		os.Remove(p.photoPath)
	}
}

func (p *Notification) Type() string {
	return p.ContentType
}

func (p *Notification) Destination() string {
	return p.CharID
}

func pushMsgQueue(req *http.Request, body []byte) map[string]interface{} {
	var data map[string]interface{}
	if admin, ok := ChatsMap[AdminKey]; ok {
		NotifyHTML(fmt.Sprintf("%s\nMessage IP: %s\n", string(body), GetMessageIP(req)), admin.ID)
		data = map[string]interface{}{
			"ok": true,
		}
	} else {
		data = map[string]interface{}{
			"ok":  false,
			"msg": "amdin id is not in chats map, send a message to bot to add it",
		}
	}
	return data
}

func pushImageQueue(req *http.Request, body []byte) map[string]interface{} {
	var data map[string]interface{}
	if admin, ok := ChatsMap[AdminKey]; ok {
		NotifyPhoto(fmt.Sprintf("%s\nMessage IP: %s\n", "New Image", GetMessageIP(req)), admin.ID, body)
		data = map[string]interface{}{
			"ok": true,
		}
	} else {
		data = map[string]interface{}{
			"ok":  false,
			"msg": "amdin id is not in chats map, send a message to bot to add it",
		}
	}
	return data
}

func sendSlack(msg string) error {
	data := map[string]interface{}{
		"text": msg,
	}
	hook := os.Getenv("SLACK_HOOK")
	if hook != "" {
		_, err := PostJson(hook, data)
		return err
	}
	return errors.New("missing env SLACK_HOOK, e.g. https://hooks.slack.com/services/xxx/xxx/xxxx")
}

func PostJson(api string, data map[string]interface{}) (map[string]interface{}, error) {
	bytesRepresentation, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(api, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func SendToSlackBot(req *http.Request, body []byte) map[string]interface{} {
	var data map[string]interface{}
	msg := fmt.Sprintf("%s\nMessage IP: %s\n", string(body), GetMessageIP(req))
	err := sendSlack(msg)
	if err != nil {
		data = map[string]interface{}{
			"ok":  false,
			"msg": err.Error(),
		}
	} else {
		data = map[string]interface{}{
			"ok": true,
		}
	}
	return data
}

func NewMessageHandler(w http.ResponseWriter, req *http.Request) {
	// json type
	w.Header().Set("Content-Type", "application/json")

	// check method
	var data map[string]interface{}
	if req.Method == POST {
		// success
		defer req.Body.Close()
		body, _ := ioutil.ReadAll(req.Body)

		// push into TelegramNotificationBox
		data = pushMsgQueue(req, body)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		data = map[string]interface{}{
			"ok":  false,
			"msg": "method not allowed",
		}
	}

	jData, err := json.Marshal(data)
	FatalErr(err)
	w.Write(jData)
}

func NewImageHandler(w http.ResponseWriter, req *http.Request) {
	// json type
	w.Header().Set("Content-Type", "application/json")

	// check method
	var data map[string]interface{}
	if req.Method == POST {
		// success
		defer req.Body.Close()
		body, _ := ioutil.ReadAll(req.Body)

		// push into TelegramNotificationBox
		data = pushImageQueue(req, body)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		data = map[string]interface{}{
			"ok":  false,
			"msg": "method not allowed",
		}
	}

	jData, err := json.Marshal(data)
	FatalErr(err)
	w.Write(jData)
}

func SlackBotHandler(w http.ResponseWriter, req *http.Request) {
	// json type
	w.Header().Set("Content-Type", "application/json")

	// check method
	var data map[string]interface{}
	if req.Method == POST {
		// success
		defer req.Body.Close()
		body, _ := ioutil.ReadAll(req.Body)

		// push into TelegramNotificationBox
		data = SendToSlackBot(req, body)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		data = map[string]interface{}{
			"ok":  false,
			"msg": "method not allowed",
		}
	}

	jData, err := json.Marshal(data)
	FatalErr(err)
	w.Write(jData)
}

func RunInboxService(listen string) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`POST text to <a href="/api/new">/api/new</a> to send me a notification.`))
	})

	http.HandleFunc("/status/users", func(w http.ResponseWriter, req *http.Request) {
		js, err := json.Marshal(ChatsMap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	// reserved for compatible purpose
	http.HandleFunc("/api/new", NewMessageHandler)
	http.HandleFunc("/api/new/telegram", NewMessageHandler)
	http.HandleFunc("/api/new/telegram/image", NewImageHandler)
	http.HandleFunc("/api/new/telegram/websocket", WebsocketHandler)
	http.HandleFunc("/api/new/slack", SlackBotHandler)

	err := http.ListenAndServe(listen, nil)
	FatalErr(err)
}
