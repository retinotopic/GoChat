package main

import (
	"strconv"
	"time"

	json "github.com/bytedance/sonic"
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
	CreatedByUserId uint64
	Users           map[uint64]User // user id to users in this room
	lastMessageID   uint64
	RoomType        string // Group or Duo
	Messages        map[int]*list.List
	MsgPageId       int
	PaginationBtns  *list.HandleSelect
	RoomItem        list.ListItem
}

// Rooms linked list
type Chat struct {
	App          *tview.Application
	UserId       uint64
	Conn         *websocket.Conn
	RoomMsgs     map[uint64]*Room // room id to room
	DuoUsers     map[uint64]*User // user id to users that Duo-only
	BlockedUsers map[uint64]User  // user id to blocked users
	FoundUsers   map[uint64]User  // user id to found users
	currentRoom  *Room            // current selected Room
	CurrentText  string           // current text for user search || set room name ||
	RoomsPanel   *list.List
	Pages        *tview.Pages
}

func NewChat() *Chat {
	return &Chat{
		RoomsPanel: &list.List{Box: tview.NewBox()},
	}
}
func (r *Chat) LoadMessagesEvent(msgsv []Message) {
	if len(msgsv) == 0 {
		return
	}
	rm, ok := r.RoomMsgs[msgsv[0].RoomId]
	if ok {
		ll := list.NewLinkedList()
		rm.MsgPageId--
		prevpgn := &list.LinkedItems{SecondaryText: strconv.Itoa(rm.MsgPageId - 1), Color: tcell.ColorBlue}
		ll.MoveToBack(prevpgn)
		for i := range len(msgsv) {
			e := &list.LinkedItems{Color: tcell.ColorWhite, MainText: rm.Users[msgsv[i].UserId].Username + ": " + msgsv[i].MessagePayload, SecondaryText: "UserId: " + strconv.FormatUint(msgsv[i].UserId, 10)}
			ll.MoveToFront(e)
		}
		nextpgn := &list.LinkedItems{SecondaryText: strconv.Itoa(rm.MsgPageId + 1), Color: tcell.ColorBlue}
		ll.MoveToFront(nextpgn)
		l := &list.List{Box: tview.NewBox().SetBorder(true), Items: ll, Current: ll.GetFront()}
		l.SelectedFunc = rm.PaginationBtns.OneOption
		rm.Messages[rm.MsgPageId] = l
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
			c.RoomsPanel.Items.Remove(rm.RoomItem) // deleting node in Rooms linked list
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
	c.RoomsPanel.Items.MoveToFront(&list.LinkedItems{MainText: rm.RoomName, SecondaryText: strconv.FormatUint(rm.RoomId, 10)})
	if rmsv.IsGroup {
		rm.RoomType = "Group"
	} else {
		rm.RoomType = "Duo"
	}
}
func main() {
	app := tview.NewApplication()

	chat := NewChat()
	mainpage := tview.NewPages()

	FindUsersForm := tview.NewForm().
		AddInputField("First name", "", 20, nil, func(text string) {
			chat.CurrentText = text
		}).
		AddButton("Find", func() {
			event := User{
				Event:    "FindUsers",
				Username: chat.CurrentText,
			}
			b, err := json.Marshal(event)
			if err != nil {
				WriteTimeout(time.Second*5, chat.Conn, b)
			}
		})

	msgInput := tview.NewInputField() // creating input for message
	SendMsg := func() {               // func for sending message
		msg := msgInput.GetText()
		event := Message{
			Event:          "SendMessage",
			MessagePayload: msg,
			RoomId:         chat.currentRoom.RoomId,
		}
		b, err := json.Marshal(event)
		if err != nil {
			WriteTimeout(time.Second*5, chat.Conn, b)
		}
	}
	msgInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			SendMsg()
		}
	})

	mainpage.AddPage("RoomPanel", chat.RoomsPanel, true, true)
	mainpage.AddPage("FindUsers", FindUsersForm, true, true)

	flexapp := tview.NewFlex().
		AddItem(chat.RoomsPanel, 0, 1, true).
		AddItem(mainpage, 0, 1, true)
	if err := app.SetRoot(flexapp, true).Run(); err != nil {
		panic(err)
	}
}
