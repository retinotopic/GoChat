package main

import (
	"container/list"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/coder/websocket"
	"github.com/gdamore/tcell/v2"
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
	Users           map[uint64]User // user id to User struct
	lastMessageID   *uint64
	RoomType        string // Group or Duo
	Messages        *LinkedListUI
	RoomLL          *list.Element // node in Rooms linked list (type Room)
}

// Rooms linked list
type Chat struct {
	MyId              uint64
	Conn              *websocket.Conn
	RoomMsgs          map[uint64]*Room // room id to room message page
	DuoRooms          map[uint64]*User // duo rooms that
	currentRoom       *Room            // type Room
	CurrentUserSearch []User
	RoomsPanel        *LinkedListUI
}

func NewChat() *Chat {
	return &Chat{
		RoomsPanel: &LinkedListUI{Box: tview.NewBox(), Items: list.New()},
	}
}

/*
	func (r *Chat) ProcessMessage(msgsv Message) *Chat {
		rm, ok := r.RoomMsgs[msgsv.RoomId]
		if ok {
			rm.Messages.InsertItem(int(msgsv.MessageId), msgsv.MessagePayload, msgsv.UserId)
		}
		return r
	}
*/
func (r *Chat) ProcessRoom(rmsv RoomServer) *Chat {

	rm, ok := r.RoomMsgs[rmsv.RoomId]
	if ok {
		for i := range len(rmsv.Users) {
			if rmsv.Users[i].UserId == r.MyId {
				r.DeleteRoom(rmsv.RoomId)
				break
			}
			if rmsv.Users[i].RoomToggle {
				delete(rm.Users, rmsv.Users[i].UserId)
			} else {
				rm.Users[rmsv.Users[i].UserId] = rmsv.Users[i]
			}
		}
	} else {
		r.AddRoom(rmsv)
	}

	return r
}

func (c *Chat) DeleteRoom(roomid uint64) *Chat {
	rm, ok := c.RoomMsgs[roomid]
	if ok {
		if rm.RoomLL != nil {
			c.RoomsPanel.Items.Remove(rm.RoomLL) // deleting node in Rooms linked list
		}
		delete(c.RoomMsgs, roomid) // deleting *Room instance in map
	}
	return c
}
func (c *Chat) AddRoom(rmsv RoomServer) *Chat {
	//fill room with users
	c.RoomMsgs[rmsv.RoomId] = &Room{Users: make(map[uint64]User)}
	rm := c.RoomMsgs[rmsv.RoomId]
	for i := range len(rmsv.Users) {
		rm.Users[rmsv.Users[i].UserId] = rmsv.Users[i]
	}
	// set new room at the top
	item := c.RoomsPanel.Items.PushFront(rm)
	rm.RoomLL = item

	// creating corresponding page for this room
	msgInput := tview.NewInputField() // creating input for message
	SendMsg := func() {               // func for sending message
		msg := msgInput.GetText()
		event := Message{
			Event:          "SendMessage",
			MessagePayload: msg,
			RoomId:         c.currentRoom.RoomId,
		}
		b, err := json.Marshal(event)
		if err != nil {
			WriteTimeout(time.Second*5, c.Conn, b)
		}
	}
	msgInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			SendMsg()
		}
	})
	if rmsv.IsGroup {
		rm.RoomType = "Group"
	} else {
		rm.RoomType = "Duo"
	}
	rm.RoomName = rmsv.RoomName
	return c
}
func main() {
	app := tview.NewApplication()

	chat := NewChat()
	mainpage := tview.NewPages()
	menubtn := tview.NewButton("menu").SetSelectedFunc(func() {
		mainpage.SwitchToPage("menuOptions")
	})

	chat.SetSelectedFunc(func(current *list.Element) {
		if chat.current != nil {
			rm, ok := chat.current.Value.(*Room)
			if ok {
				chat.RoomPages.SwitchToPage(rm.RoomIdString)
				mainpage.SwitchToPage("RoomPages")
			}
		}
	})
	menuOptionsPage := tview.NewForm().
		AddButton("Event logs", nil).
		AddButton("Find Users", nil).
		AddButton("Create Duo Room", nil).
		AddButton("Create Group Room", nil).
		AddButton("Unblock user", nil).
		AddButton("Block user", nil).
		AddButton("Change username", nil).
		AddButton("Room's actions", nil). // Change roomname, Add users to room, delete users from room, Show users.
		AddButton("Change Privacy for Duo Rooms", nil).
		AddButton("Change Privacy for Group Rooms", nil)

	FindUsersForm := tview.NewForm().
		AddInputField("First name", "", 20, nil, func(text string) {
			chat.CurrentUserSearch = text
		}).
		AddButton("Find", func() {
			event := User{
				Event:    "FindUsers",
				Username: chat.CurrentUserSearch,
			}
			b, err := json.Marshal(event)
			if err != nil {
				WriteTimeout(time.Second*5, chat.Conn, b)
			}
		}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRight {
			app.SetFocus(FindUsersResponse)
			return nil
		}
		return event
	})
	FindUsersResponse := tview.NewForm().AddButton("Create Group Room", func() {

	})
	mainpage.AddPage("RoomPages", chat.RoomPages, true, true)

	flexapp := tview.NewFlex().
		AddItem(chat, 0, 1, true).
		AddItem(mainpage, 0, 1, true).
		AddItem(menubtn, 0, 1, true)
	app.QueueUpdateDraw()
	if err := app.SetRoot(flexapp, true).Run(); err != nil {
		panic(err)
	}
}
