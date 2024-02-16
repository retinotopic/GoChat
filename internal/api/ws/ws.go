package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/retinotopic/GoChat/pkg/safectx"
)

//conn, err := upgrader.Upgrade(w, r, nil)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WsConnect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		_, ok := safectx.GetContextString(r.Context(), "sub")
		if !ok {
			log.Println("no sub")
			return
		}
		_, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
	})
}
