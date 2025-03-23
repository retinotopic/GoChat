package app

import (
	"image/color"
	"strconv"

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
	UrlConn  string
	Conn     *websocket.Conn
	MainFlex *tview.Flex

	MaxMsgsOnPage int

	RoomMsgs map[uint64]*RoomInfo // room id to room
	EventMap map[list.Content]Event

	DuoUsers     map[uint64]User // user id to users that Duo-only
	BlockedUsers map[uint64]User // user id to blocked
	FoundUsers   map[uint64]User // user id to found users

	Lists [11]*list.List /* menu[0],rooms[1],events[2]
	input[3],recentEvents[4],FoundUsers[5],DuoUsers[6]
	BlockedUsers[7],RoomUsers[8],Boolean[10]*/

	CurrentRoom *RoomInfo // current selected Room
	CurrentText string    // current text for user search || set room name || message

	NavState      int
	stopeventUI   bool
	IsInputActive bool

	ToSend *SendEvent
}

func NewChat(username, url string, maxMsgsOnPage int) (chat *Chat, err error) {
	c := &Chat{}
	c.MaxMsgsOnPage = maxMsgsOnPage
	c.ToSend = &SendEvent{}
	for i := range len(c.Lists) {
		c.Lists[i] = list.NewList()
		c.Lists[i].Items = list.NewArrayList(c.MaxMsgsOnPage)
	}
	c.Lists[1].Items = list.NewLinkedList(250)

	c.ParseAndInitUI()
	c.PreLoadNavigation()
	err = c.TryConnect(username, url)
	if err != nil {
		return nil, err
	}
	return c, err
}
func (c *Chat) Run() error {
	go c.StartEventUILoop()
	c.App.Stop()
	if err := c.App.SetRoot(c.MainFlex, true).Run(); err != nil {
		c.stopeventUI = true
		return err
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
				ll.MoveToFront(e)
			}

			nextpgn := ll.NewItem(
				[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
				"",
				"Next Page: "+strconv.Itoa(rm.MsgPageIdBack+1),
			)

			ll.MoveToFront(nextpgn)
			l := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront(), Option: c.OptionRoom}
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
					lv := l.Items.(*list.ArrayList) // to correct type assertion l.Items MUST always be *list.ArrayList
					rm.MsgPageIdFront++
					nextpgn := lv.NewItem(
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						"",
						"Next Page: "+strconv.Itoa(rm.MsgPageIdFront),
					)
					l.Items.MoveToFront(nextpgn)

					ll := list.NewArrayList(c.MaxMsgsOnPage)
					prevpgn := ll.NewItem(
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						"",
						"Prev Page: "+strconv.Itoa(rm.MsgPageIdFront-1),
					)
					ll.MoveToBack(prevpgn)

					lst := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront(), Option: c.OptionRoom}
					rm.Messages[rm.MsgPageIdFront] = lst

				} else {

					lv := l.Items.(*list.ArrayList)
					msg := lv.NewItem(
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						rm.Users[msg.UserId].Username+": "+msg.MessagePayload,
						"UserId: "+strconv.FormatUint(msg.UserId, 10),
					)
					l.Items.MoveToFront(msg)

				}
			} else {
				ll := list.NewArrayList(c.MaxMsgsOnPage)
				prevpgn := ll.NewItem(
					[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
					"",
					"Prev Page: "+strconv.Itoa(rm.MsgPageIdFront),
				)
				ll.MoveToBack(prevpgn)
				lst := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront(), Option: c.OptionRoom}
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
		ll, ok := c.Lists[1].Items.(*list.LinkedList) // MUST BE LINKED LIST
		if ok {
			rmitem := ll.NewItem([2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue}, rm.RoomName, strconv.FormatUint(rm.RoomId, 10))
			c.Lists[1].Items.MoveToFront(rmitem)
			rm.RoomItem = rmitem
		} else {
			panic("its not linked list, wtf?")
		}
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
