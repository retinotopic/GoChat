package app

import (
	"maps"
	"slices"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
	"github.com/rivo/tview"
)

type UIEvent struct {
	C         *Chat
	Content   []list.Content
	ShowLists []int
}

func (u UIEvent) ExecEvent() {
	lists := make([]tview.Primitive, 0, 5)
	if len(u.Content) > 0 {
		u.C.Lists[u.ShowLists[0]].Items.Clear()
	}
	ll, ok := u.C.Lists[u.ShowLists[0]].Items.(*list.ArrayList)
	if ok {
		// u.C.Logger.Println("Event", "Correct Branch")
		for i := range u.Content {
			a := ll.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite},
				u.Content[i].MainText, u.Content[i].SecondaryText)
			u.C.Lists[u.ShowLists[0]].Items.MoveToBack(a)
		}
		for i := range u.ShowLists {
			lists = append(lists, u.C.Lists[u.ShowLists[i]])
			u.C.Logger.Println("Event", u.ShowLists[i], "event ui list id")
		}

		u.C.UserBuf = u.C.UserBuf[:0]
		sortedKeys := slices.Sorted(maps.Keys(u.C.DuoUsers))
		for _, k := range sortedKeys {
			v := u.C.DuoUsers[k]
			u.C.UserBuf = append(u.C.UserBuf, v)
		}
		u.C.FillUsers(u.C.UserBuf, 6)

		u.C.AddItemMainFlex(lists...)
	}
}

type SendEvent struct {
	C          *Chat
	InitEvent  EventInfo
	ExecFn     func(*EventInfo, []list.Content) error
	TargetList int
}

func (s SendEvent) ExecEvent() {
	initevent := s.InitEvent
	sel := []list.Content{}
	if s.TargetList >= 0 {
		sel = s.C.Lists[s.TargetList].GetSelected()
		if len(sel) < 1 {
			return
		}
	}

	str1 := ""
	if s.C.CurrentRoom != nil {
		str1 = strconv.FormatUint(s.C.CurrentRoom.RoomId, 10)
	}
	it := s.C.Lists[3].Items.GetBack()
	if it == nil || it.IsNil() {
		return
	}

	str2 := it.GetMainText()
	sel = append(sel, list.Content{MainText: str1, SecondaryText: str2})
	err := s.ExecFn(&initevent, sel)
	if err != nil {
		s.C.Logger.Println("ExecFn SendEvent error: ", err)
		return
	}
	s.C.SendEventCh <- initevent
}
