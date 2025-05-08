package app

import (
	"log"

	json "github.com/bytedance/sonic"

	// "log"
	"strconv"
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

	state       LoadingState
	Logger      *log.Logger
	errch       chan error
	SendEventCh chan EventInfo
}

func NewChat(username, url string, maxMsgsOnPage int, debug bool, logger *log.Logger) (chat *Chat, err <-chan error) {
	c := &Chat{Logger: logger}
	c.MaxMsgsOnPage = maxMsgsOnPage
	c.ParseAndInitUI()
	c.PreLoadNavigation()
	err = c.TryConnect(username, url)
	return c, err
}

func (c *Chat) Run() {
	c.Logger.Println("App started")
	go c.ProcessEvents()
	c.errch <- c.App.SetRoot(c.MainFlex, true).Run()
}

func (c *Chat) ProcessEvents() {
	i := 0
	for {
		select {
		case e := <-c.SendEventCh:
			b, err := json.Marshal(e)
			if err != nil {
				c.Logger.Println(err, "process events marshal")
				continue

			} else if c.state.InProgressCount.Load() < 15 {

				c.state.InProgressCount.Add(1)
				go func() {
					err := WriteTimeout(time.Second*15, c.Conn, b)
					if err != nil {
						c.state.InProgressCount.Add(-1)
					}
				}()
			}
		case <-c.errch:
			return
		default:
			c.App.QueueUpdateDraw(func() {
				if c.state.InProgressCount.Load() > 0 {
					spinChar := c.state.spinner[i%len(c.state.spinner)]
					text := spinChar + " " + strconv.Itoa(int(c.state.InProgressCount.Load())) + " " + c.state.message
					item := c.Lists[0].Items.GetBack()
					item.SetSecondaryText(text)
					item.SetColor(tcell.ColorRed, 1)
					i++
					if i == len(c.state.spinner) {
						i = 0
					}
				} else {
					item := c.Lists[0].Items.GetBack()
					item.SetSecondaryText("")
					item.SetColor(tcell.ColorGrey, 1)
				}
			})
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
				rm.Users[msgsv[i].UserId].Username+": "+msgsv[i].MessagePayload,
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
					rm.Users[msg.UserId].Username+": "+msg.MessagePayload,
					strconv.FormatUint(msg.UserId, 10),
				)
				l.Items.MoveToBack(item)
				c.Logger.Println(msg.UserId, rm.RoomId, "newmsg")

			}
			// set room at the top
			c.Lists[1].Items.MoveToBack(rm.RoomItem)
			it := c.Lists[1].Items.GetBack()
			it.SetColor(tcell.ColorLightYellow, 0)
			c.Logger.Println(c.Lists[1].Items.Len())
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
			c.Logger.Println(rm)
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
				c.Logger.Println(rm == nil, rm.RoomItem == nil, rm.RoomItem)
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
		c.Logger.Println(u.Username, "username::::")
		rm.Users[u.UserId] = u
		c.Logger.Println(rmsv.RoomId, " roomid ", " user ", u.Username, u.UserId)
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
