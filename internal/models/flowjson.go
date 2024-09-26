package models

type Event struct {
	Event        string `json:"Event"`
	UserId       uint32 `json:"UserId"`
	Subscription string
	ErrorMsg     string `json:"ErrorMsg"`
	Data         []byte `json:"Data" `
}
