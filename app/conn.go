package app

import (
	"context"
	"encoding/json"
	"github.com/gdamore/tcell/v2"
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
func (c *Chat) TryConnect(username, url string) error {
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
			return err
		}
		msg := Message{}
		err = json.Unmarshal(b, &msg)
		if err != nil {
			return err
		}
		if msg.RoomId != 0 {
			c.NewMessageEvent(msg)
			continue
		}
		rm := []RoomServer{}
		err = json.Unmarshal(b, &rm)
		if err != nil {
			return err
		}
		if len(rm) != 0 {
			c.ProcessRoom(rm)
			continue
		}
		e := SendEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return err
		}
		if e.UserId != 0 {
			c.NewEvent(e)
			continue
		}
		switch e.Event {
		case "Get MessagesFrom Room":
			var msgs []Message
			err = json.Unmarshal(e.Data, &msgs)
			if err != nil {
				return err
			}
			c.LoadMessagesEvent(msgs)
			continue
		case "Find Users":
			c.FillUsers(e.Data, 3, c.FoundUsers)
		case "Get Blocked Users":
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
func (c *Chat) NewEvent(e SendEvent) {
	ll := c.Lists[4].Items.(*list.ArrayList)
	errstr := ""
	if len(e.ErrorMsg) != 0 {
		errstr = "Error: " + e.ErrorMsg
	}
	ni := ll.NewItem(
		[2]tcell.Color{tcell.ColorBlue, tcell.ColorRed},
		e.Event,
		errstr,
	)
	c.Lists[4].Items.MoveToFront(ni)
}
