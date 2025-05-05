package app

import (
	// "log"
	"log"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

type ListInfo struct {
	f     func(list.ListItem)
	title string
}

func (c *Chat) PreLoadNavigation() {

	// options := []func(list.ListItem){c.OptionEvent, c.OptionRoom, c.OptionEvent, c.OptionInput,
	// 	OneOption, OneOption, MultOption, OneOption, MultOption, OneOption}
	options := []ListInfo{
		{f: c.OptionEvent, title: "Sidebar"},
		{f: c.OptionRoom, title: "Rooms"},
		{f: c.OptionEvent, title: "Menu"},
		{f: c.OptionInput, title: "Input"},
		{f: OneOption, title: "Events"},
		{f: OneOption, title: "FoundUsers"},
		{f: MultOption, title: "DuoUsers"},
		{f: OneOption, title: "BlockedUsers"},
		{f: MultOption, title: "RoomUsers"},
		{f: OneOption, title: "Allow or not"},
	}
	c.Logger.Println("AIODJFOPFJODIPFJOPIDJ 1")
	for i := range len(c.Lists) {
		c.Lists[i] = list.NewList(list.NewArrayList(c.MaxMsgsOnPage), options[i].f, options[i].title, c.Logger)
	}

	c.Lists[1].Items = list.NewLinkedList(250)
	c.Lists[3].Items = list.NewLinkedList(3)
	c.Lists[6].Items = list.NewLinkedList(250)

	c.MainFlex = tview.NewFlex()

	c.Logger.Println("AIODJFOPFJODIPFJOPIDJ 2")
	ll := c.Lists[0].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "Events", ""))
	c.Logger.Println("AIODJFOPFJODIPFJOPIDJ 21")
	ll = c.Lists[0].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "Menu", ""))

	ll = c.Lists[0].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "This Room (Not Selected)", ""))

	l2 := c.Lists[3].Items.(*list.LinkedList)
	l2.MoveToBack(l2.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "", "Press Enter Here To Type Text"))
	ll = c.Lists[9].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "true", ""))
	ll = c.Lists[9].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "false", ""))

	c.Logger.Println("AIODJFOPFJODIPFJOPIDJ 3")
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
					mr = mr[:len(mr)]
					c.Lists[3].Items.GetBack().SetMainText(string(mr), 2)
				}
			}
		}
		switch event.Key() {
		case tcell.KeyLeft:
			if c.NavState > 0 && c.MainFlex.GetItemCount() > 0 {
				c.NavState -= 1
				prm := c.MainFlex.GetItem(c.NavState)
				c.App.SetFocus(prm)
				l := prm.(*list.List)
				l.Current = l.Items.GetBack()

			}
		case tcell.KeyRight:
			if c.NavState < c.MainFlex.GetItemCount()-1 {
				c.NavState += 1
				prm := c.MainFlex.GetItem(c.NavState)
				c.App.SetFocus(prm)
				l := prm.(*list.List)
				l.Current = l.Items.GetBack()

			}
		}
		return event
	})
	c.Logger.Println("AIODJFOPFJODIPFJOPIDJ")
	c.App = tview.NewApplication()
	c.AddItemMainFlex()
}

/*
one of the most important methods, pressing a functional option either changes the state of the current UI
or sends a request to the server
*/
func (c *Chat) OptionEvent(item list.ListItem) {
	if item == nil {
		log.Fatalln(item)
		return
	}
	key := list.Content{}
	key.MainText = item.GetMainText()
	key.SecondaryText = item.GetSecondaryText()
	ev, ok := c.EventMap[key]
	if ok {
		c.Logger.Println(ev, "option event start")
		if len(key.MainText) == 0 {
			l := ev.targets[0]
			sel := []list.Content{}
			c.Logger.Println("yes nil3", l)
			if l >= 0 {
				sel = c.Lists[l].GetSelected()
				if len(sel) < 1 {
					c.Logger.Println("yes nil2")
					return
				}
			}

			c.Logger.Println("yes nil7")

			str1 := ""
			if c.CurrentRoom != nil {
				str1 = strconv.FormatUint(c.CurrentRoom.RoomId, 10)
			}
			c.Logger.Println("yes nil8")
			it := c.Lists[3].Items.GetBack()
			if it == nil || it.IsNil() {
				c.Logger.Println("yes nil")
				return
			}

			c.Logger.Println("yes nil9")
			str2 := it.GetMainText()
			sel = append(sel, list.Content{MainText: str1, SecondaryText: str2})
			/* by default always adds-
			the current room's id and input area text*/
			c.Logger.Println("yes nilnnn")

			ev.Kind(sel, ev.targets[1:]...)
			c.Logger.Println(sel, ev.targets[1:], "option event func")
			return
		}
		ev.Kind(ev.content, ev.targets...)
		c.Logger.Println(ev.content, ev.targets, "option event UI")
	}
}

func (c *Chat) OptionRoom(item list.ListItem) {
	if item == nil {
		return
	}
	main := item.GetSecondaryText()
	/*When changing a room, add current users to the list for current users of
	the room of the selected room (it does not allocate anything, just copies it).*/
	v, err := strconv.ParseUint(main, 10, 64)
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
				strconv.FormatUint(v.UserId, 10),
				v.Username,
			)
			c.Lists[8].Items.MoveToBack(navitem)
		}
		if c.CurrentRoom.IsGroup {
			if c.CurrentRoom.IsAdmin {
				c.Lists[0].Items.GetFront().SetMainText("This Group Room(Admin)", 0)
			} else {
				c.Lists[0].Items.GetFront().SetMainText("This Group Room", 0)
			}
		} else {
			c.Lists[0].Items.GetFront().SetMainText("This Duo Room", 0)
		}

		c.AddItemMainFlex(rm.Messages[rm.MsgPageIdFront], c.Lists[3])
		l := c.Lists[3].Items.(*list.LinkedList)
		l.MoveToFront(l.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "", "Send Message"))
	}
}

func (c *Chat) OptionPagination(item list.ListItem) {
	v, err := strconv.Atoi(item.GetSecondaryText())
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
			{MainText: str1, SecondaryText: str2}})
	}
}

func (c *Chat) OptionInput(item list.ListItem) {
	if item == nil {
		return
	}
	c.IsInputActive = true
	c.Lists[3].Items.GetBack().SetSecondaryText("Type The Text")
	it := c.Lists[3].Items.GetFront()
	if it != nil && it.GetSecondaryText() == "Send Message" && c.Lists[3].Current == it {
		c.OptionEvent(it)
	}
}
func (c *Chat) AddItemMainFlex(prmtvs ...tview.Primitive) {

	itemio := c.Lists[3].Items.GetBack()
	if itemio != nil && !itemio.IsNil() {
		// itemio.SetSecondaryText("Press Enter Here To Type Text")

		c.MainFlex.Clear()
		c.MainFlex.AddItem(c.Lists[0], 0, 1, true)
		c.MainFlex.AddItem(c.Lists[1], 0, 1, false)

		for _, v := range prmtvs {
			c.Logger.Println(v, "in prmtvs additemmainflex")
			c.MainFlex.AddItem(v, 0, 2, false)
		}

		c.InputToDefault()
		c.UserBuf = c.UserBuf[:0]
		for _, v := range c.DuoUsers {
			c.UserBuf = append(c.UserBuf, v)
		}
		c.FillUsers(c.UserBuf, 6)

	} else {
		panic("this shouldnt happen")
	}
}
func (c *Chat) InputToDefault() {
	c.Lists[3].Items.Clear()
	c.IsInputActive = false
	l := c.Lists[3].Items.(*list.LinkedList)
	l.MoveToBack(l.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "", "Press Enter Here To Type Text"))
}
