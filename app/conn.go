package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/retinotopic/GoChat/app/list"
)

func WriteTimeout(timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}
func (c *Chat) TryConnect(username, url string) {
	hd := http.Header{}
	cookie := http.Cookie{
		Name:     "username",
		Value:    username,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	if v := cookie.String(); v != "" {
		hd.Add("Cookie", v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	opts := &websocket.DialOptions{HTTPHeader: hd}
	conn, _, err := websocket.Dial(ctx, url, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.CloseNow()
	for {
		_, b, err := conn.Read(context.TODO())
		if err != nil {
			return
		}
		msg := Message{}
		err = json.Unmarshal(b, &msg)
		if err != nil {
			return
		}
		if msg.RoomId != 0 {
			c.NewMessageEvent(msg)
			continue
		}
		rm := RoomServer{}
		err = json.Unmarshal(b, &rm)
		if err != nil {
			return
		}
		if rm.RoomId != 0 {
			c.ProcessRoom([]RoomServer{rm})
			continue
		}
		e := Event{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return
		}
		switch e.Event {
		case "GetMessagesFromRoom":
			var msgs []Message
			err = json.Unmarshal(e.Data, &msgs)
			if err != nil {
				return
			}
			c.LoadMessagesEvent(msgs)
			continue
		case "FindUser":
			c.FillUsers(e.Data, 3, c.FoundUsers)
		case "GetBlockedUsers":
			c.FillUsers(e.Data, 1, c.BlockedUsers)
		}

	}

}
func (c *Chat) FillUsers(data []byte, idx int, m map[uint64]User) {
	c.App.QueueUpdateDraw(func() {
		var usrs []User
		err := json.Unmarshal(data, &usrs)
		if err != nil {
			return
		}
		c.Lists[idx].Items.Clear()
		for _, v := range usrs {
			m[v.UserId] = v
			c.Lists[idx].Items.MoveToFront(list.ArrayItem{MainText: v.Username,
				SecondaryText: strconv.FormatUint(v.UserId, 10)})
		}
	})

}
