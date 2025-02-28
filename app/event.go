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
func (r Room) AddDeleteUsersInRoom(args []string, ...int) {
	// len(r.RoomIds) == 0 || len(r.UserIds) == 0
	//r.RoomIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.UserIds = make([]uint64, len(args)-2)
	for i, id := range args[2:] {
		r.UserIds[i], _ = strconv.ParseUint(id, 10, 64)
	}
	r.Event = args[0] // "add_users_to_room" or "delete_users_from_room"
}

func (r Room) BlockUnblockUser(args []Content,...int) {
	// len(r.UserIds) == 0
	//r.UserIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.Event = args[0] // "block_user" or "unblock_user"
}

func (r Room) CreateDuoRoom(args []Content,...int) {
	// CreateDuoRoom
	// len(r.UserIds) == 0
	//r.UserIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.Event = "create_duo_room"
}

func (r Room) CreateGroupRoom(args []Content,...int) {
	// len(r.RoomName) == 0 || len(r.UserIds) == 0
	r.RoomName = args[1]
	r.UserIds = make([]uint64, len(args)-2)
	for i, id := range args[2:] {
		r.UserIds[i], _ = strconv.ParseUint(id, 10, 64)
	}
	r.Event = "create_group_room"
}

func (r Room) ChangeRoomname(args []Content,...int) {
	//  len(r.RoomIds) == 0 || len(r.RoomName) == 0
	//r.RoomIds = []uint64{strconv.ParseUint(args[1], 10, 64)}
	r.RoomName = args[2]
	r.Event = "change_roomname"
}

type Message struct {
	Event          string `json:"Event" `
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
}

func (m Message) SendMessage(args []Content,...int) {
	// len(m.MessagePayload) == 0 || m.RoomId == 0
	m.RoomId, _ = strconv.ParseUint(args[1], 10, 64)
	m.MessagePayload = args[2]
	m.Event = "send_message"
}

func (m Message) GetMessagesFromRoom(args []Content,...int) {
	// m.RoomId == 0
	m.RoomId, _ = strconv.ParseUint(args[1], 10, 64)
	if len(args) > 2 {
		m.MessageId, _ = strconv.ParseUint(args[2], 10, 64)
	}
	m.Event = "get_messages"
}

type User struct {
	Event      string `json:"Event" `
	UserId     uint64 `json:"UserId"`
	Username   string `json:"Username" `
	RoomToggle bool   `json:"Bool" `
}

// User SendEvents
func (u User) ChangePrivacy(args []Content,...int) {
	// ChangePrivacyDirect , ChangePrivacyGroup
	u.RoomToggle = args[1] == "true"
	u.Event = args[0] // "change_privacy_direct" or "change_privacy_group"
}

func (u User) ChangeUsernameFindUsers(args []Content,...int) {
	// len(u.Username) == 0
	u.Username = args[1]
	u.Event = args[0] // "change_username" or "find_users"
}
