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
		val := u.Elem.Next()
		if val != nil && val.Value != nil {
			val, ok := val.Value.(*UnrolledItems)
			if ok && !val.IsNil() {
				val.CurrentIndex = 0
				return val
			}
		}
		return nil
	}
	u.CurrentIndex++
	return u
}
func (u *UnrolledItems) Prev() ListItem {
	if u.CurrentIndex == 0 {
		val := u.Elem.Prev()
		if val != nil && val.Value != nil {
			val, ok := val.Value.(*UnrolledItems)
			if ok && !val.IsNil() {
				val.CurrentIndex = len(val.Items) - 1
				return val
			}
		}
		return nil
	}
	u.CurrentIndex--
	return u
}
func (u *UnrolledItems) IsNil() bool {
	return u == nil
}
