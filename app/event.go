package app

import (
	"log"
	"strconv"

	json "github.com/bytedance/sonic"
	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

type SendEvent struct {
	Event    string `json:"Event"`
	ErrorMsg string `json:"ErrorMsg"`
	UserId   uint64 `json:"UserId"`
	Data     []byte `json:"-"`
}

type Room struct {
	ToSend   chan SendEvent
	Event    string   `json:"Event" `
	UserIds  []uint64 `json:"UserIds" `
	RoomIds  []uint64 `json:"RoomIds" `
	RoomName string   `json:"RoomName" `
	IsGroup  bool     `json:"IsGroup" `
}

func (c *Chat) EventUI(cnt []list.Content, trg ...int) {
	lists := make([]tview.Primitive, 0, 5)
	log.Println(len(cnt), trg)
	c.Lists[trg[0]].Items.Clear()
	ll, ok := c.Lists[trg[0]].Items.(*list.ArrayList)
	log.Println(ok, "DFIJMIODFJOISDJGFIOSDOIFJOIS")
	if ok {
		for i := range cnt {
			a := list.ArrayItem{}
			a.ArrList = ll
			a.Color = [2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}

			a.MainText = cnt[i].MainText

			a.SecondaryText = cnt[i].SecondaryText

			c.Lists[trg[0]].Items.MoveToBack(a)
		}
	}
	for i := range trg {
		lists = append(lists, c.Lists[trg[i]])
	}
	c.AddItemMainFlex(lists...)
}

// Room SendEvents
func (r Room) AddDeleteUsersInRoom(args []list.Content, trg ...int) {
	// should be: len(r.RoomIds) == 0 || len(r.UserIds) == 0
	e := SendEvent{}
	n, err := strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		e.ErrorMsg = err.Error()
		return
	}
	r.RoomIds = append(r.RoomIds, n)
	for i := range args {
		n, err = strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			e.ErrorMsg = err.Error()
			return
		}
		r.UserIds = append(r.UserIds, n)
	}
	r.Event = SendEventNames[trg[0]] // "add users to room" or "delete users from room"
	if e.Data, err = json.Marshal(r); err != nil {
		e.ErrorMsg = err.Error()
	}
	r.ToSend <- e
}

func (r Room) BlockUnblockUser(args []list.Content, trg ...int) {
	// should be: len(r.UserIds) == 0
	e := SendEvent{}
	n, err := strconv.ParseUint(args[1].MainText, 10, 64)
	if err != nil {
		e.ErrorMsg = err.Error()
		return
	}
	r.UserIds = append(r.UserIds, n)
	r.Event = SendEventNames[trg[0]] // "block user" or "unblock user"
	if e.Data, err = json.Marshal(r); err != nil {
		e.ErrorMsg = err.Error()
	}
	r.ToSend <- e
}

func (r Room) CreateDuoRoom(args []list.Content, trg ...int) {
	// should be: CreateDuoRoom
	// len(r.UserIds) == 0
	e := SendEvent{}
	n, err := strconv.ParseUint(args[1].MainText, 10, 64)
	if err != nil {
		e.ErrorMsg = err.Error()
		return
	}
	r.UserIds = append(r.UserIds, n)
	r.Event = "Create Duo Room"
	if e.Data, err = json.Marshal(r); err != nil {
		e.ErrorMsg = err.Error()
	}
	r.ToSend <- e
}

func (r Room) CreateGroupRoom(args []list.Content, trg ...int) {
	// should be: len(r.RoomName) == 0 || len(r.UserIds) == 0
	e := SendEvent{}
	var err error
	r.RoomName = args[len(args)-1].MainText
	for i := range args {
		var n uint64
		n, err = strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			e.ErrorMsg = err.Error()
			return
		}
		r.UserIds = append(r.UserIds, n)
	}
	r.Event = "Create Group Room"

	if e.Data, err = json.Marshal(r); err != nil {
		e.ErrorMsg = err.Error()
	}
	r.ToSend <- e
}

func (r Room) ChangeRoomName(args []list.Content, trg ...int) {
	// should be:  len(r.RoomIds) == 0 || len(r.RoomName) == 0
	e := SendEvent{}
	var err error
	//r.RoomIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.RoomName = args[len(args)-1].MainText
	r.Event = "Change Room Name"
	if e.Data, err = json.Marshal(r); err != nil {
		e.ErrorMsg = err.Error()
	}
	r.ToSend <- e
}

type Message struct {
	ToSend         chan SendEvent
	Event          string `json:"Event" `
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
}

func (m Message) SendMessage(args []list.Content, trg ...int) {
	// should be: len(m.MessagePayload) == 0 || m.RoomId == 0
	e := SendEvent{}
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		e.ErrorMsg = err.Error()
		return
	}
	m.MessagePayload = args[len(args)-1].MainText
	m.Event = "Send Message"
	if e.Data, err = json.Marshal(m); err != nil {
		e.ErrorMsg = err.Error()
	}
	m.ToSend <- e
}

func (m Message) GetMessagesFromRoom(args []list.Content, trg ...int) {
	// should be: m.RoomId == 0
	e := SendEvent{}
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		e.ErrorMsg = err.Error()
		return
	}
	m.MessageId, err = strconv.ParseUint(args[0].MainText, 10, 64)
	if err != nil {
		e.ErrorMsg = err.Error()
		return
	}
	m.Event = "Get Messages From Room"
	if e.Data, err = json.Marshal(m); err != nil {
		e.ErrorMsg = err.Error()
	}
	m.ToSend <- e
}

type User struct {
	ToSend     chan SendEvent
	Event      string `json:"Event" `
	UserId     uint64 `json:"UserId"`
	Username   string `json:"Username" `
	RoomToggle bool   `json:"RoomToggle" `
}

// User SendEvents
func (u User) ChangePrivacy(args []list.Content, trg ...int) {
	// should be: ChangePrivacyDirect , ChangePrivacyGroup
	var err error
	e := SendEvent{}
	u.RoomToggle = args[0].MainText == "true"
	u.Event = SendEventNames[trg[0]] // "change privacy direct" or "change privacy group"
	if e.Data, err = json.Marshal(u); err != nil {
		e.ErrorMsg = err.Error()
	}
	u.ToSend <- e
}

func (u User) ChangeUsernameFindUsers(args []list.Content, trg ...int) {
	// should be: len(u.Username) == 0
	e := SendEvent{}
	u.Username = args[len(args)-1].MainText
	u.Event = SendEventNames[trg[0]] // "change username" or "find users"
	var err error
	if e.Data, err = json.Marshal(u); err != nil {
		e.ErrorMsg = err.Error()
	}
	u.ToSend <- e
}
func (u User) GetBlockedUsers(args []list.Content, trg ...int) {
	u.Event = "Get Blocked Users"
	e := SendEvent{}
	var err error
	if e.Data, err = json.Marshal(u); err != nil {
		e.ErrorMsg = err.Error()
	}
	u.ToSend <- e
}
