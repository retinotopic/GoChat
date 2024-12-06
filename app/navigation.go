package main

import (
	"encoding/json"
	"time"

	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

func (c *Chat) PreLoadElems() {
	c.NavText = [13]string{"Event logs", "Create Duo Room", "Create Group Room", "Unblock User", "Change Username", "Change Privacy for Duo Rooms",
		"Change Privacy for Group Rooms", "Add Users To Room", "Delete Users From Room", "Show users", "Change Room Name", "Block", "Leave Room"}
	c.MainFlex = tview.NewFlex()
	FindUsersForm := tview.NewForm().
		AddInputField("First name", "", 20, nil, func(text string) {
			c.CurrentText = text
		}).
		AddButton("Find", func() {
			event := User{
				Event:    "FindUsers",
				Username: c.CurrentText,
			}
			b, err := json.Marshal(event)
			if err != nil {
				WriteTimeout(time.Second*5, c.Conn, b)
			}
		})
	//------------------------------------------
	roommenu := tview.NewForm()
	roommenu.AddInputField("Enter text", "", 0, func(textToCheck string, lastChar rune) bool {
		if len(textToCheck) == 0 {
			return false
		}
		return true
	}, func(msg string) {
		c.CurrentText = msg
	})
	roommenu.AddButton("Send Message", func() {
		event := Message{
			Event:          "SendMessage",
			MessagePayload: c.CurrentText,
			RoomId:         c.currentRoom.RoomId,
		}
		b, err := json.Marshal(event)
		if err != nil {
			WriteTimeout(time.Second*5, c.Conn, b)
		}
	})
	/*roommenu.AddButton("Room Actions", func() {
		c.NavigationOptions("Room Actions")
	})

	c.MainFlex.AddItem()*/
}
func (c *Chat) NavigationOptions(item list.ListItem) {
	switch item.GetMainText() {
	case "Event logs":
	case "Create Duo Room":
	case "Create Group Room":
	case "Unblock User":
	case "Change Username":
	case "Current Room Actions":
	case "Update Blocked Users":
	case "Change Privacy for Duo Rooms":
	case "Change Privacy for Group Rooms":
	case "Add Users To Room":
	case "Delete Users From Room":
	case "Show users":
	case "Change Room Name":
	case "Block":
	case "Leave Room":
	case "Menu":
	}
}
