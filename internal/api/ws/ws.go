package ws

import "github.com/gorilla/websocket"

//conn, err := upgrader.Upgrade(w, r, nil)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
