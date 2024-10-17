package models

import "bytes"

type Event struct {
	Event       string   `json:"Event"`
	ErrorMsg    string   `json:"ErrorMsg"`
	Kind        string   // subscribe or unsubscribe
	PubChannels []string // publish channels
	SubChannel  string   // subscription channel
	UserId      uint32   `json:"UserId"`
	Data        []byte   `json:"-"`
}

func (e *Event) GetEventName() string { // in order to not marshaling twice, but for the cards to be empty
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
