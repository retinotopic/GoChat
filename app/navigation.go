package main

import (
	"strconv"
	"time"

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
				item := c.Lists[0].Items.GetFront()
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
				item := c.Lists[0].Items.GetFront()
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
func (c *Chat) PreLoadNavigation() {
	c.MainFlex = tview.NewFlex()
	c.MainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			c.App.SetFocus(c.MainFlex.GetItem(c.NavState - 1))
			return event
		case tcell.KeyRight:
			c.App.SetFocus(c.MainFlex.GetItem(c.NavState + 1))
			return event
		case tcell.KeyRune:
			if c.IsInputActive {
				txt := c.Lists[3].Items.GetFront().GetMainText()
				if len([]rune(txt)) <= 300 {
					r := event.Rune()
					c.Lists[3].Items.GetFront().SetMainText(txt + string(r))
				}
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if c.IsInputActive {
				main := c.Lists[3].Items.GetFront().GetMainText()
				mr := []rune(main)
				if len(mr) != 0 {
					mr = mr[:len(mr)-1]
					c.Lists[3].Items.GetFront().SetMainText(string(mr))
				}
			}
		}
		return event
	})
}

func (c *Chat) OptionEvent(item list.ListItem) {
	key := list.Content{}
	text := item.GetMainText()
	key.IsMain = true
	key.Text = text
	ev := c.EventMap[key]
	if len(text) == 0 {
		key.IsMain = false
		l := ev.targets[0]
		sel := c.Lists[l].GetSelected()
		str1 := strconv.FormatUint(c.currentRoom.RoomId)
		str2 := c.Lists[3].Items.GetFront().GetMainText()
		sel = append(sel, list.Content{Text: str1}, list.Content{Text: str2})
		ev.Kind(sel, ev.targets[1:]...)

	}
	ev.Kind(ev.content, ev.targets...)
}

func (c *Chat) OptionRoom(item list.ListItem) {
	sec := item.GetSecondaryText()
	main := item.GetMainText()
	if sec[:9] == "Room Id: " {
		v, err := strconv.ParseUint(sec[9:], 10, 64)
		if err != nil {
			return
		}
		rm, ok := c.RoomMsgs[v]
		if ok {
			c.currentRoom = rm
			c.Lists[8].Items.Clear()
			ll, ok := c.Lists[8].Items.(*list.ArrayList)
			if ok {
				for _, v := range c.currentRoom.Users {
					navitem := list.NewArrayItem(
						ll,
						[2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
						v.Username,
						strconv.FormatUint(v.UserId, 10),
					)
					c.Lists[8].Items.MoveToFront(navitem)
				}
			}
			//c.NavigateEventMap["Current Room Actions"].Lists
			c.AddItemMainFlex(rm.Messages[rm.MsgPageIdFront], c.Lists[3])
		}
	} else {
		v, err := strconv.Atoi(main[11:])
		if err != nil {
			return
		}
		c.AddItemMainFlex(c.currentRoom.Messages[v], c.Lists[3])
	}
}
func (c *Chat) OptionInput(item list.ListItem) {
	c.IsInputActive = true
}
func (c *Chat) AddItemMainFlex(prmtvs ...tview.Primitive) {
	c.MainFlex.Clear()
	for _, v := range prmtvs {
		c.MainFlex.AddItem(v, 0, 2, true)
	}
}
