package app

import (
	// "log"
	"log"
	"maps"
	"slices"
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
		{f: c.OneOption, title: "Events"},
		{f: c.OneOption, title: "FoundUsers"},
		{f: c.MultOption, title: "DuoUsers"},
		{f: c.OneOption, title: "BlockedUsers"},
		{f: c.MultOption, title: "RoomUsers"},
		{f: c.OneOption, title: "Allow or not"},
	}
	for i := range len(c.Lists) {
		c.Lists[i] = list.NewList(list.NewArrayList(c.MaxMsgsOnPage), options[i].f, options[i].title, c.TestLogger)
	}

	c.Lists[1].Items = list.NewLinkedList(250)
	c.Lists[3].Items = list.NewLinkedList(3)

	c.MainFlex = tview.NewFlex()

	ll := c.Lists[0].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
		"Events", ""))

	ll = c.Lists[0].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
		"Menu", ""))

	ll = c.Lists[0].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
		"This Room (Not Selected)", ""))

	l2 := c.Lists[3].Items.(*list.LinkedList)
	l2.MoveToBack(l2.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
		"", "Press Enter Here To Type Text"))

	ll = c.Lists[9].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
		"true", ""))

	ll = c.Lists[9].Items.(*list.ArrayList)
	ll.MoveToBack(ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
		"false", ""))

	c.MainFlex.SetInputCapture(c.MainFlexNavigation)

	c.AddItemMainFlex()
}

/*
pressing a functional option either changes the state of the current UI
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
		c.Logger.Println("Event", "option event start")
		ev.ExecEvent()
	}
}

func (c *Chat) OptionRoom(item list.ListItem) {
	if item == nil || item.IsNil() {
		return
	}
	main := item.GetSecondaryText()
	v, err := strconv.ParseUint(main, 10, 64)
	if err != nil {
		return
	}
	rm, ok := c.RoomMsgs[v]
	if ok {
		c.CurrentRoom = rm
		c.Lists[8].Items.Clear()
		sortedKeys := slices.Sorted(maps.Keys(c.CurrentRoom.Users))
		for _, k := range sortedKeys {
			v := c.CurrentRoom.Users[k]
			navitem := c.Lists[8].Items.NewItem(
				[2]tcell.Color{tcell.ColorBlue, tcell.ColorWhite},
				v.Username,
				strconv.FormatUint(v.UserId, 10),
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
		item.SetColor(tcell.ColorBlue, 0)
		item.SetColor(tcell.ColorWhite, 1)

		c.AddItemMainFlex(rm.Messages[rm.MsgPageIdFront], c.Lists[3])
		l := c.Lists[3].Items.(*list.LinkedList)
		l.MoveToFront(l.NewItem([2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue}, "", "Send Message"))
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
		l.Current = l.Items.GetBack()
		li := c.Lists[3].Items.(*list.LinkedList)
		li.MoveToFront(li.NewItem([2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue}, "", "Send Message"))
	} else {
		ev := c.EventMap[list.Content{SecondaryText: "Get Messages From Room"}]
		ev.ExecEvent()
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

		c.MainFlex.Clear()
		c.MainFlex.AddItem(c.Lists[0], 0, 1, true)
		c.MainFlex.AddItem(c.Lists[1], 0, 1, false)

		for _, v := range prmtvs {
			c.MainFlex.AddItem(v, 0, 2, false)
		}
		c.InputToDefault()
		c.NavState = 0
		prm := c.MainFlex.GetItem(c.NavState)
		c.App.SetFocus(prm)

		l := prm.(*list.List)
		l.Current = l.Items.GetBack()

		c.Logger.Println("Event", c.NavState, "navstatte")

	} else {
		panic("this shouldnt happen")
	}
}
func (c *Chat) InputToDefault() {
	c.Lists[3].Items.Clear()
	c.IsInputActive = false
	l := c.Lists[3].Items.(*list.LinkedList)
	l.MoveToBack(l.NewItem([2]tcell.Color{tcell.ColorBlue, tcell.ColorBlue}, "", "Press Enter Here To Type Text"))
}

func (c *Chat) MainFlexNavigation(event *tcell.EventKey) *tcell.EventKey {
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
			if len(main) != 0 {
				c.Lists[3].Items.GetBack().SetMainText("", 2)
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
}
