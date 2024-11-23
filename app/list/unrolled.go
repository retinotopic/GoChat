package list

import (
	lst "container/list"

	"github.com/gdamore/tcell/v2"
)

func NewUnrolledList() *UnrolledList {
	ll := &UnrolledList{}
	ll.Items = lst.New()
	return ll
}

type UnrolledList struct {
	Items *lst.List
}

func (u *UnrolledList) MoveToFront(e ListItem) {
	uitem, ok := e.(*UnrolledItems)
	if ok {
		if uitem.Elem != nil {
			u.Items.MoveToFront(uitem.Elem)
		} else {
			uitem.Elem = u.Items.PushFront(uitem)
		}
	}
}
func (u *UnrolledList) MoveToBack(e ListItem) {
	uitem, ok := e.(*UnrolledItems)
	if ok {
		if uitem.Elem != nil {
			u.Items.MoveToBack(uitem.Elem)
		} else {
			uitem.Elem = u.Items.PushBack(uitem)
		}
	}
}
func (u *UnrolledList) GetFront() ListItem {
	return u.Items.Front().Value.(*UnrolledItems)
}
func (u *UnrolledList) Remove(e ListItem) {
	uitem, ok := e.(*UnrolledItems)
	if ok {
		if uitem.Elem != nil {
			u.Items.Remove(uitem.Elem)
		}
	}
}
func (u *UnrolledList) Clear() {
	u.Items.Init()
}

type UnrolledItems struct {
	Elem         *lst.Element
	Items        []UnrolledItem
	prev         *UnrolledItems
	next         *UnrolledItems
	CurrentIndex int
}
type UnrolledItem struct {
	Color         tcell.Color
	MainText      string
	SecondaryText string
}

func (u *UnrolledItems) GetMainText() string {
	return u.Items[u.CurrentIndex].MainText
}
func (u *UnrolledItems) GetSecondaryText() string {
	return u.Items[u.CurrentIndex].SecondaryText
}
func (u *UnrolledItems) GetColor() tcell.Color {
	return u.Items[u.CurrentIndex].Color
}
func (u *UnrolledItems) SetMainText(str string) {
	u.Items[u.CurrentIndex].MainText = str
}
func (u *UnrolledItems) SetSecondaryText(str string) {
	u.Items[u.CurrentIndex].SecondaryText = str
}
func (u *UnrolledItems) SetColor(clr tcell.Color) {
	u.Items[u.CurrentIndex].Color = clr
}
func (u *UnrolledItems) Next() ListItem {
	if u.CurrentIndex == len(u.Items)-1 {
		return u.next
	}
	u.CurrentIndex++
	return u
}
func (u *UnrolledItems) Prev() ListItem {
	if u.CurrentIndex == 0 {
		return u.prev
	}
	u.CurrentIndex--
	return u
}
