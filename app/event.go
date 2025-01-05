package main

var EventMap map[string]EventButton

type EventButton struct {
	ListIdx   int
	SendEvent func([]string)
}

type Event struct {
	Event    string `json:"Event"`
	ErrorMsg string `json:"ErrorMsg"`
	UserId   uint64 `json:"UserId"`
	Data     []byte `json:"-"`
}

type RoomRequest struct {
	Event    string   `json:"Event" `
	UserIds  []uint64 `json:"UserIds" `
	RoomIds  []uint64 `json:"RoomIds" `
	RoomName string   `json:"RoomName" `
	IsGroup  bool     `json:"IsGroup" `
}

func (r RoomRequest) SendEvent() {

}

type Message struct {
	Event          string `json:"Event" `
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
}

func (m Message) SendEvent() {

}

type User struct {
	Event      string `json:"Event" `
	UserId     uint64 `json:"UserId"`
	Username   string `json:"Username" `
	RoomToggle bool   `json:"Bool" `
}

func (u User) SendEvent() {

}
