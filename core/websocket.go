package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if PrintErr(err) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		data := map[string]interface{}{
			"ok":  false,
			"msg": fmt.Sprintf("unknow error when upgrade protocol: %s", err),
		}
		jData, err := json.Marshal(data)
		FatalErr(err)
		w.Write(jData)
		return
	}
	go ProcessMessage(conn, r)
}

func ProcessMessage(conn *websocket.Conn, req *http.Request) {
	for {
		messageType, p, err := conn.ReadMessage()
		if PrintErr(err) {
			return
		}

		var data map[string]interface{}
		switch messageType {
		case websocket.TextMessage:
			data = pushMsgQueue(req, p)
		case websocket.BinaryMessage:
			data = map[string]interface{}{
				"ok":  false,
				"msg": "binary message is not supported",
			}
		}

		// send back
		jData, err := json.Marshal(data)
		FatalErr(err)
		if err := conn.WriteMessage(websocket.TextMessage, jData); PrintErr(err) {
			return
		}
	}
}
