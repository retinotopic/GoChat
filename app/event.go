package main

import "strconv"

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
func (r Room) AddDeleteUsersInRoom(args []string, trg ...int) {

	// len(r.RoomIds) == 0 || len(r.UserIds) == 0
	//r.RoomIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.UserIds = make([]uint64, len(args)-2)
	for i, id := range args[2:] {
		r.UserIds[i], _ = strconv.ParseUint(id, 10, 64)
	}
	r.Event = SendEventNames[trg[0]] // "add_users_to_room" or "delete_users_from_room"
}

func (r Room) BlockUnblockUser(args []Content, trg ...int) {
	// len(r.UserIds) == 0
	//r.UserIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.Event = SendEventNames[trg[0]] // "block_user" or "unblock_user"
}

func (r Room) CreateDuoRoom(args []Content, trg ...int) {
	// CreateDuoRoom
	// len(r.UserIds) == 0
	//r.UserIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.Event = "create_duo_room"
}

func (r Room) CreateGroupRoom(args []Content, trg ...int) {
	// len(r.RoomName) == 0 || len(r.UserIds) == 0
	r.RoomName = args[1].Text
	r.UserIds = make([]uint64, len(args)-2)
	for i, _ := range args[2:] {
		r.UserIds[i], _ = strconv.ParseUint(id, 10, 64)
	}
	r.Event = "create_group_room"
}

func (r Room) ChangeRoomname(args []Content, trg ...int) {
	//  len(r.RoomIds) == 0 || len(r.RoomName) == 0
	//r.RoomIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.RoomName = args[2].Text
	r.Event = "Change Roomname"
}

type Message struct {
	Event          string `json:"Event" `
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
}

func (m Message) SendMessage(args []Content, trg ...int) {
	// len(m.MessagePayload) == 0 || m.RoomId == 0
	m.RoomId, _ = strconv.ParseUint(args[1].Text, 10, 64)
	m.MessagePayload = args[2].Text
	m.Event = "Send Message"
}

func (m Message) GetMessagesFromRoom(args []Content, trg ...int) {
	// m.RoomId == 0
	roomid, err := strconv.ParseUint(args[1].Text, 10, 64)

	if err != nil {
		return
	}
	m.RoomId = roomid
	if len(args) > 2 {
		m.MessageId, _ = strconv.ParseUint(args[2].Text, 10, 64)
	}
	m.Event = "Get Messages From Room"
}

type User struct {
	Event      string `json:"Event" `
	UserId     uint64 `json:"UserId"`
	Username   string `json:"Username" `
	RoomToggle bool   `json:"Bool" `
}

// User SendEvents
func (u User) ChangePrivacy(args []Content, trg ...int) {
	// ChangePrivacyDirect , ChangePrivacyGroup
	u.RoomToggle = args[1].Text == "true"
	u.Event = SendEventNames[trg[0]] // "change_privacy_direct" or "change_privacy_group"
}

func (u User) ChangeUsernameFindUsers(args []Content, trg ...int) {
	// len(u.Username) == 0
	u.Username = args[1].Text
	u.Event = SendEventNames[trg[0]] // "change_username" or "find_users"
}
