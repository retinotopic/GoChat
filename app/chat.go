package app

import (
	"errors"
	"log"
	// "log"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
	"golang.org/x/sync/errgroup"
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
	UrlConn  string
	Conn     *websocket.Conn
	MainFlex *tview.Flex

	MaxMsgsOnPage int

	RoomMsgs map[uint64]*RoomInfo // room id to room
	EventMap map[list.Content]Event

	DuoUsers     map[uint64]User // user id to users that Duo-only
	BlockedUsers map[uint64]User // user id to blocked
	FoundUsers   map[uint64]User // user id to found users

	Lists [10]*list.List /* sidebar[0],rooms[1],menu[2]
	input[3],Events[4],FoundUsers[5],DuoUsers[6]
	BlockedUsers[7],RoomUsers[8],Boolean[9]*/

	CurrentRoom *RoomInfo // current selected Room
	CurrentText string    // current text for user search || set room name || message

	NavState      int
	IsInputActive bool

	Logger      *log.Logger
	errgroup    errgroup.Group
	errch       chan error
	SendEventCh chan []byte
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
	defer func(err error) {
		c.errch <- errors.New("UI is crashed")
	}(c.App.SetRoot(c.MainFlex, true).Run())
}

func (c *Chat) ProcessEvents() {
	i := 0
	var NotIdle bool
	for {
		select {
		case b := <-c.SendEventCh:
			c.errgroup.Go(func() error {
				return WriteTimeout(time.Second*15, c.Conn, b)
			})
		case <-c.errch:
			return
		}
		if state.InProgressCount > 0 {
			c.App.QueueUpdateDraw(func() {
				spinChar := state.spinner[i%len(state.spinner)]
				text := spinChar + " " + strconv.Itoa(state.InProgressCount) + " " + state.message
				item := c.Lists[0].Items.GetBack()
				item.SetSecondaryText(text)
				item.SetColor(tcell.ColorRed, 1)
			})
			i++
			if i == len(state.spinner) {
				i = 0
			}
			NotIdle = true
		} else if NotIdle {
			c.App.QueueUpdateDraw(func() {
				item := c.Lists[0].Items.GetBack()
				item.SetSecondaryText(strconv.Itoa(state.InProgressCount) + " " + state.message)
				item.SetColor(tcell.ColorGrey, 1)
			})
			NotIdle = false
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
				"Prev Page: "+strconv.Itoa(rm.MsgPageIdBack-1),
				"",
			)

			ll.MoveToBack(prevpgn)

			for i := range len(msgsv) {
				rm.LastMessageID = msgsv[i].MessageId

				e := ll.NewItem(
					[2]tcell.Color{tcell.ColorWhite, tcell.ColorGray},
					rm.Users[msgsv[i].UserId].Username+": "+msgsv[i].MessagePayload,
					"UserId: "+strconv.FormatUint(msgsv[i].UserId, 10),
				)
				ll.MoveToBack(e)
			}

			nextpgn := ll.NewItem(
				[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
				"",
				"Next Page: "+strconv.Itoa(rm.MsgPageIdBack+1),
			)

			ll.MoveToBack(nextpgn)
			l := list.NewList(ll, c.OptionRoom, strconv.Itoa(rm.MsgPageIdBack))
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
						"",
						"Next Page: "+strconv.Itoa(rm.MsgPageIdFront),
					)
					l.Items.MoveToBack(nextpgn)

					ll := list.NewArrayList(c.MaxMsgsOnPage)
					prevpgn := ll.NewItem(
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						"",
						"Prev Page: "+strconv.Itoa(rm.MsgPageIdFront-1),
					)
					ll.MoveToBack(prevpgn)

					lst := list.NewList(ll, c.OptionRoom, strconv.Itoa(rm.MsgPageIdFront))
					rm.Messages[rm.MsgPageIdFront] = lst

				} else {
					msg := l.Items.NewItem(
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						rm.Users[msg.UserId].Username+": "+msg.MessagePayload,
						"UserId: "+strconv.FormatUint(msg.UserId, 10),
					)
					l.Items.MoveToBack(msg)

				}
			} else {
				ll := list.NewArrayList(c.MaxMsgsOnPage)
				prevpgn := ll.NewItem(
					[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
					"",
					"Prev Page: "+strconv.Itoa(rm.MsgPageIdFront),
				)
				ll.MoveToBack(prevpgn)
				lst := list.NewList(ll, c.OptionRoom, strconv.Itoa(rm.MsgPageIdFront))
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
			for i := range len(rmsv.Users) {
				if rmsv.Users[i].UserId == c.UserId {
					c.Lists[1].Items.Remove(c.RoomMsgs[rmsv.RoomId].RoomItem)
					delete(c.RoomMsgs, rmsv.RoomId)
					break
				}
				if rmsv.Users[i].RoomToggle {
					delete(rm.Users, rmsv.Users[i].UserId)
				} else {
					rm.Users[rmsv.Users[i].UserId] = rmsv.Users[i]
				}
			}
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
		for i := range len(rmsv.Users) {
			rm.Users[rmsv.Users[i].UserId] = rmsv.Users[i]
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
