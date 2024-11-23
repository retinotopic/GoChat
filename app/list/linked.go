package list

import (
	lst "container/list"

	"github.com/gdamore/tcell/v2"
)

func NewLinkedList() *LinkedList {
	ll := &LinkedList{}
	ll.Items = lst.New()
	return ll
}

type LinkedList struct {
	Items *lst.List
}

func (l *LinkedList) MoveToFront(e ListItem) {
	uitem, ok := e.(*LinkedItems)
	if ok {
		if uitem.Elem != nil {
			l.Items.MoveToFront(uitem.Elem)
		} else {
			uitem.Elem = l.Items.PushFront(uitem)
		}
	}
}
func (l *LinkedList) MoveToBack(e ListItem) {
	uitem, ok := e.(*LinkedItems)
	if ok {
		if uitem.Elem != nil {
			l.Items.MoveToBack(uitem.Elem)
		} else {
			uitem.Elem = l.Items.PushBack(uitem)
		}
	}
}
func (l *LinkedList) GetFront() ListItem {
	return l.Items.Front().Value.(*LinkedItems)
}
func (l *LinkedList) Remove(e ListItem) {
	uitem, ok := e.(*LinkedItems)
	if ok {
		if uitem.Elem != nil {
			l.Items.Remove(uitem.Elem)
		}
	}
}
func (l *LinkedList) Clear() {
	l.Items.Init()
}

type LinkedItems struct {
	Elem          *lst.Element
	prev          *LinkedItems
	next          *LinkedItems
	Color         tcell.Color
	MainText      string
	SecondaryText string
}

func (l *LinkedItems) GetMainText() string {
	return l.MainText
}
func (l *LinkedItems) GetSecondaryText() string {
	return l.SecondaryText
}
func (l *LinkedItems) GetColor() tcell.Color {
	return l.Color
}
func (l *LinkedItems) SetMainText(str string) {
	l.MainText = str
}
func (l *LinkedItems) SetSecondaryText(str string) {
	l.SecondaryText = str
}
func (l *LinkedItems) SetColor(clr tcell.Color) {
	l.Color = clr
}
func (l *LinkedItems) Next() ListItem {
	return l.next
}
func (l *LinkedItems) Prev() ListItem {
	return l.prev
}
