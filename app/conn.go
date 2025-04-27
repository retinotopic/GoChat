package app

import (
	"context"
	json "github.com/bytedance/sonic"
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

type IncomingEvent struct {
	Event    string `json:"Event"`
	ErrorMsg string `json:"ErrorMsg"`
	UserId   uint64 `json:"UserId"`
	Data     []byte `json:"Data"`
}

func (c *Chat) TryConnect(username, url string) <-chan error {
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
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	opts := &websocket.DialOptions{HTTPHeader: hd}
	c.Conn, _, err = websocket.Dial(ctx, url, opts)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			msgType, b, err := c.Conn.Read(context.TODO())
			if err != nil {
				c.errch <- err
				c.Conn.CloseNow()
			}
			if msgType == websocket.MessageText && len(b) > 0 {

				c.Logger.Println(b, "CONN READ")
				msg := Message{}
				err = json.Unmarshal(b, &msg)
				if err == nil && msg.RoomId != 0 {
					c.NewMessageEvent(msg)
					continue
				}
				rm := []RoomServer{}
				err = json.Unmarshal(b, &rm)
				if err == nil && len(rm) != 0 {
					c.ProcessRoom(rm)
					continue
				}
				e := IncomingEvent{}
				err = json.Unmarshal(b, &e)
				if err == nil && e.UserId != 0 {
					c.Logger.Println(e, "conn SendEvent")
					if c.NewEventNotification(e) {
						continue
					}
					switch e.Event {
					case "Get Messages From Room":
						var msgs []Message
						err = json.Unmarshal(e.Data, &msgs)
						if err != nil {
							continue
						}
						c.LoadMessagesEvent(msgs)
					case "Find Users":
						c.Logger.Println(e, "Find Users")
						c.FillUsers(e.Data, 3, c.FoundUsers)
					case "Get Blocked Users":
						c.Logger.Println(e, "Get Blocked Users")
						c.FillUsers(e.Data, 1, c.BlockedUsers)
					}
				}
			}
		}
	}()

	return c.errch

}
func (c *Chat) FillUsers(data []byte, idx int, m map[uint64]User) {
	c.App.QueueUpdateDraw(func() {
		log.Fatalln("fill users")
		var usrs []User
		err := json.Unmarshal(data, &usrs)
		if err != nil {
			return
		}
		c.Lists[idx].Items.Clear()
		for _, v := range usrs {
			m[v.UserId] = v
			c.Lists[idx].Items.MoveToBack(list.ArrayItem{MainText: v.Username,
				SecondaryText: strconv.FormatUint(v.UserId, 10)})
		}
	})

}
func (c *Chat) NewEventNotification(e IncomingEvent) (isErr bool) {
	ll := c.Lists[4].Items.(*list.ArrayList)
	errstr := "Success"
	if len(e.ErrorMsg) != 0 {
		errstr = "Error: " + e.ErrorMsg
		isErr = true
		c.Logger.Println(errstr)
	}
	en := ll.NewItem(
		[2]tcell.Color{tcell.ColorBlue, tcell.ColorRed},
		e.Event,
		errstr,
	)
	c.Lists[4].Items.MoveToBack(en)
	c.Logger.Println(en, "event:", e, "new event notification")
	return isErr
}
