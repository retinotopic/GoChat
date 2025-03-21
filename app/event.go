package app

import (
	"strconv"

	"github.com/retinotopic/GoChat/app/list"
)

type SendEvent struct {
	Event    string `json:"Event"`
	ErrorMsg string `json:"ErrorMsg"`
	UserId   uint64 `json:"UserId"`
	Data     []byte `json:"-"`
}

type Room struct {
	Event    string   `json:"Event" `
	UserIds  []uint64 `json:"UserIds" `
	RoomIds  []uint64 `json:"RoomIds" `
	RoomName string   `json:"RoomName" `
	IsGroup  bool     `json:"IsGroup" `
}

// Room SendEvents
func (r *Room) AddDeleteUsersInRoom(args []list.Content, trg ...int) {
	// len(r.RoomIds) == 0 || len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		return
	}
	r.UserIds = r.UserIds[:0]
	r.RoomIds = r.RoomIds[:0]
	r.RoomIds = append(r.RoomIds, n)
	for i := range args {
		n, err := strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			return
		}
		r.UserIds = append(r.UserIds, n)
	}
	r.Event = SendEventNames[trg[0]] // "add users to room" or "delete users from room"

}

func (r *Room) BlockUnblockUser(args []list.Content, trg ...int) {
	// len(r.UserIds) == 0
	r.UserIds = r.UserIds[:0]
	n, err := strconv.ParseUint(args[1].MainText, 10, 64)
	if err != nil {
		return
	}
	r.UserIds = append(r.UserIds, n)
	r.Event = SendEventNames[trg[0]] // "block user" or "unblock user"
}

func (r *Room) CreateDuoRoom(args []list.Content, trg ...int) {
	// CreateDuoRoom
	// len(r.UserIds) == 0
	r.UserIds = r.UserIds[:0]
	n, err := strconv.ParseUint(args[1].MainText, 10, 64)
	if err != nil {
		return
	}
	r.UserIds = append(r.UserIds, n)
	r.Event = "Create Duo Room"
}

func (r *Room) CreateGroupRoom(args []list.Content, trg ...int) {
	// len(r.RoomName) == 0 || len(r.UserIds) == 0
	r.RoomName = args[len(args)-1].MainText
	r.UserIds = r.UserIds[:0]
	for i := range args {
		n, err := strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			return
		}
		r.UserIds = append(r.UserIds, n)
	}
	r.Event = "Create Group Room"
}

func (r *Room) ChangeRoomname(args []list.Content, trg ...int) {
	//  len(r.RoomIds) == 0 || len(r.RoomName) == 0
	//r.RoomIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.RoomName = args[len(args)-1].MainText
	r.Event = "Change Room Name"
}

type Message struct {
	Event          string `json:"Event" `
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
}

func (m *Message) SendMessage(args []list.Content, trg ...int) {
	// len(m.MessagePayload) == 0 || m.RoomId == 0
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-2].MainText, 10, 64)
	if err != nil {
		return
	}
	m.MessagePayload = args[len(args)-1].MainText
	m.Event = "Send Message"
}

func (m *Message) GetMessagesFromRoom(args []list.Content, trg ...int) {
	// m.RoomId == 0
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
}

type User struct {
	Event      string `json:"Event" `
	UserId     uint64 `json:"UserId"`
	Username   string `json:"Username" `
	RoomToggle bool   `json:"RoomToggle" `
}

// User SendEvents
func (u *User) ChangePrivacy(args []list.Content, trg ...int) {
	// ChangePrivacyDirect , ChangePrivacyGroup
	u.RoomToggle = args[0].MainText == "true"
	u.Event = SendEventNames[trg[0]] // "change privacy direct" or "change privacy group"
}

func (u *User) ChangeUsernameFindUsers(args []list.Content, trg ...int) {
	// len(u.Username) == 0
	u.Username = args[len(args)-1].MainText
	u.Event = SendEventNames[trg[0]] // "change username" or "find users"
}
