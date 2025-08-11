package app

import (
	"log"
	"strconv"

	json "github.com/bytedance/sonic"
	"github.com/retinotopic/GoChat/app/list"
)

type Room struct {
	UserIds  []uint64 `json:"UserIds" `
	RoomIds  []uint64 `json:"RoomIds" `
	RoomName string   `json:"RoomName" `
	IsGroup  bool     `json:"IsGroup" `
}

func (r Room) AddDeleteUsersInRoom(ev *EventInfo, args []list.Content) error {
	// should be len(r.RoomIds) == 0 || len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[len(args)-1].MainText, 10, 64)
	if err != nil {
		return err
	}
	r.RoomIds = append(r.RoomIds, n)
	for i := range len(args) - 1 {
		n, err = strconv.ParseUint(args[i].SecondaryText, 10, 64)
		if err != nil {
			log.Println(err)
		}
		r.UserIds = append(r.UserIds, n)
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	ev.Data = data
	return nil
}

func (r Room) BlockUnblockUser(ev *EventInfo, args []list.Content) error {
	// should be: len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[0].SecondaryText, 10, 64)
	if err != nil {
		return err
	}
	r.UserIds = append(r.UserIds, n)
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	ev.Data = data
	return nil
}

func (r Room) CreateDuoRoom(ev *EventInfo, args []list.Content) error {
	// should be: CreateDuoRoom
	// len(r.UserIds) == 0
	n, err := strconv.ParseUint(args[0].SecondaryText, 10, 64)
	if err != nil {
		return err
	}
	r.UserIds = append(r.UserIds, n)
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	ev.Data = data
	ev.Event = "Create Duo Room"
	return nil
}

func (r Room) CreateGroupRoom(ev *EventInfo, args []list.Content) error {
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
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	ev.Data = data
	ev.Event = "Create Group Room"
	return nil
}

func (r Room) ChangeRoomName(ev *EventInfo, args []list.Content) error {
	// should be:  len(r.RoomIds) == 0 || len(r.RoomName) == 0
	n, err := strconv.ParseUint(args[len(args)-1].MainText, 10, 64)
	if err != nil {
		log.Fatalln(err)
	}
	r.RoomIds = []uint64{n}
	r.RoomName = args[len(args)-1].SecondaryText
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	ev.Data = data
	ev.Event = "Change Room Name"
	return nil
}

type Message struct {
	MessagePayload string `json:"MessagePayload"`
	Username       string `json:"Username" `
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
	C              *Chat  `json:"-"`
}

func (m Message) SendMessage(ev *EventInfo, args []list.Content) error {
	// should be: len(m.MessagePayload) == 0 || m.RoomId == 0
	var err error
	m.RoomId, err = strconv.ParseUint(args[len(args)-1].MainText, 10, 64)
	if err != nil {
		return err
	}
	m.MessagePayload = args[len(args)-1].SecondaryText
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	ev.Data = data
	ev.Event = "Send Message"
	return nil
}

func (m Message) GetMessagesFromRoom(ev *EventInfo, args []list.Content) error {
	// should be: m.RoomId == 0
	var err error
	m.RoomId = m.C.CurrentRoom.RoomId
	m.MessageId = m.C.CurrentRoom.LastMessageID
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	ev.Data = data
	ev.Event = "Get Messages From Room"
	return nil
}

type User struct {
	UserId     uint64 `json:"UserId"`
	Username   string `json:"Username" `
	RoomToggle bool   `json:"RoomToggle" `
}

// User SendEvents
func (u User) ChangePrivacy(ev *EventInfo, args []list.Content) error {
	// should be: ChangePrivacyDirect , ChangePrivacyGroup
	u.RoomToggle = args[0].MainText == "true"
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	ev.Data = data
	return nil
}

func (u User) ChangeUsernameFindUsers(ev *EventInfo, args []list.Content) error {
	// should be: len(u.Username) == 0
	u.Username = args[len(args)-1].SecondaryText
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	ev.Data = data
	return nil
}
func (u User) GetBlockedUsers(ev *EventInfo, args []list.Content) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	ev.Data = data
	ev.Event = "Get Blocked Users"
	return nil
}
