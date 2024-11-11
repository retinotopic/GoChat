package main

import (
	"container/list"
	"strconv"

	"github.com/rivo/tview"
)

type RoomRequest struct {
	Event    string   `json:"Event" `
	UserIds  []uint64 `json:"UserIds" `
	RoomIds  []uint64 `json:"RoomIds" `
	RoomName string   `json:"RoomName" `
	IsGroup  bool     `json:"IsGroup" `
}

type Message struct {
	Event          string `json:"Event" `
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
}
type User struct {
	Event    string `json:"Event" `
	UserId   uint64 `json:"UserId"`
	Username string `json:"Username" `
	Bool     bool   `json:"Bool" `
}

type RoomClient struct {
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
	Users           map[uint64]User
	RoomType        string // Group or Duo
	RoomIdString    string
	Messages        *tview.List
	RoomLL          *list.Element // node in Rooms linked list (type Room)
}

// Rooms linked list
type Rooms struct {
	MyId      uint64
	RoomMsgs  map[uint64]*Room // room id to room message page
	DuoRooms  map[uint64]*User // duo rooms that
	RoomPages *tview.Pages
	*tview.Box
	items        *list.List
	current      *list.Element // type Room
	offset       int           // scroll offset
	selectedFunc func(*list.Element)
}

func NewRooms() *Rooms {
	return &Rooms{
		Box:   tview.NewBox(),
		items: list.New(),
	}
}

func (r *Rooms) ProcessRoom(rmcl RoomClient) *Rooms {

	rm, ok := r.RoomMsgs[rmcl.RoomId]
	if ok {
		for i := range len(rmcl.Users) {
			if rmcl.Users[i].UserId == r.MyId {
				r.DeleteRoom(rmcl.RoomId)
			}
			if rmcl.Users[i].Bool {
				delete(rm.Users, rmcl.Users[i].UserId)
			}
			rm.Users[rmcl.Users[i].UserId] = rmcl.Users[i]
		}
	} else {
		r.RoomMsgs[rmcl.RoomId] = &Room{Users: make(map[uint64]User)}
		rm = r.RoomMsgs[rmcl.RoomId]
		for i := range len(rmcl.Users) {
			rm.Users[rmcl.Users[i].UserId] = rmcl.Users[i]
		}
		item := r.items.PushFront(rm)
		rm.RoomLL = item
		rm.Messages = tview.NewList()
		flexroom := tview.NewFlex().
			AddItem(rm.Messages, 0, 1, false).
			AddItem(tview.NewButton("sendmessage").SetSelectedFunc(func() {

			}), 0, 1, false)
		if rmcl.IsGroup {
			rm.RoomType = "Group"
		} else {
			rm.RoomType = "Duo"
		}
		rm.RoomName = rmcl.RoomName
		rm.RoomIdString = strconv.FormatUint(rmcl.RoomId, 10)
		r.RoomPages.AddPage(rm.RoomIdString, flexroom, false, false)
	}

	return r
}
func (r *Rooms) MoveToFront(e *list.Element) *Rooms {
	r.items.MoveToFront(e)
	return r
}
func (r *Rooms) DeleteRoom(roomid uint64) *Rooms {
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

func main() {
	app := tview.NewApplication()

	rooms := NewRooms()
	mainpage := tview.NewPages()
	menubtn := tview.NewButton("menu").SetSelectedFunc(func() {
		mainpage.SwitchToPage("menuOptions")
	})

	rooms.SetSelectedFunc(func(current *list.Element) {
		if rooms.current != nil {
			rm, ok := rooms.current.Value.(*Room)
			if ok {
				rooms.RoomPages.SwitchToPage(rm.RoomIdString)
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

	mainpage.AddPage("menuOptions", menuOptionsPage, true, true)
	mainpage.AddPage("RoomPages", rooms.RoomPages, true, true)

	flexapp := tview.NewFlex().
		AddItem(rooms, 0, 1, true).
		AddItem(mainpage, 0, 1, true).
		AddItem(menubtn, 0, 1, true)

	if err := app.SetRoot(flexapp, true).Run(); err != nil {
		panic(err)
	}
}
