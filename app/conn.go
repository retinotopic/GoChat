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

type EventInfo struct {
	Event    string                `json:"Event"`
	ErrorMsg string                `json:"ErrorMsg"`
	UserId   uint64                `json:"UserId"`
	Type     int                   `json:"Type"`
	Data     json.NoCopyRawMessage `json:"Data"`
}

func (c *Chat) Connect(keyIdent, url string) <-chan error {
	c.errch = make(chan error)
	hd := http.Header{}
	cookie := http.Cookie{
		Name:     "username",
		Value:    keyIdent,
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
		close(c.errch)
		c.Logger.Println("Event", err)
		return c.errch
	}

	go func() {
		for {
			msgType, b, err := c.Conn.Read(context.TODO())
			if err != nil {
				c.Logger.Println("Event", err)
				close(c.errch)
				c.Conn.CloseNow()
				return
			}
			if msgType == websocket.MessageText && len(b) > 0 {
				c.ProcessIncomingEvent(b)
			}
		}
	}()

	return c.errch

}
func (c *Chat) FillUsers(usrs []User, idx int) {
	c.Lists[idx].Items.Clear()
	for _, v := range usrs {
		lst := c.Lists[idx].Items
		lst.MoveToBack(lst.NewItem([2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
			v.Username, strconv.FormatUint(v.UserId, 10)))
	}
}

func (c *Chat) NewEventNotification(e EventInfo) (isSkip bool) {
	c.Logger.Println("Event", "NEW EVENT NOTIFICATION")
	addinfo := " "
	if e.UserId == c.UserId {
		c.InProgressCount.Add(-1)
		addinfo = " by me: " + c.Username
	}
	ll := c.Lists[4].Items.(*list.ArrayList)
	errstr := "Success"
	if len(e.ErrorMsg) != 0 {
		errstr = "Error: " + e.ErrorMsg
		isSkip = true
		c.Logger.Println("Event", errstr)
	}
	if e.Type == 4 {
		isSkip = true
	}
	en := ll.NewItem(
		[2]tcell.Color{tcell.ColorBlue, tcell.ColorRed},
		e.Event+addinfo,
		errstr,
	)
	c.Lists[4].Items.MoveToBack(en)
	return isSkip
}

func (c *Chat) ProcessIncomingEvent(b []byte) {
	c.App.QueueUpdateDraw(func() {
		if c.UserId == 0 {
			u := User{}
			err := json.Unmarshal(b, &u)
			if err != nil {
				panic(err)
			}
			c.Username = u.Username
			c.UserId = u.UserId
		}
		c.Logger.Println("Event", b, "CONN READ")
		e := EventInfo{}
		err := json.Unmarshal(b, &e)
		if err != nil {
			return
		}
		isSkip := c.NewEventNotification(e)
		if isSkip {
			return
		}

		switch e.Event {
		case "Get Messages From Room":
			var msgs []Message
			err = json.Unmarshal(e.Data, &msgs)
			if err != nil {
				return
			}
			c.LoadMessagesEvent(msgs)
			return
		case "Find Users":
			var usrs []User
			err := json.Unmarshal(e.Data, &usrs)
			if err != nil {
				return
			}
			c.Logger.Println("Event", e.Type, "Find Users")
			c.FillUsers(usrs, 5)
			return
		case "Get Blocked Users":
			var usrs []User
			err := json.Unmarshal(e.Data, &usrs)
			if err != nil {
				return
			}
			c.Logger.Println("Event", e.Type, "Get Blocked Users")
			c.FillUsers(usrs, 7)
			return
		case "Change Username":
			u := User{}
			err := json.Unmarshal(e.Data, &u)
			if err != nil {
				return
			}
			c.Username = u.Username
			return
		}

		switch e.Type {
		case 1:
			msg := Message{}
			err := json.Unmarshal(e.Data, &msg)
			if err != nil {
				return
			}
			c.NewMessageEvent(msg)
			return
		case 2:
			rm := []RoomServer{}
			err := json.Unmarshal(e.Data, &rm)
			if err != nil {
				log.Fatalln(err)
				return
			}
			c.ProcessRoom(rm)
		case 3:
			usr := User{}
			err := json.Unmarshal(e.Data, &usr)
			if err != nil {
				log.Fatalln(err)
				return
			}
			c.Username = usr.Username
		}
	})
}
