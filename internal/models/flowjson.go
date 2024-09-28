package models

type Event struct {
	Event        string `json:"Event"`
	Subscription bool   // publish with subscription or not
	Unsubscribe  bool   // subscribe or unsubsribe
	Publish      string // publish channel
	UserId       uint32 `json:"UserId"`
	ErrorMsg     string `json:"ErrorMsg"`
	Data         []byte `json:"-"`
}
