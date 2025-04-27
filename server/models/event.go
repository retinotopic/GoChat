package models

import (
	"bytes"
)

type EventMetadata struct {
	Event     string   `json:"Event"`
	ErrorMsg  string   `json:"ErrorMsg"`
	UserId    uint64   `json:"UserId"`
	Data      []byte   `json:"Data"` // data to be sent over connection
	Kind      string   `json:"-"`    // subscribe or unsubscribe, "0" means unsubscribe, "1" means subscribe
	SubForPub []string `json:"-"`    // a channels to publish to
	PubForSub []string `json:"-"`    // publish in "user" channels for subscribe/unsubscribe only
	OrderCmd  [2]int   `json:"-"`    // value 1 means PublishWithMessage, value 2 means PublishWithSubscriptions, 0 means nothing
}

func (e *EventMetadata) GetEventName() {
	start := bytes.IndexByte(e.Data, '"')
	if start == -1 {
		e.Event = ""
	}
	start++
	end := bytes.IndexByte(e.Data[start:], '"')
	if end == -1 {
		e.Event = ""
	}
	e.Event = string(e.Data[start : start+end])
}
