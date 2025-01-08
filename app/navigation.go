package main

import (
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

var NavText = []string{"Menu", "Create Duo Room", "Create Group Room", "Unblock User", "Change Username", "Current Room Actions", "Change Privacy",
	"Current Room Actions", "Block", "Leave Room", "Show users", "Add Users To Room", "Delete Users From Room", "Change Room Name",
	"Events", "Change Privacy", "for Duo Rooms", "for Group Rooms", "Unblock User", "Get Blocked Users", "Unblock User"}

var NavigateEventtMap map[string]NavigateEvent

type NavigateEvent struct {
	From  int // slicing NavText from -> to, like NavText[From:To]
	To    int
	Lists []int // lists for main *tview.Flex
}
type LoadingState struct {
	message         string
	spinner         []string
	color           string
	InProgressCount int
}

var state = LoadingState{
	message:         "In Progress",
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

// MARK: ELEMS
func (c *Chat) PreLoadElems() {

	c.MainFlex = tview.NewFlex()
	//----------------------------------------------------------------
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
		event.Rune()
		return event
	})

}
func (c *Chat) Option(item list.ListItem) {

	// todo : rewrite
}

func (c *Chat) AddItemMainFlex(prmtvs ...tview.Primitive) {
	// todo: rewrite
	c.MainFlex.Clear()
	c.MainFlex.AddItem(c.Lists[0], 0, 2, true)
	c.MainFlex.AddItem(c.Lists[7], 0, 2, true)
	for _, v := range prmtvs {
		c.MainFlex.AddItem(v, 0, 2, true)
	}
}
