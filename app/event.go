package app

import (
	"encoding/json"
	"strconv"

	"github.com/retinotopic/GoChat/app/list"
)

type SendEvent struct {
	Event    string `json:"Event"`
	ErrorMsg string `json:"ErrorMsg"`
	UserId   uint64 `json:"UserId"`
	Data     []byte `json:"-"`
}

func (s SendEvent) Copy() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return b, err
	}
	return b, err
}

type Room struct {
	ToSend   *SendEvent
	Event    string   `json:"Event" `
	UserIds  []uint64 `json:"UserIds" `
	RoomIds  []uint64 `json:"RoomIds" `
	RoomName string   `json:"RoomName" `
	IsGroup  bool     `json:"IsGroup" `
}

// Room SendEvents
func (r *Room) AddDeleteUsersInRoom(args []list.Content, trg ...int) {
	// should be: len(r.RoomIds) == 0 || len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		r.ToSend.ErrorMsg = err.Error()
		return
	}
	r.UserIds = r.UserIds[:0]
	r.RoomIds = r.RoomIds[:0]
	r.RoomIds = append(r.RoomIds, n)
	for i := range args {
		n, err := strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			r.ToSend.ErrorMsg = err.Error()
			return
		}
		r.UserIds = append(r.UserIds, n)
	}
	r.Event = SendEventNames[trg[0]] // "add users to room" or "delete users from room"
	if r.ToSend.Data, err = json.Marshal(r); err != nil {
		r.ToSend.ErrorMsg = err.Error()
	}
}

func (r *Room) BlockUnblockUser(args []list.Content, trg ...int) {
	// should be: len(r.UserIds) == 0
	r.UserIds = r.UserIds[:0]
	n, err := strconv.ParseUint(args[1].MainText, 10, 64)
	if err != nil {
		r.ToSend.ErrorMsg = err.Error()
		return
	}
	r.UserIds = append(r.UserIds, n)
	r.Event = SendEventNames[trg[0]] // "block user" or "unblock user"
	if r.ToSend.Data, err = json.Marshal(r); err != nil {
		r.ToSend.ErrorMsg = err.Error()
	}
}

func (r *Room) CreateDuoRoom(args []list.Content, trg ...int) {
	// should be: CreateDuoRoom
	// len(r.UserIds) == 0
	r.UserIds = r.UserIds[:0]
	n, err := strconv.ParseUint(args[1].MainText, 10, 64)
	if err != nil {
		r.ToSend.ErrorMsg = err.Error()
		return
	}
	r.UserIds = append(r.UserIds, n)
	r.Event = "Create Duo Room"
	if r.ToSend.Data, err = json.Marshal(r); err != nil {
		r.ToSend.ErrorMsg = err.Error()
	}
}

func (r *Room) CreateGroupRoom(args []list.Content, trg ...int) {
	// should be: len(r.RoomName) == 0 || len(r.UserIds) == 0
	r.RoomName = args[len(args)-1].MainText
	r.UserIds = r.UserIds[:0]
	for i := range args {
		n, err := strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			r.ToSend.ErrorMsg = err.Error()
			return
		}
		r.UserIds = append(r.UserIds, n)
	}
	r.Event = "Create Group Room"
	var err error
	if r.ToSend.Data, err = json.Marshal(r); err != nil {
		r.ToSend.ErrorMsg = err.Error()
	}
}

func (r *Room) ChangeRoomName(args []list.Content, trg ...int) {
	// should be:  len(r.RoomIds) == 0 || len(r.RoomName) == 0
	//r.RoomIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.RoomName = args[len(args)-1].MainText
	r.Event = "Change Room Name"
}

type Message struct {
	ToSend         *SendEvent
	Event          string `json:"Event" `
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
}

func (m *Message) SendMessage(args []list.Content, trg ...int) {
	// should be: len(m.MessagePayload) == 0 || m.RoomId == 0
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		m.ToSend.ErrorMsg = err.Error()
		return
	}
	m.MessagePayload = args[len(args)-1].MainText
	m.Event = "Send Message"
	if m.ToSend.Data, err = json.Marshal(m); err != nil {
		m.ToSend.ErrorMsg = err.Error()
	}
}

func (m *Message) GetMessagesFromRoom(args []list.Content, trg ...int) {
	// should be: m.RoomId == 0
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		m.ToSend.ErrorMsg = err.Error()
		return
	}
	m.MessageId, err = strconv.ParseUint(args[0].MainText, 10, 64)
	if err != nil {
		m.ToSend.ErrorMsg = err.Error()
		return
	}
	m.Event = "Get Messages From Room"
	if m.ToSend.Data, err = json.Marshal(m); err != nil {
		m.ToSend.ErrorMsg = err.Error()
	}
}

type User struct {
	ToSend     *SendEvent
	Event      string `json:"Event" `
	UserId     uint64 `json:"UserId"`
	Username   string `json:"Username" `
	RoomToggle bool   `json:"RoomToggle" `
}

// User SendEvents
func (u *User) ChangePrivacy(args []list.Content, trg ...int) {
	// should be: ChangePrivacyDirect , ChangePrivacyGroup
	u.RoomToggle = args[0].MainText == "true"
	u.Event = SendEventNames[trg[0]] // "change privacy direct" or "change privacy group"
	var err error
	if u.ToSend.Data, err = json.Marshal(u); err != nil {
		u.ToSend.ErrorMsg = err.Error()
	}
}

func (u *User) ChangeUsernameFindUsers(args []list.Content, trg ...int) {
	// should be: len(u.Username) == 0
	u.Username = args[len(args)-1].MainText
	u.Event = SendEventNames[trg[0]] // "change username" or "find users"
	var err error
	if u.ToSend.Data, err = json.Marshal(u); err != nil {
		u.ToSend.ErrorMsg = err.Error()
	}
}
func (u *User) GetBlockedUsers(args []list.Content, trg ...int) {
	u.Event = "Get Blocked Users"
	var err error
	if u.ToSend.Data, err = json.Marshal(u); err != nil {
		u.ToSend.ErrorMsg = err.Error()
	}
}
