package models

import (
	json "github.com/bytedance/sonic"
)

type EventConstr interface {
	Message | RoomRequest | User
}

func UnmarshalEvent[T EventConstr](src []byte) (T, error) {
	var v T
	err := json.Unmarshal(src, &v)
	return v, err
}

type EventMetadata struct {
	Event     string                `json:"Event"`
	ErrorMsg  string                `json:"ErrorMsg"`
	UserId    uint64                `json:"UserId"`
	Type      int                   `json:"Type"`
	Data      json.NoCopyRawMessage `json:"Data"` // data to be sent over connection
	Kind      string                `json:"-"`    // subscribe or unsubscribe, "0" means unsubscribe, "1" means subscribe
	SubForPub []string              `json:"-"`    // a channels to publish to
	PubForSub []string              `json:"-"`    // publish in "user" channels for subscribe/unsubscribe only
	OrderCmd  [2]int                `json:"-"`    // value 1 means PublishWithMessage, value 2 means PublishWithSubscriptions, 0 means nothing
}

type RoomRequest struct {
	UserIds  []uint64 `json:"UserIds" `
	RoomIds  []uint64 `json:"RoomIds" `
	RoomName string   `json:"RoomName" `
	IsGroup  bool     `json:"IsGroup" `
}

type User struct {
	UserId     uint64 `json:"UserId"`
	Username   string `json:"Username" `
	RoomToggle bool   `json:"Bool" `
}

type Message struct {
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
}
