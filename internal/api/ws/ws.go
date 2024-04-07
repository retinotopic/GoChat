package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/internal/db"
	"github.com/retinotopic/GoChat/pkg/safectx"
)

type HandlerWS struct {
	Rc   *redis.Client
	DBc  *db.PostgresClient
	Conn *websocket.Conn
}

// conn, err := upgrader.Upgrade(w, r, nil)
func NewHandlerWS(dbc *db.PostgresClient, conn *websocket.Conn) *HandlerWS {
	return &HandlerWS{
		DBc:  dbc,
		Conn: conn,
		Rc: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WsConnect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		sub, ok := safectx.GetContextString(r.Context(), "sub")
		if !ok {
			log.Println("no sub")
			return
		}
		dbconn, err := db.ConnectToDB(os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Println("wrong sub", err)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		dbc, err := db.NewClient(sub, dbconn)
		if err != nil {
			//write error to plain http
			w.WriteHeader(http.StatusInternalServerError)
		}
		wsc := NewHandlerWS(dbc, conn)
		wsc.WsHandle()
	})
}
func (h HandlerWS) WsHandle() error {
	defer func() {
		h.DBc.Conn.Release()
		h.Conn.Close()
	}()
	FuncMap := make(map[string]func(*db.FlowJSON))
	FuncMap["SendMessage"] = h.DBc.SendMessage
	FuncMap["GetTopMessages"] = h.DBc.GetAllRooms
	FuncMap["CreateDuoRoom"] = h.DBc.CreateDuoRoom
	FuncMap["CreateGroupRoom"] = h.DBc.CreateRoom
	flowjson1 := &db.FlowJSON{}
	h.DBc.GetAllRooms(flowjson1)
	flowjson2 := &db.FlowJSON{}
	h.DBc.GetMessagesFromNextRooms(flowjson2)

	if flowjson1.Err != nil || flowjson2.Err != nil {
		log.Println(flowjson1.Err)
		return errors.New("cant get info")
	}

	go h.WsReceive(FuncMap)
	go h.WsSend(FuncMap)
	return nil
}
func (h HandlerWS) WsReceive(funcMap map[string]func(*db.FlowJSON)) {
	rps := h.Rc.Subscribe(context.Background(), "chat")
	flowjson := &db.FlowJSON{}

	for {
		message, err := rps.ReceiveMessage(context.Background())
		if err != nil {
			log.Println(err)
		}
		if err := json.Unmarshal([]byte(message.Payload), &flowjson); err != nil {
			log.Fatalln(err, "unmarshalling error")
		}
		if err != nil {
			log.Println(err)
		}
		err = h.Conn.WriteJSON(flowjson)
		if err != nil {
			log.Println(err)
			break
		}
	}
}
func (h HandlerWS) WsSend(funcMap map[string]func(*db.FlowJSON)) {
	for {
		flowjson := &db.FlowJSON{}
		err := h.Conn.ReadJSON(flowjson)
		if err != nil {
			log.Println(err)
			break
		}
		h.DBc.TxBegin(flowjson)
		if flowjson.Err != nil {
			funcMap[flowjson.Mode](flowjson)
		}
		h.DBc.TxCommit(flowjson)
		if flowjson.Err != nil {
			return
		}

		if flowjson.Status == "bad" || flowjson.Status == "senderonly" {
			h.Conn.WriteJSON(flowjson)
		} else {
			payload, err := json.Marshal(flowjson)
			if err != nil {
				log.Fatalln(err, "marshalling error")
			}
			if err := h.Rc.Publish(context.Background(), fmt.Sprintf("%d", flowjson.Room), payload).Err(); err != nil {
				log.Println(err)
			}
		}
	}
}
