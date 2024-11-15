package main

import (
	"container/list"
	"strconv"
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
	RoomIdString    string
	Messages        *tview.List
	RoomLL          *list.Element // node in Rooms linked list (type Room)
}

// Rooms linked list
type Chat struct {
	MyId      uint64
	Conn      *websocket.Conn
	RoomMsgs  map[uint64]*Room // room id to room message page
	DuoRooms  map[uint64]*User // duo rooms that
	RoomPages *tview.Pages
	*tview.Box
	items        *list.List
	current      *list.Element // type Room
	currentRoom  *Room         // type Room
	offset       int           // scroll offset
	selectedFunc func(*list.Element)

	CurrentUserSearch string
}

func NewChat() *Chat {
	return &Chat{
		Box:   tview.NewBox(),
		items: list.New(),
	}
}

func (r *Chat) ProcessRoom(rmcl RoomServer) *Chat {

	rm, ok := r.RoomMsgs[rmcl.RoomId]
	if ok {
		for i := range len(rmcl.Users) {
			if rmcl.Users[i].UserId == r.MyId {
				r.DeleteRoom(rmcl.RoomId)
				break
			}
			if rmcl.Users[i].RoomToggle {
				delete(rm.Users, rmcl.Users[i].UserId)
			} else {
				rm.Users[rmcl.Users[i].UserId] = rmcl.Users[i]
			}
		}
	} else {
		r.AddRoom(rmcl)
	}

	return r
}
func (r *Chat) MoveToFront(e *list.Element) *Chat {
	r.items.MoveToFront(e)
	return r
}
func (r *Chat) DeleteRoom(roomid uint64) *Chat {
	rm, ok := r.RoomMsgs[roomid]
	if ok {
		if rm.RoomLL != nil {
			r.items.Remove(rm.RoomLL) // deleting node in Rooms linked list
		}
		rm.Messages.Clear()                     // deleting messages from room *tview.List
		r.RoomPages.RemovePage(rm.RoomIdString) // deleting corresponding page
		delete(r.RoomMsgs, roomid)              // deleting *Room instance in map
	}
	return r
}
func (r *Chat) AddRoom(rmcl RoomServer) *Chat {
	//fill room with users
	r.RoomMsgs[rmcl.RoomId] = &Room{Users: make(map[uint64]User)}
	rm := r.RoomMsgs[rmcl.RoomId]
	for i := range len(rmcl.Users) {
		rm.Users[rmcl.Users[i].UserId] = rmcl.Users[i]
	}
	// set new room at the top
	item := r.items.PushFront(rm)
	rm.RoomLL = item

	// creating corresponding page for this room
	msgInput := tview.NewInputField() // creating input for message
	SendMsg := func() {               // func for sending message
		msg := msgInput.GetText()
		event := Message{
			Event:          "SendMessage",
			MessagePayload: msg,
			RoomId:         r.currentRoom.RoomId,
		}
		b, err := json.Marshal(event)
		if err != nil {
			WriteTimeout(time.Second*5, r.Conn, b)
		}
	}
	msgInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			SendMsg()
		}
	})
	rm.Messages = tview.NewList()
	flexroom := tview.NewFlex().
		AddItem(rm.Messages, 0, 1, false).
		AddItem(msgInput, 0, 1, false).
		AddItem(tview.NewButton("sendmessage").SetSelectedFunc(func() {
			SendMsg()
		}), 0, 1, false)
	if rmcl.IsGroup {
		rm.RoomType = "Group"
	} else {
		rm.RoomType = "Duo"
	}
	rm.RoomName = rmcl.RoomName
	rm.RoomIdString = strconv.FormatUint(rmcl.RoomId, 10)
	r.RoomPages.AddPage(rm.RoomIdString, flexroom, false, false)
	return r
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
		AddButton("Show Blocked Users", nil).
		AddButton("Change username", nil).
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

	if err := app.SetRoot(flexapp, true).Run(); err != nil {
		panic(err)
	}
}
