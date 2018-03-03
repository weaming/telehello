package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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
}

func (p *Notification) Message() string {
	return fmt.Sprintf("%v\n\n%v", prettyTime(p.ReceiveTime), p.Content)
}

func (p *Notification) Type() string {
	return p.ContentType
}

func (p *Notification) Destination() string {
	return p.CharID
}

func RunInboxService(listen string) {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`POST content to <a href="/api/new">/api/new</a> to send notification to me.`))
	})

	http.HandleFunc("/api/new", func(w http.ResponseWriter, req *http.Request) {
		// json type
		w.Header().Set("Content-Type", "application/json")

		// check method
		var data map[string]interface{}
		if req.Method == POST {
			// success
			defer req.Body.Close()
			body, _ := ioutil.ReadAll(req.Body)

			// push into TelegramNotificationBox
			if user, ok := ChatsMap["root"]; ok {
				NotifyText(fmt.Sprintf("%s\nMessage IP: %s\n", string(body),
					strings.Split(req.RemoteAddr, ":")[0]), user.ID)
				data = map[string]interface{}{
					"ok": true,
				}
			} else {
				data = map[string]interface{}{
					"ok": false,
				}
			}
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			data = map[string]interface{}{
				"ok":  false,
				"msg": "method not allowed",
			}
		}

		jData, err := json.Marshal(data)
		printErr(err)
		w.Write(jData)
	})

	err := http.ListenAndServe(listen, nil)
	fatalErr(err)
}
