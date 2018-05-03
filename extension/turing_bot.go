package extension

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var http_client_turing = NewHTTPClient(5)

type TuringBot struct {
	Key  string
	API  string
	Name string
}

func NewTuringBot(key, name string) *TuringBot {
	return &TuringBot{
		Key:  key,
		Name: name,
		API:  "http://www.tuling123.com/openapi/api",
	}
}

func (p *TuringBot) Answer(content, userID string) string {
	var rv string

	var jsonStr = []byte(fmt.Sprintf(`{"key":"%v","info":"%v","userid"="%v"}`,
		p.Key, content, userID))
	req, err := http.NewRequest("POST", p.API, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http_client_turing.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err != nil {
		rv = "404 Not Found"
	} else {
		rv = res["text"].(string)
	}
	log.Printf("<< TuringBot bot <%v>: %v\n", resp.Status, rv)
	return rv
}
