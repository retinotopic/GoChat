package models

import "bytes"

type Event struct {
	Event     string   `json:"Event"`
	ErrorMsg  string   `json:"ErrorMsg"`
	UserId    uint32   `json:"UserId"`
	Kind      string   `json:"-"` // subscribe or unsubscribe, "0" means unsubscribe, "1" means subscribe
	SubForPub []string `json:"-"` // a channels to publish to
	PubForSub []string `json:"-"` // publish in user channelsfor subscribe/unsubscribe only
	OrderCmd  [2]int   `json:"-"` // value 1 means PublishWithMessage, value 2 means PublishWithSubscriptions, 0 means nothing
	Data      []byte   `json:"-"`
}

func (e *Event) GetEventName() string { // in order to not unmarshaling twice
	start := bytes.IndexByte(e.Data, '"')
	if start == -1 {
		return ""
	}
	start++
	end := bytes.IndexByte(e.Data[start:], '"')
	if end == -1 {
		return ""
	}
	return string(e.Data[start : start+end])
}
