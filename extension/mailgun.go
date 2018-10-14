package extension

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mailgun/mailgun-go"
)

func init() {
	http.HandleFunc("/api/new/email", NewEmailHandler)
}

func SendMail(mg mailgun.Mailgun, sender, subject, body, recipient string) (string, string, error) {
	message := mg.NewMessage(sender, subject, body, recipient)
	return mg.Send(message)
}

func NewEmailHandler(w http.ResponseWriter, req *http.Request) {
	// json type
	w.Header().Set("Content-Type", "application/json")

	// check method
	var data map[string]interface{}
	if req.Method == POST {
		// success
		defer req.Body.Close()
		body, _ := ioutil.ReadAll(req.Body)

		// send mail via mailgun
		data = sendMail(req, body)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		data = map[string]interface{}{
			"ok":  false,
			"msg": "method not allowed",
		}
	}

	jData, err := json.Marshal(data)
	PrintErr(err)
	w.Write(jData)
}

func sendMail(req *http.Request, body []byte) map[string]interface{} {
	var data map[string]interface{}

	// get email subject
	m, _ := url.ParseQuery(req.URL.RawQuery)
	subject := m.Get("subject")
	if subject == "" {
		subject = "No Subject"
	}
	title := m.Get("title")
	sender := "noreply@drink.cafe"
	recipient := "garden.yuen@gmail.com"

	bodyStr := fmt.Sprintf("%s\n\nMessage IP: %s\n", string(body), strings.Split(req.RemoteAddr, ":")[0])
	if title != "" {
		bodyStr = fmt.Sprintf("%v\n\n%v", title, bodyStr)
	}

	var MG_DOMAIN = os.Getenv("MG_DOMAIN")
	var MG_API_KEY = os.Getenv("MG_API_KEY")
	if MG_DOMAIN == "" || MG_API_KEY == "" {
		data = map[string]interface{}{
			"ok":     false,
			"reason": "setup your MG_DOMAIN and MG_API_KEY",
		}
		return data
	}
	var mg = mailgun.NewMailgun(MG_DOMAIN, MG_API_KEY)

	response, id, err := SendMail(mg, sender, subject, bodyStr, recipient)
	if err != nil {
		data = map[string]interface{}{
			"ok": false,
		}
	} else {
		data = map[string]interface{}{
			"ok":       true,
			"response": response,
			"id":       id,
		}
	}
	return data
}

func PrintErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
