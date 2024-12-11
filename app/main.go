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
	MsgPageIdBack   int                      // for old messages
	MsgPageIdFront  int                      // for new messages
	PaginationBtns  func(item list.ListItem) // navigation between Room.Messages lists
	RoomItem        list.ListItem
}
type ListAndHandlers struct {
}

// Rooms linked list
type Chat struct {
	App           *tview.Application
	UserId        uint64
	Conn          *websocket.Conn
	RoomMsgs      map[uint64]*Room // room id to room
	DuoUsers      map[uint64]*User // user id to users that Duo-only
	BlockedUsers  map[uint64]User  // user id to blocked
	FoundUsers    map[uint64]User  // user id to found users
	Lists         [6]*list.List    // rooms panel , duo users,blocked users,found users,events, AND ALL NAVIGATION list
	currentRoom   *Room            // current selected Room
	CurrentText   string           // current text for user search || set room name ||
	MainFlex      *tview.Flex
	NavText       [13]string
	NavState      int
	FindUsersForm *tview.Form
	RoomMenuForm  *tview.Form
}

func NewChat() *Chat {
	c := &Chat{}
	for i := range len(c.Lists) {
		c.Lists[i] = &list.List{Box: tview.NewBox()}
	}
	return c
}
func (r *Chat) LoadMessagesEvent(msgsv []Message) {
	if len(msgsv) == 0 {
		return
	}
	rm, ok := r.RoomMsgs[msgsv[0].RoomId]
	if ok {
		ll := list.NewLinkedList()
		rm.MsgPageIdBack--
		prevpgn := &list.LinkedItems{SecondaryText: strconv.Itoa(rm.MsgPageIdBack - 1), Color: tcell.ColorBlue}
		ll.MoveToBack(prevpgn)
		for i := range len(msgsv) {
			rm.lastMessageID = msgsv[i].MessageId

			e := &list.LinkedItems{Color: tcell.ColorWhite, MainText: rm.Users[msgsv[i].UserId].Username + ": " + msgsv[i].MessagePayload, SecondaryText: "UserId: " + strconv.FormatUint(msgsv[i].UserId, 10)}
			ll.MoveToFront(e)
		}
		nextpgn := &list.LinkedItems{SecondaryText: strconv.Itoa(rm.MsgPageIdBack + 1), Color: tcell.ColorBlue}
		ll.MoveToFront(nextpgn)
		l := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront()}
		//l.Selector = rm.ch
		rm.Messages[rm.MsgPageIdBack] = l
	}

}
func (c *Chat) NewMessageEvent(msg Message) {
	rm, ok := c.RoomMsgs[msg.RoomId]
	if ok {
		l, ok2 := rm.Messages[rm.MsgPageIdFront]
		if ok2 {
			if l.Items.Len() >= 30 {

				rm.MsgPageIdFront++ // dont forget prev btn at creating new room
				nextpgn := &list.LinkedItems{SecondaryText: strconv.Itoa(rm.MsgPageIdBack), Color: tcell.ColorBlue}
				l.Items.MoveToFront(nextpgn)

				ll := list.NewLinkedList()
				prevpgn := &list.LinkedItems{SecondaryText: strconv.Itoa(rm.MsgPageIdBack - 1), Color: tcell.ColorBlue}
				ll.MoveToBack(prevpgn)
				lst := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront()}
				rm.Messages[rm.MsgPageIdFront] = lst
			} else {
				l.Items.MoveToFront(&list.LinkedItems{Color: tcell.ColorWhite, MainText: rm.Users[msg.UserId].Username + ": " + msg.MessagePayload, SecondaryText: "UserId: " + strconv.FormatUint(msg.UserId, 10)})
			}
		}

	}

}
func (c *Chat) ProcessRoom(rmsv RoomServer) {
	rm, ok := c.RoomMsgs[rmsv.RoomId]
	if ok {
		for i := range len(rmsv.Users) {
			if rmsv.Users[i].UserId == c.UserId {
				c.DeleteRoom(rmsv.RoomId)
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

func (c *Chat) DeleteRoom(roomid uint64) *Chat {
	rm, ok := c.RoomMsgs[roomid]
	if ok {
		if rm.RoomItem != nil && !rm.RoomItem.IsNil() {
			c.Lists[0].Items.Remove(rm.RoomItem) // deleting node in Rooms linked list
		}
		delete(c.RoomMsgs, roomid) // deleting *Room instance in map
	}
	return c
}
func (c *Chat) AddRoom(rmsv RoomServer) {
	//fill room with users
	c.RoomMsgs[rmsv.RoomId] = &Room{Users: make(map[uint64]User), Messages: make(map[int]*list.List), RoomName: rmsv.RoomName, RoomId: rmsv.RoomId}
	rm := c.RoomMsgs[rmsv.RoomId]
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
