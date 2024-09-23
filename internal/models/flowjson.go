package models

import "github.com/goccy/go-json"

type Event struct {
	Mode       string          `json:"Mode"`
	PayloadArr []uint32        `json:"Users"`
	UserId     uint32          `json:"UserId"`
	Payload    string          `json:"Payload" `
	ErrorMsg   string          `json:"ErrorMsg"`
	Response   json.RawMessage `json:"Response" `
}
