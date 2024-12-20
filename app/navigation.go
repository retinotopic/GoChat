package main

import (
	"encoding/json"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

type LoadingState struct {
	message         string
	spinner         []string
	color           string
	InProgressCount int
}

var state = LoadingState{
	message:         " In Progress",
	spinner:         []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	color:           "yellow",
	InProgressCount: 0,
}

func (c *Chat) PreLoadElems() {
	c.NavText = [13]string{"Event logs", "Create Duo Room", "Create Group Room", "Unblock User", "Change Username", "Change Privacy for Duo Rooms",
		"Change Privacy for Group Rooms", "Add Users To Room", "Delete Users From Room", "Show users", "Change Room Name", "Block", "Leave Room"}
	c.MainFlex = tview.NewFlex()
	c.FindUsersForm = tview.NewForm().
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
	c.RoomMenuForm = tview.NewForm()
	c.RoomMenuForm.AddInputField("Enter text", "", 0, func(textToCheck string, lastChar rune) bool {
		return len(textToCheck) != 0
	}, func(msg string) {
		c.CurrentText = msg
	})
	c.RoomMenuForm.AddButton("Send Message", func() {
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
	c.RoomMenuForm.AddButton("Room Actions", func() {

	})
	//-------------------------------------------------
	c.MainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyLeft {
			c.App.SetFocus(c.MainFlex.GetItem(c.NavState - 1))
			return nil
		}
		if event.Key() == tcell.KeyRight {
			c.App.SetFocus(c.MainFlex.GetItem(c.NavState + 1))
			return nil
		}
		return event
	})

}
func (c *Chat) Option(item list.ListItem) {

	switch item.GetMainText() {
	case "Event logs":
		c.AddItemMainFlex(c.Lists[0], c.Lists[4])
	case "Create Duo Room":
		//c.AddItemMainFlex(c.Lists[0], c.Lists[3])
	case "Create Group Room":
		//c.AddItemMainFlex(c.Lists[0], c.Lists[3])
	case "Unblock User":
		//c.AddItemMainFlex(c.Lists[0], c.Lists[3])
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
	default:

	}
}
func (c *Chat) Clear() {
}
func (c *Chat) GetItems() []string {
	return []string{}
}

func (c *Chat) AddItemMainFlex(prmtvs ...tview.Primitive) {
	c.MainFlex.Clear()
	for _, v := range prmtvs {
		c.MainFlex.AddItem(v, 0, 2, true)
	}
}
