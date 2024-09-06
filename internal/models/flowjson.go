package models

type Flowjson struct {
	Mode      string   `json:"Mode"`
	Message   string   `json:"Message"`
	Users     []uint32 `json:"Users"`
	UserId    uint32   `json:"UserId"`
	RoomId    uint32   `json:"RoomId" `
	Name      string   `json:"Name" `
	MessageId string   `json:"MessageId" `
	ErrorMsg  string   `json:"ErrorMsg"`
	Bool      bool     `json:"Bool"`
}
