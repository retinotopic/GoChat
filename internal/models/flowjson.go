package models

type Flowjson struct {
	Mode      string   `json:"Mode"`
	Message   string   `json:"Message" db:"payload"`
	Users     []uint32 `json:"Users"`
	User      uint32   `json:"User" db:"user_id"`
	Room      uint32   `json:"Room" db:"room_id"`
	Name      string   `json:"Name" db:"name"`
	MessageId string   `json:"Offset" db:"message_id"`
	ErrorMsg  string   `json:"ErrorMsg"`
	Bool      bool     `json:"Bool"`
}
