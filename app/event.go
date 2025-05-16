package app

import (
	"log"
	"strconv"

	json "github.com/bytedance/sonic"
	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

type Room struct {
	SendCh   chan EventInfo `json:"-"`
	UserIds  []uint64       `json:"UserIds" `
	RoomIds  []uint64       `json:"RoomIds" `
	RoomName string         `json:"RoomName" `
	IsGroup  bool           `json:"IsGroup" `
}

func (c *Chat) EventUI(cnt []list.Content, trg ...int) {
	lists := make([]tview.Primitive, 0, 5)
	if len(cnt) > 0 {
		c.Lists[trg[0]].Items.Clear()
	}
	ll, ok := c.Lists[trg[0]].Items.(*list.ArrayList)
	if ok {
		c.Logger.Println("Event", "Correct Branch")
		for i := range cnt {
			a := ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
				cnt[i].MainText, cnt[i].SecondaryText)
			c.Lists[trg[0]].Items.MoveToBack(a)
		}
		for i := range trg {
			lists = append(lists, c.Lists[trg[i]])
			c.Logger.Println("Event", trg[i], "event ui list id")
		}

		c.UserBuf = c.UserBuf[:0]
		for _, v := range c.DuoUsers {
			c.UserBuf = append(c.UserBuf, v)

		}
		c.FillUsers(c.UserBuf, 6)

		c.AddItemMainFlex(lists...)
	}
}

// Room SendEvents
func (r Room) AddDeleteUsersInRoom(args []list.Content, trg ...int) {
	// should be: len(r.RoomIds) == 0 || len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[len(args)-1].MainText, 10, 64)
	if err != nil {
		return
	}
	r.RoomIds = append(r.RoomIds, n)
	for i := range len(args) - 1 {
		n, err = strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		r.UserIds = append(r.UserIds, n)
	}
	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- EventInfo{Data: data, Type: 2, Event: SendEventNames[trg[0]]} // "add users to room" or "delete users from room"
	}
}

func (r Room) BlockUnblockUser(args []list.Content, trg ...int) {
	// should be: len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[0].SecondaryText, 10, 64)
	if err != nil {
		return
	}
	r.UserIds = append(r.UserIds, n)
	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- EventInfo{Data: data, Type: 2, Event: SendEventNames[trg[0]]} // "block user" or "unblock user"
	}
}

func (r Room) CreateDuoRoom(args []list.Content, trg ...int) {
	// should be: CreateDuoRoom
	// len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[0].SecondaryText, 10, 64)
	if err != nil {
		return
	}
	r.UserIds = append(r.UserIds, n)
	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- EventInfo{Data: data, Type: 2, Event: "Create Duo Room"}
	}
}

func (r Room) CreateGroupRoom(args []list.Content, trg ...int) {
	// should be: len(r.RoomName) == 0 || len(r.UserIds) == 0
	var err error
	r.RoomName = args[len(args)-1].SecondaryText
	for i := range len(args) - 1 {
		var n uint64
		n, err = strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		r.UserIds = append(r.UserIds, n)
	}
	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- EventInfo{Data: data, Type: 2, Event: "Create Group Room"}
	}
}

func (r Room) ChangeRoomName(args []list.Content, trg ...int) {
	// should be:  len(r.RoomIds) == 0 || len(r.RoomName) == 0
	n, err := strconv.ParseUint(args[len(args)-1].MainText, 10, 64)
	if err != nil {
		log.Fatalln(err)
	}
	r.RoomIds = []uint64{n}
	r.RoomName = args[len(args)-1].SecondaryText
	if data, err := json.Marshal(r); err == nil {
		r.SendCh <- EventInfo{Data: data, Type: 2, Event: "Change Room Name"}
	}
}

type Message struct {
	SendCh         chan EventInfo `json:"-"`
	MessagePayload string         `json:"MessagePayload"`
	Username       string         `json:"Username" `
	MessageId      uint64         `json:"MessageId" `
	RoomId         uint64         `json:"RoomId" `
	UserId         uint64         `json:"UserId" `
}

func (m Message) SendMessage(args []list.Content, trg ...int) {
	// should be: len(m.MessagePayload) == 0 || m.RoomId == 0
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-1].MainText, 10, 64)
	if err != nil {
		return
	}
	m.MessagePayload = args[len(args)-1].SecondaryText
	if data, err := json.Marshal(m); err == nil {
		m.SendCh <- EventInfo{Data: data, Type: 1, Event: "Send Message"}
	}
}

func (m Message) GetMessagesFromRoom(args []list.Content, trg ...int) {
	// should be: m.RoomId == 0
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-1].MainText, 10, 64)
	if err != nil {
		return
	}
	m.MessageId, err = strconv.ParseUint(args[0].MainText, 10, 64) //lastmessageid in room
	if err != nil {
		return
	}
	if data, err := json.Marshal(m); err == nil {
		m.SendCh <- EventInfo{Data: data, Type: 1, Event: "Get Messages From Room"}
	}
}

type User struct {
	SendCh     chan EventInfo `json:"-"`
	UserId     uint64         `json:"UserId"`
	Username   string         `json:"Username" `
	RoomToggle bool           `json:"RoomToggle" `
}

// User SendEvents
func (u User) ChangePrivacy(args []list.Content, trg ...int) {
	// should be: ChangePrivacyDirect , ChangePrivacyGroup
	u.RoomToggle = args[0].MainText == "true"
	if data, err := json.Marshal(u); err == nil {
		u.SendCh <- EventInfo{Data: data, Type: 3, Event: SendEventNames[trg[0]]} // "change privacy direct" or "change privacy group"
	}
}

func (u User) ChangeUsernameFindUsers(args []list.Content, trg ...int) {
	// should be: len(u.Username) == 0
	u.Username = args[len(args)-1].SecondaryText
	if data, err := json.Marshal(u); err == nil {
		u.SendCh <- EventInfo{Data: data, Type: 3, Event: SendEventNames[trg[0]]} // "change username" or "find users"
	}
}
func (u User) GetBlockedUsers(args []list.Content, trg ...int) {
	if data, err := json.Marshal(u); err == nil {
		u.SendCh <- EventInfo{Data: data, Type: 3, Event: "Get Blocked Users"}
	}
}
