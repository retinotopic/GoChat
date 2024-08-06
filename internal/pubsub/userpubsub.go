package pubsub

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/goccy/go-json"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/internal/db"
)

type UserPubSub struct {
	Db      *db.PgClient
	Conn    *websocket.Conn
	WriteCh chan db.FlowJSON
	Upubsub *redis.PubSub
	Pub     *redis.Client
}

// conn, err := upgrader.Upgrade(w, r, nil)

func (u *UserPubSub) WsHandle() {
	defer func() {
		u.Conn.Close()
	}()
	go u.ReadDB()
	errs := make(chan error, 1)
	u.Conn.SetPongHandler(func(string) error {
		u.Conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		return nil
	})
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			<-ticker.C
			err := u.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(9*time.Second))
			if err != nil {
				log.Println("server ping error:", err)
				errs <- err
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	u.Db.GetAllRooms(ctx, &db.FlowJSON{})
	go u.WsReadRedis()
	for {
		flowjson := &db.FlowJSON{}
		err := u.Conn.ReadJSON(flowjson)
		if err != nil {
			return
		}
		go u.Db.TxManage(flowjson)
	}
}
func (u *UserPubSub) WsReadRedis() {
	flowjson := db.FlowJSON{}
	chps := u.Upubsub.Channel()
	for message := range chps {
		go func(message redis.Message) {
			if err := json.Unmarshal([]byte(message.Payload), &flowjson); err != nil {
				log.Fatalln(err, "unmarshalling error")
			}
			switch flowjson.Mode {
			case "CreateGroupRoom", "CreateDuoRoom", "AddUserToRoom":
				u.Upubsub.Subscribe(context.Background(), fmt.Sprintf("%d%s", flowjson.Room, "room"))
			case "DeleteUsersFromRoom", "BlockUser":
				u.Upubsub.Unsubscribe(context.Background(), fmt.Sprintf("%d%s", flowjson.Room, "room"))
			}
			u.WriteCh <- flowjson
		}(*message)
	}
}

// stream for writing json to ws connection
func (h *UserPubSub) WsWrite() {
	for flowjson := range h.WriteCh {
		err := h.Conn.WriteJSON(&flowjson)
		if err != nil {
			return
		}
	}
}
func (u *UserPubSub) ReadDB() {
	ch := u.Db.Channel()
	for flowjson := range ch {
		u.WriteCh <- flowjson
		if flowjson.Err != nil {
			continue
		}
		payload, err := json.Marshal(&flowjson)
		if err != nil {
			log.Fatalln(err, "unmarshalling error")
		}
		switch flowjson.Mode {
		case "SendMessage":
			u.Pub.Publish(context.Background(), fmt.Sprintf("%d%s", flowjson.Room, "room"), payload)
		default:
			i := 1
			for i = range flowjson.Users {
				u.Pub.Publish(context.Background(), fmt.Sprintf("%d%s", flowjson.Users[i], "user"), payload)
			}
		}
	}
}
