package models

type Event struct {
	Event        string `json:"Event"`
	UserId       uint32 `json:"UserId"`
	Subscription string
	PubSub       uint32
	ErrorMsg     string `json:"ErrorMsg"`
	Data         []byte `json:"Data" `
}
