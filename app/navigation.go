package main

import (
	"encoding/json"
	"strconv"
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

func (c *Chat) StartEventUILoop() {
	i := 0
	var NotIdle bool
	for {
		if state.InProgressCount > 0 {
			c.App.QueueUpdateDraw(func() {

				spinChar := state.spinner[i%len(state.spinner)]
				text := spinChar + " " + strconv.Itoa(state.InProgressCount) + " " + state.message
				item := c.Lists[3].Items.GetFront()
				item.SetSecondaryText(text)
				item.SetColor(tcell.ColorRed, 1)
			})
			i++
			if i == len(state.spinner) {
				i = 0
			}
			NotIdle = true
		} else if NotIdle {
			c.App.QueueUpdateDraw(func() {
				item := c.Lists[3].Items.GetFront()
				item.SetSecondaryText(" ")
				item.SetColor(tcell.ColorGrey, 1)
			})
			NotIdle = false
		} else if c.stopeventUI {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func (c *Chat) PreLoadElems() {
	c.NavText = [16]string{"Menu", "Event logs", "Create Duo Room", "Create Group Room",
		"Unblock User", "Get Blocked Users", "Change Username", "Change Privacy",
		"for Duo Rooms", " for Group Rooms", "Add Users To Room",
		"Delete Users From Room", "Show users", "Change Room Name", "Block", "Leave Room"}
	c.MainFlex = tview.NewFlex()
	//----------------------------------------------------------------
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
	//----------------------------------------------------------------
	c.InputField = tview.NewForm()
	c.InputField.AddInputField("Enter text", "", 0, func(textToCheck string, lastChar rune) bool {
		return len(textToCheck) != 0
	}, func(msg string) {
		c.CurrentText = msg
	})
	//----------------------------------------------------------------
	c.SendMsgBtn = tview.NewForm()
	c.SendMsgBtn.AddButton("Send Message", func() {
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
	//----------------------------------------------------------------
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
		//
		c.AddItemMainFlex(c.Lists[3], c.Lists[4])
	case "Create Duo Room":
		c.AddItemMainFlex(c.Lists[3], c.Lists[4])
	case "Create Group Room":
		c.AddItemMainFlex(c.Lists[3], c.Lists[4])
	case "Unblock User":
		if c.LastNavigation == "Unblock User" {
			itms := c.Lists[1].Selector.GetItems()
			if len(itms) != 0 {
				n, err := strconv.ParseUint(itms[0], 10, 64)
				if err != nil {
					c.Lists[5].Items.MoveToFront(list.ArrayItem{MainText: "Unblock User",
						SecondaryText: "Error: no selected user"})
				}
				event := User{
					Event:  "UnblockUser",
					UserId: n,
				}
				b, err := json.Marshal(event)
				if err != nil {
					WriteTimeout(time.Second*5, c.Conn, b)
				}
			}
		} else {
			c.Lists[4].Items.Clear()
			for i := 4; i < 6; i++ {
				c.Lists[4].Items.MoveToFront(list.ArrayItem{MainText: c.NavText[i]})
			}
			c.AddItemMainFlex(c.Lists[3], c.Lists[4], c.Lists[1])
		}
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
	c.LastNavigation = item.GetMainText()
}
func (c *Chat) Clear() {
	c.LastNavigation = ""
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
