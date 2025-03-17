package chat

import (
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

// animation for in progress events
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
		case tcell.KeyRune: // write letter via buffer.WriteString
			if c.IsInputActive {
				txt := c.Lists[3].Items.GetFront().GetMainText()
				if len([]rune(txt)) <= 300 {
					r := event.Rune()
					c.Lists[3].Items.GetFront().SetMainText(string(r), 1)
				}
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2: // trimming via buffer.truncate
			if c.IsInputActive {
				main := c.Lists[3].Items.GetFront().GetMainText()
				mr := []rune(main)
				if len(mr) != 0 {
					mr = mr[:len(mr)-1]
					c.Lists[3].Items.GetFront().SetMainText(string(mr), 2)
				}
			}
		}
		return event
	})
}

/*
one of the most important methods, pressing a functional option either changes the state of the current UI
or sends a request to the server if it is a form submission
*/
func (c *Chat) OptionEvent(item list.ListItem) {
	key := list.Content{}
	key.MainText = item.GetMainText()
	key.SecondaryText = item.GetSecondaryText()
	ev := c.EventMap[key]
	if len(key.MainText) == 0 {
		l := ev.targets[0]
		sel := []list.Content{}
		if l >= 0 {
			sel = c.Lists[l].GetSelected()
		}
		str1 := strconv.FormatUint(c.currentRoom.RoomId, 10)
		str2 := c.Lists[3].Items.GetFront().GetMainText()

		sel = append(sel, list.Content{MainText: str1}, list.Content{MainText: str2}) /* by default always adds-
		the current room's id and input area text*/
		ev.Kind(sel, ev.targets[1:]...)
		return
	}
	ev.Kind(ev.content, ev.targets...)
}

func (c *Chat) OptionRoom(item list.ListItem) {
	sec := item.GetSecondaryText()
	main := item.GetMainText()
	if sec[:9] == "Room Id: " {
		/*When changing a room, add current users to the list for current users of
		the room of the selected room (it does not allocate anything, just copies it).*/
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
					navitem := ll.NewItem(
						[2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
						v.Username,
						strconv.FormatUint(v.UserId, 10),
					)
					c.Lists[8].Items.MoveToFront(navitem)
				}
			}
			if c.currentRoom.IsGroup {
				if c.currentRoom.IsAdmin {
					c.Lists[0].Items.GetFront().SetMainText("This Group Room(Admin)", 0)
				} else {
					c.Lists[0].Items.GetFront().SetMainText("This Group Room", 0)
				}
			} else {
				c.Lists[0].Items.GetFront().SetMainText("This Duo Room", 0)
			}
			c.AddItemMainFlex(rm.Messages[rm.MsgPageIdFront], c.Lists[3])
		}
	} else {
		v, err := strconv.Atoi(main[11:])
		if err != nil {
			return
		}
		l, ok := c.currentRoom.Messages[v]
		if ok {
			c.AddItemMainFlex(l, c.Lists[3])
		} else {
			ev := c.EventMap[list.Content{SecondaryText: "Get Messages From Room"}]

			str1 := strconv.FormatUint(c.currentRoom.RoomId, 10)
			str2 := c.Lists[3].Items.GetFront().GetMainText()
			ev.Kind([]list.Content{{MainText: strconv.FormatUint(c.currentRoom.lastMessageID, 10)},
				{MainText: str1}, {MainText: str2}})
		}

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
