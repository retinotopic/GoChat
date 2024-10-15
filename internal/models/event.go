package models

type Event struct {
	Event       string   `json:"Event"`
	ErrorMsg    string   `json:"ErrorMsg"`
	Kind        string   // subscribe or unsubscribe
	PubChannels []string // publish channels
	SubChannel  string   // subscription channel
	UserId      uint32   `json:"UserId"`
	Data        []byte   `json:"-"`
}
