package app

import (
	"strconv"

	json "github.com/bytedance/sonic"
	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

type Room struct {
	SendCh   chan []byte `json:"-"`
	Event    string      `json:"Event" `
	UserIds  []uint64    `json:"UserIds" `
	RoomIds  []uint64    `json:"RoomIds" `
	RoomName string      `json:"RoomName" `
	IsGroup  bool        `json:"IsGroup" `
}

func (c *Chat) EventUI(cnt []list.Content, trg ...int) {
	lists := make([]tview.Primitive, 0, 5)
	c.Lists[trg[0]].Items.Clear()
	ll, ok := c.Lists[trg[0]].Items.(*list.ArrayList)
	if ok {
		for i := range cnt {
			a := ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
				cnt[i].MainText, cnt[i].SecondaryText)
			c.Lists[trg[0]].Items.MoveToBack(a)
		}
		for i := range trg {

			lists = append(lists, c.Lists[trg[i]])
			c.Logger.Println(c.Lists[trg[i]], "event ui")
		}
		c.AddItemMainFlex(lists...)
	}
}

// Room SendEvents
func (r Room) AddDeleteUsersInRoom(args []list.Content, trg ...int) {
	// should be: len(r.RoomIds) == 0 || len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		return
	}
	r.RoomIds = append(r.RoomIds, n)
	for i := range args {
		n, err = strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			return
		}
		r.UserIds = append(r.UserIds, n)
	}
	r.Event = SendEventNames[trg[0]] // "add users to room" or "delete users from room"
	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- data
	}
}

func (r Room) BlockUnblockUser(args []list.Content, trg ...int) {
	// should be: len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[1].MainText, 10, 64)
	if err != nil {
		return
	}
	r.UserIds = append(r.UserIds, n)
	r.Event = SendEventNames[trg[0]] // "block user" or "unblock user"
	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- data
	}
}

func (r Room) CreateDuoRoom(args []list.Content, trg ...int) {
	// should be: CreateDuoRoom
	// len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[1].MainText, 10, 64)
	if err != nil {
		return
	}
	r.UserIds = append(r.UserIds, n)
	r.Event = "Create Duo Room"
	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- data
	}
}

func (r Room) CreateGroupRoom(args []list.Content, trg ...int) {
	// should be: len(r.RoomName) == 0 || len(r.UserIds) == 0
	var err error
	r.RoomName = args[len(args)-1].MainText
	for i := range args {
		var n uint64
		n, err = strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			return
		}
		r.UserIds = append(r.UserIds, n)
	}
	r.Event = "Create Group Room"

	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- data
	}
}

func (r Room) ChangeRoomName(args []list.Content, trg ...int) {
	// should be:  len(r.RoomIds) == 0 || len(r.RoomName) == 0
	//r.RoomIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.RoomName = args[len(args)-1].MainText
	r.Event = "Change Room Name"
	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- data
	}
}

type Message struct {
	SendCh         chan []byte `json:"-"`
	Event          string      `json:"Event" `
	MessagePayload string      `json:"MessagePayload"`
	MessageId      uint64      `json:"MessageId" `
	RoomId         uint64      `json:"RoomId" `
	UserId         uint64      `json:"UserId" `
}

func (m Message) SendMessage(args []list.Content, trg ...int) {
	// should be: len(m.MessagePayload) == 0 || m.RoomId == 0
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		return
	}
	m.MessagePayload = args[len(args)-1].MainText
	m.Event = "Send Message"
	if data, err := json.Marshal(m); err == nil {
		m.SendCh <- data
	}
}

func (m Message) GetMessagesFromRoom(args []list.Content, trg ...int) {
	// should be: m.RoomId == 0
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		return
	}
	m.MessageId, err = strconv.ParseUint(args[0].MainText, 10, 64)
	if err != nil {
		return
	}
	m.Event = "Get Messages From Room"
	if data, err := json.Marshal(m); err == nil {
		m.SendCh <- data
	}
}

type User struct {
	SendCh     chan []byte `json:"-"`
	Event      string      `json:"Event" `
	UserId     uint64      `json:"UserId"`
	Username   string      `json:"Username" `
	RoomToggle bool        `json:"RoomToggle" `
}

// User SendEvents
func (u User) ChangePrivacy(args []list.Content, trg ...int) {
	// should be: ChangePrivacyDirect , ChangePrivacyGroup
	u.RoomToggle = args[0].MainText == "true"
	u.Event = SendEventNames[trg[0]] // "change privacy direct" or "change privacy group"
	if data, err := json.Marshal(u); err == nil {
		u.SendCh <- data
	}
}

func (u User) ChangeUsernameFindUsers(args []list.Content, trg ...int) {
	// should be: len(u.Username) == 0
	u.Username = args[len(args)-1].MainText
	u.Event = SendEventNames[trg[0]] // "change username" or "find users"
	if data, err := json.Marshal(u); err == nil {
		u.SendCh <- data
	}
}
func (u User) GetBlockedUsers(args []list.Content, trg ...int) {
	u.Event = "Get Blocked Users"
	if data, err := json.Marshal(u); err == nil {
		u.SendCh <- data
	}
}
