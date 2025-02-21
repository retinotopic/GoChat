package main

import (
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

type Room struct {
	RoomId   uint64
	RoomName string

	IsGroup bool
	IsAdmin bool

	CreatedByUserId uint64

	Users         map[uint64]User // user id to users in this room
	lastMessageID uint64
	RoomType      string // Group or Duo
	Messages      map[int]*list.List

	MsgPageIdBack  int // for old messages
	MsgPageIdFront int // for new messages

	RoomItem list.ListItem
}

// Rooms linked list
type Chat struct {
	App      *tview.Application
	UserId   uint64
	Conn     *websocket.Conn
	MainFlex *tview.Flex

	RoomMsgs         map[uint64]*Room // room id to room
	SendEventMap     map[string]SendEvent
	NavigateEventMap map[string]NavigateEvent

	DuoUsers     map[uint64]User // user id to users that Duo-only
	BlockedUsers map[uint64]User // user id to blocked
	FoundUsers   map[uint64]User // user id to found users

	Lists []*list.List /* menu[0],rooms[1],navigation[2]
	input[3],events[4],FoundUsers[5],DuoUsers[6]
	BlockedUsers[7],RoomUsers[8],SendEvents[9],Boolean[10]*/

	currentRoom *Room  // current selected Room
	CurrentText string // current text for user search || set room name || message

	NavState      int
	stopeventUI   bool
	IsInputActive bool
}

func NewChat() *Chat {
	c := &Chat{}
	for i := range len(c.Lists) {
		c.Lists[i] = &list.List{Box: tview.NewBox()}
	}
	return c
}
func (c *Chat) LoadMessagesEvent(msgsv []Message) {
	if len(msgsv) == 0 {
		return
	}
	rm, ok := c.RoomMsgs[msgsv[0].RoomId]
	if ok {
		c.App.QueueUpdate(func() {
			ll := list.NewArrayList()
			rm.MsgPageIdBack--
			prevpgn := list.NewArrayItem(
				ll,
				[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
				"Prev Page: "+strconv.Itoa(rm.MsgPageIdBack-1),
				"",
			)

			ll.MoveToBack(prevpgn)

			for i := range len(msgsv) {
				rm.lastMessageID = msgsv[i].MessageId

				e := list.NewArrayItem(
					ll,
					[2]tcell.Color{tcell.ColorWhite, tcell.ColorGray},
					rm.Users[msgsv[i].UserId].Username+": "+msgsv[i].MessagePayload,
					"UserId: "+strconv.FormatUint(msgsv[i].UserId, 10),
				)
				ll.MoveToFront(e)
			}

			nextpgn := list.NewArrayItem(
				ll,
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

				if l.Items.Len() >= 30 {
					lv := l.Items.(*list.ArrayList) // to correct type assertion l.Iterms MUST always be *list.ArrayList
					rm.MsgPageIdFront++
					nextpgn := list.NewArrayItem(
						lv,
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						"",
						"Next Page: "+strconv.Itoa(rm.MsgPageIdFront),
					)
					l.Items.MoveToFront(nextpgn)

					ll := list.NewArrayList()
					prevpgn := list.NewArrayItem(
						ll,
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						"",
						"Prev Page: "+strconv.Itoa(rm.MsgPageIdFront-1),
					)
					ll.MoveToBack(prevpgn)

					lst := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront(), Option: c.OptionRoom}
					rm.Messages[rm.MsgPageIdFront] = lst

				} else {

					lv := l.Items.(*list.ArrayList)
					msg := list.NewArrayItem(
						lv,
						[2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
						rm.Users[msg.UserId].Username+": "+msg.MessagePayload,
						"UserId: "+strconv.FormatUint(msg.UserId, 10),
					)
					l.Items.MoveToFront(msg)

				}
			} else {
				ll := list.NewArrayList()
				prevpgn := list.NewArrayItem(
					ll,
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
					delete(c.RoomMsgs, rmsv.RoomId)
					c.Lists[6].Items.Remove()
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
		c.RoomMsgs[rmsv.RoomId] = &Room{Users: make(map[uint64]User), Messages: make(map[int]*list.List), RoomName: rmsv.RoomName, RoomId: rmsv.RoomId}
		rm := c.RoomMsgs[rmsv.RoomId]

		//fill room with users
		for i := range len(rmsv.Users) {
			rm.Users[rmsv.Users[i].UserId] = rmsv.Users[i]
		}
		// set new room at the top
		c.Lists[1].Items.MoveToFront(&list.LinkedItem{MainText: rm.RoomName, SecondaryText: strconv.FormatUint(rm.RoomId, 10)})
		if rmsv.IsGroup {
			rm.RoomType = "Group"
		} else {
			rm.RoomType = "Duo"
		}
	})
}
func main() {
	chat := NewChat()
	chat.App = tview.NewApplication()

	chat.MainFlex.AddItem(chat.Lists[0], 0, 1, true)
	go chat.StartEventUILoop()
	if err := chat.App.SetRoot(chat.MainFlex, true).Run(); err != nil {
		panic(err)
	}
}
