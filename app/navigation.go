package main

import (
	"strconv"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

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
	c.MainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyLeft {
			c.App.SetFocus(c.MainFlex.GetItem(c.NavState - 1))
			return nil
		}
		if event.Key() == tcell.KeyRight {
			c.App.SetFocus(c.MainFlex.GetItem(c.NavState + 1))
			return nil
		}
		if c.IsInputActive {
			k := []rune(event.Name())[5:6][0] // take the name of the event, convert to runes, take only the name of key via slicing
			if unicode.IsPrint(k) {           // write only printable key
				txt := c.Lists[3].Items.GetFront().GetMainText()
				c.Lists[3].Items.GetFront().SetMainText(txt + string(k))
			}
		}
		return event
	})

}
func (c *Chat) OptionBtn(item list.ListItem) {

	// todo : rewrite
}
func (c *Chat) OptionRoom(item list.ListItem) {
	sec := item.GetSecondaryText() // room id
	//main := item.GetMainText()     // btn prev or next
	if sec[:9] == "Room Id: " {
		v, err := strconv.ParseUint(sec[9:], 10, 64)
		if err != nil {
			return
		}
		rm, ok := c.RoomMsgs[v]
		if ok {
			c.AddItemMainFlex(rm.Messages[rm.MsgPageIdFront], c.Lists[3])
		}
	}
	// todo : rewrite
}
func (c *Chat) OptionInput(item list.ListItem) {
	c.IsInputActive = true
	// todo : rewrite
}
func (c *Chat) AddItemMainFlex(prmtvs ...tview.Primitive) {
	// todo: rewrite
	c.MainFlex.Clear()
	c.MainFlex.AddItem(c.Lists[0], 0, 2, true)
	c.MainFlex.AddItem(c.Lists[1], 0, 2, true)
	for _, v := range prmtvs {
		c.MainFlex.AddItem(v, 0, 2, true)
	}
}
