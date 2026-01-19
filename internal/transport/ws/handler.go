package ws

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type Socket struct {
	upgrader *websocket.Upgrader
}

func Shi() {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Restrict in production!
	}
}
