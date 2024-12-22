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
	RoomId          uint64
	RoomName        string
	IsGroup         bool
	IsAdmin         bool
	CreatedByUserId uint64
	Users           map[uint64]User // user id to users in this room
	lastMessageID   uint64
	RoomType        string // Group or Duo
	Messages        map[int]*list.List
	MsgPageIdBack   int // for old messages
	MsgPageIdFront  int // for new messages
	RoomItem        list.ListItem
}

// Rooms linked list
type Chat struct {
	App            *tview.Application
	UserId         uint64
	Conn           *websocket.Conn
	RoomMsgs       map[uint64]*Room // room id to room
	DuoUsers       map[uint64]*User // user id to users that Duo-only
	BlockedUsers   map[uint64]User  // user id to blocked
	FoundUsers     map[uint64]User  // user id to found users
	Lists          [6]*list.List    // rooms[0], BlockedUsers[1], DuoUsers[2], FoundUsers[3], navigation[4],events[5],
	currentRoom    *Room            // current selected Room
	CurrentText    string           // current text for user search || set room name ||
	LastNavigation string
	MainFlex       *tview.Flex
	NavText        [16]string
	NavState       int
	FindUsersForm  *tview.Form
	SendMsgBtn     *tview.Form
	InputField     *tview.Form
	stopeventUI    bool
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

		ll := list.NewArrayList()
		rm.MsgPageIdBack--
		prevpgn := list.ArrayItem{ArrList: ll, SecondaryText: "Previos Page: " + strconv.Itoa(rm.MsgPageIdBack-1),
			Color: [2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue}}
		ll.MoveToBack(prevpgn)

		for i := range len(msgsv) {
			rm.lastMessageID = msgsv[i].MessageId

			e := list.ArrayItem{ArrList: ll, Color: [2]tcell.Color{tcell.ColorWhite, tcell.ColorGray},
				MainText:      rm.Users[msgsv[i].UserId].Username + ": " + msgsv[i].MessagePayload,
				SecondaryText: "UserId: " + strconv.FormatUint(msgsv[i].UserId, 10)}
			ll.MoveToFront(e)
		}

		nextpgn := list.ArrayItem{ArrList: ll, SecondaryText: "Next Page: " + strconv.Itoa(rm.MsgPageIdBack+1),
			Color: [2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue}}
		ll.MoveToFront(nextpgn)
		l := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront(), Selector: c}
		rm.Messages[rm.MsgPageIdBack] = l
	}

}
func (c *Chat) NewMessageEvent(msg Message) {
	rm, ok := c.RoomMsgs[msg.RoomId]
	if ok {
		l, ok2 := rm.Messages[rm.MsgPageIdFront]
		if ok2 {
			if l.Items.Len() >= 30 {

				rm.MsgPageIdFront++
				nextpgn := list.ArrayItem{SecondaryText: "Next Page: " + strconv.Itoa(rm.MsgPageIdFront),
					Color: [2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue}}
				l.Items.MoveToFront(nextpgn)

				ll := list.NewArrayList()
				prevpgn := list.ArrayItem{SecondaryText: "Previos Page: " + strconv.Itoa(rm.MsgPageIdFront-1),
					Color: [2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue}}
				ll.MoveToBack(prevpgn)

				lst := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront(), Selector: c}
				rm.Messages[rm.MsgPageIdFront] = lst

			} else {
				l.Items.MoveToFront(list.ArrayItem{Color: [2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue},
					MainText:      rm.Users[msg.UserId].Username + ": " + msg.MessagePayload,
					SecondaryText: "UserId: " + strconv.FormatUint(msg.UserId, 10)})

			}
		} else {
			ll := list.NewArrayList()
			prevpgn := list.ArrayItem{SecondaryText: "Previos Page: " + strconv.Itoa(rm.MsgPageIdFront),
				Color: [2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue}}
			ll.MoveToBack(prevpgn)
			lst := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront(), Selector: c}
			rm.Messages[rm.MsgPageIdFront] = lst
		}

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
	c.RoomMsgs[rmsv.RoomId] = &Room{Users: make(map[uint64]User), Messages: make(map[int]*list.List), RoomName: rmsv.RoomName, RoomId: rmsv.RoomId}
	rm := c.RoomMsgs[rmsv.RoomId]

	//fill room with users
	for i := range len(rmsv.Users) {
		rm.Users[rmsv.Users[i].UserId] = rmsv.Users[i]
	}
	// set new room at the top
	c.Lists[0].Items.MoveToFront(&list.LinkedItems{MainText: rm.RoomName, SecondaryText: strconv.FormatUint(rm.RoomId, 10)})
	if rmsv.IsGroup {
		rm.RoomType = "Group"
	} else {
		rm.RoomType = "Duo"
	}
}
func main() {
	chat := NewChat()
	chat.App = tview.NewApplication()

	chat.MainFlex.AddItem(chat.Lists[0], 0, 1, true)

	if err := chat.App.SetRoot(chat.MainFlex, true).Run(); err != nil {
		panic(err)
	}
}
