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
	RoomType       string             // Group or Duo
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
			if c.state.InProgressCount.Load() > 0 {
				c.App.QueueUpdateDraw(func() {
					spinChar := c.state.spinner[i%len(c.state.spinner)]
					text := spinChar + " " + strconv.Itoa(int(c.state.InProgressCount.Load())) + " " + c.state.message
					item := c.Lists[0].Items.GetBack()
					item.SetSecondaryText(text)
					item.SetColor(tcell.ColorRed, 1)
				})
				i++
				if i == len(c.state.spinner) {
					i = 0
				}
			} else {
				c.App.QueueUpdateDraw(func() {
					item := c.Lists[0].Items.GetBack()
					item.SetSecondaryText("")
					item.SetColor(tcell.ColorGrey, 1)
				})
			}
		}
	}
}

func (c *Chat) LoadMessagesEvent(msgsv []Message) {
	if len(msgsv) == 0 {
		return
	}
	rm, ok := c.RoomMsgs[msgsv[0].RoomId]
	if ok {
		c.App.QueueUpdate(func() {
			ll := list.NewArrayList(c.MaxMsgsOnPage)
			rm.MsgPageIdBack--
			prevpgn := ll.NewItem(
				[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
				"Prev Page",
				strconv.Itoa(rm.MsgPageIdBack-1),
			)

			ll.MoveToBack(prevpgn)

			for i := range len(msgsv) {
				rm.LastMessageID = msgsv[i].MessageId

				e := ll.NewItem(
					[2]tcell.Color{tcell.ColorWhite, tcell.ColorGray},
					rm.Users[msgsv[i].UserId].Username+": "+msgsv[i].MessagePayload,
					strconv.FormatUint(msgsv[i].UserId, 10),
				)
				ll.MoveToBack(e)
			}

			nextpgn := ll.NewItem(
				[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
				"Next Page",
				strconv.Itoa(rm.MsgPageIdBack+1),
			)

			ll.MoveToBack(nextpgn)
			l := list.NewList(ll, c.OptionPagination, strconv.Itoa(rm.MsgPageIdBack))
			rm.Messages[rm.MsgPageIdBack] = l
		})
	}

}

func (c *Chat) NewMessageEvent(msg Message) {
	rm, ok := c.RoomMsgs[msg.RoomId]
	if ok {
		c.App.QueueUpdate(func() {
			l, ok2 := rm.Messages[rm.MsgPageIdFront]
			if ok2 {
				if l.Items.Len() >= c.MaxMsgsOnPage {
					rm.MsgPageIdFront++
					nextpgn := l.Items.NewItem(
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						"Next Page",
						strconv.Itoa(rm.MsgPageIdFront),
					)
					l.Items.MoveToBack(nextpgn)
					///// OLD PAGE

					//// NEW PAGE
					ll := list.NewArrayList(c.MaxMsgsOnPage)
					prevpgn := ll.NewItem(
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						"Prev Page",
						strconv.Itoa(rm.MsgPageIdFront-1),
					)
					ll.MoveToBack(prevpgn)

					lst := list.NewList(ll, c.OptionPagination, strconv.Itoa(rm.MsgPageIdFront))
					rm.Messages[rm.MsgPageIdFront] = lst

				} else {
					msg := l.Items.NewItem(
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						rm.Users[msg.UserId].Username+": "+msg.MessagePayload,
						strconv.FormatUint(msg.UserId, 10),
					)
					l.Items.MoveToBack(msg)

				}
				// set room at the top
				c.Lists[1].Items.MoveToBack(rm.RoomItem)
			} else {
				ll := list.NewArrayList(c.MaxMsgsOnPage)
				prevpgn := ll.NewItem(
					[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
					"Prev Page",
					strconv.Itoa(rm.MsgPageIdFront),
				)
				ll.MoveToBack(prevpgn)
				lst := list.NewList(ll, c.OptionPagination, strconv.Itoa(rm.MsgPageIdFront))
				rm.Messages[rm.MsgPageIdFront] = lst
			}
		})
	}

}
func (c *Chat) ProcessRoom(rmsvs []RoomServer) {
	for _, rmsv := range rmsvs {
		rm, ok := c.RoomMsgs[rmsv.RoomId]
		if ok {
			rm.RoomName = rmsv.RoomName
			isKicked := false
			c.UserBuf = c.UserBuf[:0]
			for i, u := range rmsv.Users {
				if u.RoomToggle {
					if !rm.IsGroup {
						delete(c.DuoUsers, u.UserId)
					}
					if u.UserId == c.UserId {
						isKicked = true
					}
					delete(rm.Users, rmsv.Users[i].UserId)
				} else {
					rm.Users[u.UserId] = u
				}
			}
			if isKicked {
				c.Lists[1].Items.Remove(rm.RoomItem)
				clear(rm.Users)
				clear(rm.Messages)
				delete(c.RoomMsgs, rmsv.RoomId)
			}
			for _, v := range c.DuoUsers {
				c.UserBuf = append(c.UserBuf, v)
			}
			c.FillUsers(c.UserBuf, 6)

		} else {
			c.AddRoom(rmsv)
		}
	}
}

func (c *Chat) AddRoom(rmsv RoomServer) {
	c.App.QueueUpdate(func() {

		c.RoomMsgs[rmsv.RoomId] = &RoomInfo{Users: make(map[uint64]User),
			Messages: make(map[int]*list.List), RoomName: rmsv.RoomName,
			RoomId: rmsv.RoomId}

		rm := c.RoomMsgs[rmsv.RoomId]
		//fill room with users
		for _, u := range rmsv.Users {
			rm.Users[u.UserId] = u
			if u.UserId != c.UserId && !rmsv.IsGroup {
				c.DuoUsers[u.UserId] = u
			}
		}
		// set new room at the top
		rmitem := c.Lists[1].Items.NewItem([2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
			rm.RoomName, strconv.FormatUint(rm.RoomId, 10))
		c.Lists[1].Items.MoveToBack(rmitem)
		rm.RoomItem = rmitem

		if rmsv.CreatedByUserId == c.UserId {
			rm.IsAdmin = true
		}
		if rmsv.IsGroup {
			rm.RoomType = "Group"
		} else {
			rm.RoomType = "Duo"
		}
	})
}
