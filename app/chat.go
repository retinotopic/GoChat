package app

import (
	"log"

	json "github.com/bytedance/sonic"

	// "log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

type RoomServer struct {
	RoomId          uint64 `json:"RoomId" `
	RoomName        string `json:"RoomName" `
	IsGroup         bool   `json:"IsGroup" `
	CreatedByUserId uint64 `json:"CreatedByUserId" `
	Users           []User `json:"Users" `
}

type RoomInfo struct {
	RoomId   uint64
	RoomName string

	IsGroup bool
	IsAdmin bool

	LastMessageID uint64

	Users          map[uint64]User    // user id to users in this room
	Messages       map[int]*list.List // id to the page that contains the messages
	MsgPageIdBack  int                // for old messages
	MsgPageIdFront int                // for new messages

	RoomItem list.ListItem
}

type Chat struct {
	App      *tview.Application
	UserId   uint64
	Username string
	UrlConn  string
	Conn     *websocket.Conn
	MainFlex *tview.Flex

	MaxMsgsOnPage int

	RoomMsgs map[uint64]*RoomInfo // room id to room
	EventMap map[list.Content]Event
	DuoUsers map[uint64]User

	Lists [10]*list.List /* sidebar[0],rooms[1],menu[2]
	input[3],Events[4],FoundUsers[5],DuoUsers[6]
	BlockedUsers[7],RoomUsers[8],Boolean[9]*/

	CurrentRoom *RoomInfo // current selected Room
	CurrentText string    // current text for user search || set room name || message

	NavState      int
	IsInputActive bool
	UserBuf       []User

	keyIdent        string
	url             string
	InProgressCount atomic.Int32
	Logger          *log.Logger
	errch           chan error
	SendEventCh     chan EventInfo
}

func NewChat(keyIdent, url string, maxMsgsOnPage int, debug bool, logger *log.Logger) (chat *Chat) {
	c := &Chat{Logger: logger, url: url, keyIdent: keyIdent}
	c.MaxMsgsOnPage = maxMsgsOnPage
	c.ParseAndInitUI()
	return c
}

func (c *Chat) TryConnect() <-chan error {
	c.PreLoadNavigation()
	errch := c.Connect(c.keyIdent, c.url)
	return errch
}

func (c *Chat) ProcessEvents() {
	c.Logger.Println("Event", "App started")
	go func() {
		err := c.App.SetRoot(c.MainFlex, true).Run()
		if err != nil {
			c.Conn.CloseNow()
		}
	}()
	for {
		select {
		case e := <-c.SendEventCh:
			b, err := json.Marshal(e)
			if err != nil {
				c.Logger.Println("Event", err, "process events marshal")
				continue

			} else if c.InProgressCount.Load() < 15 {
				c.InProgressCount.Add(1)
				go func() {
					err := WriteTimeout(time.Second*15, c.Conn, b)
					if err != nil {
						c.InProgressCount.Add(-1)
					}
				}()
			}
		case <-c.errch:
			return
		}
	}
}

func (c *Chat) LoadMessagesEvent(msgsv []Message) {
	if len(msgsv) == 0 {
		return
	}
	rm, ok := c.RoomMsgs[msgsv[0].RoomId]
	if ok {
		ll := list.NewArrayList(c.MaxMsgsOnPage)
		prevpgn := ll.NewItem(
			[2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
			"Prev Page",
			strconv.Itoa(rm.MsgPageIdBack-1),
		)

		ll.MoveToBack(prevpgn)

		for i := range len(msgsv) {
			rm.LastMessageID = msgsv[i].MessageId
			e := ll.NewItem(
				[2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
				msgsv[i].Username+": "+msgsv[i].MessagePayload,
				strconv.FormatUint(msgsv[i].UserId, 10),
			)
			ll.MoveToBack(e)
		}

		nextpgn := ll.NewItem(
			[2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
			"Next Page",
			strconv.Itoa(rm.MsgPageIdBack+1),
		)

		ll.MoveToBack(nextpgn)
		l := list.NewList(ll, c.OptionPagination, strconv.Itoa(rm.MsgPageIdBack), c.Logger)
		rm.Messages[rm.MsgPageIdBack] = l
		rm.MsgPageIdBack--
	}

}

func (c *Chat) NewMessageEvent(msg Message) {
	rm, ok := c.RoomMsgs[msg.RoomId]
	if ok {
		l, ok2 := rm.Messages[rm.MsgPageIdFront]
		if ok2 {

			if l.Items.Len() >= c.MaxMsgsOnPage {
				rm.MsgPageIdFront++
				nextpgn := l.Items.NewItem(
					[2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
					"Next Page",
					strconv.Itoa(rm.MsgPageIdFront),
				)
				l.Items.MoveToBack(nextpgn)
				///// OLD PAGE

				//// NEW PAGE
				ll := list.NewArrayList(c.MaxMsgsOnPage)
				prevpgn := ll.NewItem(
					[2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
					"Prev Page",
					strconv.Itoa(rm.MsgPageIdFront-1),
				)
				ll.MoveToBack(prevpgn)

				lst := list.NewList(ll, c.OptionPagination, strconv.Itoa(rm.MsgPageIdFront), c.Logger)
				rm.Messages[rm.MsgPageIdFront] = lst

			} else {
				item := l.Items.NewItem(
					[2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
					msg.Username+": "+msg.MessagePayload,
					strconv.FormatUint(msg.UserId, 10),
				)
				l.Items.MoveToBack(item)
				c.Logger.Println("Event", msg.UserId, rm.RoomId, "newmsg")

			}
			// set room at the top
			c.Lists[1].Items.MoveToBack(rm.RoomItem)
			it := c.Lists[1].Items.GetBack()
			it.SetColor(tcell.ColorLightYellow, 0)
			c.Logger.Println("Event", c.Lists[1].Items.Len())
		} else {
			ll := list.NewArrayList(c.MaxMsgsOnPage)
			prevpgn := ll.NewItem(
				[2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
				"Prev Page",
				strconv.Itoa(rm.MsgPageIdFront),
			)
			ll.MoveToBack(prevpgn)
			lst := list.NewList(ll, c.OptionPagination, strconv.Itoa(rm.MsgPageIdFront), c.Logger)
			rm.Messages[rm.MsgPageIdFront] = lst
		}
	}
}

func (c *Chat) ProcessRoom(rmsvs []RoomServer) {
	for _, rmsv := range rmsvs {
		rm, ok := c.RoomMsgs[rmsv.RoomId]
		if ok {
			c.Logger.Println("Event", rm.RoomId)
			rm.RoomName = rmsv.RoomName
			isKicked := false
			for _, u := range rmsv.Users {
				if u.RoomToggle {
					if !rm.IsGroup {
						delete(c.DuoUsers, u.UserId)
					}
					if u.UserId == c.UserId {
						isKicked = true
					}
					delete(rm.Users, u.UserId)
				} else {
					rm.Users[u.UserId] = u
				}
			}
			if isKicked {
				c.Logger.Println("Event", rm == nil, rm.RoomItem == nil, rm.RoomItem.GetMainText())
				c.Lists[1].Items.Remove(rm.RoomItem)
				clear(rm.Users)
				clear(rm.Messages)
				delete(c.RoomMsgs, rmsv.RoomId)
			}
			rm.RoomItem.SetMainText(rm.RoomName, 0)
		} else {
			c.AddRoom(rmsv)
		}
	}
}

func (c *Chat) AddRoom(rmsv RoomServer) {

	c.RoomMsgs[rmsv.RoomId] = &RoomInfo{Users: make(map[uint64]User),
		Messages: make(map[int]*list.List), RoomName: rmsv.RoomName,
		RoomId: rmsv.RoomId, IsGroup: rmsv.IsGroup}

	rm := c.RoomMsgs[rmsv.RoomId]
	//fill room with users
	for _, u := range rmsv.Users {
		c.Logger.Println("Event", "username: ", u.Username)
		rm.Users[u.UserId] = u
		c.Logger.Println("Event", rmsv.RoomId, " roomid ", " user ", u.Username, u.UserId)
		if u.UserId != c.UserId && !rmsv.IsGroup {
			c.DuoUsers[u.UserId] = u
		}
	}

	ll := list.NewArrayList(c.MaxMsgsOnPage)
	lst := list.NewList(ll, c.OptionPagination, strconv.Itoa(rm.MsgPageIdFront), c.Logger)
	rm.Messages[0] = lst
	rm.MsgPageIdBack--

	prevpgn := ll.NewItem(
		[2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
		"Prev Page",
		strconv.Itoa(rm.MsgPageIdBack),
	)
	lst.Items.MoveToBack(prevpgn)

	// set new room at the top
	rmitem := c.Lists[1].Items.NewItem([2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
		rm.RoomName, strconv.FormatUint(rm.RoomId, 10))
	c.Lists[1].Items.MoveToBack(rmitem)
	rm.RoomItem = rmitem

	if rmsv.CreatedByUserId == c.UserId {
		rm.IsAdmin = true
	}
}
