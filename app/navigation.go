package app

import (
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

func (c *Chat) PreLoadNavigation() {

	options := []func(list.ListItem){c.OptionEvent, c.OptionRoom, c.OptionEvent, c.OptionInput,
		OneOption, OneOption, MultOption, OneOption, MultOption, OneOption}

	for i := range len(c.Lists) {
		c.Lists[i] = list.NewList(list.NewArrayList(c.MaxMsgsOnPage), options[i])
	}

	c.Lists[1].Items = list.NewLinkedList(250)
	c.MainFlex = tview.NewFlex()

	ll := c.Lists[0].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "Events", ""))
	ll = c.Lists[0].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "Menu", ""))

	ll = c.Lists[3].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "", "Enter Text Here"))
	ll = c.Lists[9].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "true", ""))
	ll = c.Lists[9].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "false", ""))
	c.AddItemMainFlex()

	log.Println(c.Lists[0].Items.Len(), "WTF IS THIS GAME AOBUT")
	item := c.Lists[0].Items.GetBack()
	for item != nil && !item.IsNil() {
		log.Println(item.GetMainText(), item.GetSecondaryText(), item.GetColor(0))
		item = item.Next()
	}

	c.MainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if c.IsInputActive {
			switch event.Key() {
			case tcell.KeyRune: // write letter via buffer.WriteString
				txt := c.Lists[3].Items.GetBack().GetMainText()
				if len([]rune(txt)) <= 300 {
					r := event.Rune()
					c.Lists[3].Items.GetBack().SetMainText(string(r), 1)
				}
			case tcell.KeyBackspace, tcell.KeyBackspace2: // trimming via buffer.truncate
				main := c.Lists[3].Items.GetBack().GetMainText()
				mr := []rune(main)
				if len(mr) != 0 {
					mr = mr[:len(mr)-1]
					c.Lists[3].Items.GetBack().SetMainText(string(mr), 2)
				}
			}
		}
		switch event.Key() {
		case tcell.KeyLeft:
			if c.NavState > 0 && c.MainFlex.GetItemCount() > 0 {
				c.NavState -= 1
				c.App.SetFocus(c.MainFlex.GetItem(c.NavState))

			}
		case tcell.KeyRight:
			if c.NavState < c.MainFlex.GetItemCount()-1 {
				c.NavState += 1
				c.App.SetFocus(c.MainFlex.GetItem(c.NavState))
			}
		}
		return event
	})
	c.App = tview.NewApplication()
}

/*
one of the most important methods, pressing a functional option either changes the state of the current UI
or sends a request to the server
*/
func (c *Chat) OptionEvent(item list.ListItem) {
	key := list.Content{}
	key.MainText = item.GetMainText()
	key.SecondaryText = item.GetSecondaryText()
	ev, ok := c.EventMap[key]
	if ok {
		if len(key.MainText) == 0 {
			l := ev.targets[0]
			sel := []list.Content{}
			if l >= 0 {
				sel = c.Lists[l].GetSelected()
				if len(sel) < 1 {
					return
				}
			}
			str1 := strconv.FormatUint(c.CurrentRoom.RoomId, 10)
			str2 := c.Lists[3].Items.GetBack().GetMainText()
			sel = append(sel, list.Content{MainText: str1}, list.Content{MainText: str2}) /* by default always adds-
			the current room's id and input area text*/
			ev.Kind(sel, ev.targets[1:]...)
			return
		}
		ev.Kind(ev.content, ev.targets...)
	}
}

func (c *Chat) OptionRoom(item list.ListItem) {
	if item == nil {
		return
	}
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
			c.CurrentRoom = rm
			c.Lists[8].Items.Clear()

			for _, v := range c.CurrentRoom.Users {
				navitem := c.Lists[8].Items.NewItem(
					[2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
					v.Username,
					strconv.FormatUint(v.UserId, 10),
				)
				c.Lists[8].Items.MoveToBack(navitem)
			}

			if c.CurrentRoom.IsGroup {
				if c.CurrentRoom.IsAdmin {
					c.Lists[0].Items.GetBack().SetMainText("This Group Room(Admin)", 0)
				} else {
					c.Lists[0].Items.GetBack().SetMainText("This Group Room", 0)
				}
			} else {
				c.Lists[0].Items.GetBack().SetMainText("This Duo Room", 0)
			}
			c.AddItemMainFlex(rm.Messages[rm.MsgPageIdFront], c.Lists[3])
		}
	} else {
		c.PaginationRoom(main)
	}
}

func (c *Chat) PaginationRoom(maintxt string) {
	v, err := strconv.Atoi(maintxt[11:])
	if err != nil {
		return
	}
	l, ok := c.CurrentRoom.Messages[v]
	if ok {
		c.AddItemMainFlex(l, c.Lists[3])
	} else {
		ev := c.EventMap[list.Content{SecondaryText: "Get Messages From Room"}]

		str1 := strconv.FormatUint(c.CurrentRoom.RoomId, 10)
		str2 := c.Lists[3].Items.GetBack().GetMainText()
		ev.Kind([]list.Content{{MainText: strconv.FormatUint(c.CurrentRoom.LastMessageID, 10)},
			{MainText: str1}, {MainText: str2}})
	}
}

func (c *Chat) OptionInput(item list.ListItem) {
	c.IsInputActive = true
	c.Lists[3].Items.GetBack().SetSecondaryText("Type The Text")
}
func (c *Chat) AddItemMainFlex(prmtvs ...tview.Primitive) {
	log.Println(len(prmtvs))
	c.NavState = 0
	c.IsInputActive = false
	itemio := c.Lists[3].Items.GetBack()
	if itemio != nil && !itemio.IsNil() {
		itemio.SetSecondaryText("Press Enter Here To Type Text")

		c.MainFlex.Clear()
		c.MainFlex.AddItem(c.Lists[0], 0, 2, true)
		c.MainFlex.AddItem(c.Lists[1], 0, 2, true)
		for _, v := range prmtvs {
			c.MainFlex.AddItem(v, 0, 2, true)
		}
	} else {
		panic("this shouldnt happen")
	}
}
