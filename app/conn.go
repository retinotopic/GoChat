package app

import (
	"context"

	// "github.com/rogpeppe/go-internal/lockedfile"
	"log"
	"net/http"

	// "os"
	"strconv"
	"time"

	"encoding/base64"

	json "github.com/bytedance/sonic"
	"github.com/gdamore/tcell/v2"

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
			_, b, err := c.Conn.Read(context.TODO())
			if err != nil {
				c.Logger.Println("Event read", err)
				close(c.errch)
				c.Conn.CloseNow()
				return
			}
			c.Logger.Println("New event", b)
			// if msgType == websocket.MessageText && len(b) > 0 {
			if c.WaitForTest {
				c.Mtx.Lock()
				c.ProcIncomingEv(b)
				c.Mtx.Unlock()
				<-c.TestCh
			} else {
				c.App.QueueUpdateDraw(func() {
					c.ProcIncomingEv(b)
				})
			}
			// }
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

func (c *Chat) NewEventNotification(e EventInfo, data []byte) (isSkip bool) {
	c.Logger.Println("Event", "NEW EVENT NOTIFICATION")
	addinfo := " "
	if e.UserId == c.UserId {
		addinfo = e.Event + " by me: " + c.Username
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
		addinfo,
		errstr,
	)

	if c.IsDebug {
		c.Hash.Reset()
		c.Hash.Write(data)
		checksum := base64.StdEncoding.EncodeToString(c.Hash.Sum(nil))

		c.Checksums = append(c.Checksums, checksum, ":", e.Event, ":")

		if c.TestLogger != nil {
			c.TestLogger.Println("Sgn:", checksum)
		}

	}
	c.Lists[4].Items.MoveToBack(en)

	return isSkip
}

func (c *Chat) ProcIncomingEv(b []byte) {
	c.Logger.Println("still here")
	if c.UserId == 0 {
		u := User{}
		err := json.Unmarshal(b, &u)
		if err != nil {
			panic(err)
		}
		c.Username = u.Username
		c.UserId = u.UserId

	}
	c.Logger.Println("still here1")

	e := EventInfo{}
	err := json.Unmarshal(b, &e)
	if err != nil {
		c.Logger.Println(err)
		return
	}
	c.Logger.Println("still here2")
	isSkip := c.NewEventNotification(e, b)
	if isSkip {
		return
	}
	c.Logger.Println("still here3")

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
		c.Logger.Println("EventUX", rm)
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
}
