package list

import (
	"bytes"
	"github.com/gdamore/tcell/v2"
	"unicode/utf8"
)

func NewLinkedList(lenl int) *LinkedList {
	ll := &LinkedList{}
	ll.items = make([]LinkedItem, lenl)
	ll.stack = make([]int, lenl)
	ll.root.next = &ll.root
	ll.root.prev = &ll.root
	for i := range ll.items {
		ll.items[i].idx = i
		ll.items[i].parent = ll
	}
	ll.Clear()
	return ll
}

type LinkedList struct {
	root  LinkedItem
	stack []int
	items []LinkedItem
}

func (l *LinkedList) MoveToBack(e ListItem) {
	uitem, ok := e.(*LinkedItem)
	if !ok || uitem.parent != l || l.root.next == uitem {
		return
	}
	l.move(uitem, &l.root)

}
func (l *LinkedList) MoveToFront(e ListItem) {
	uitem, ok := e.(*LinkedItem)
	if !ok || uitem.parent != l || l.root.prev == uitem {
		return
	}
	l.move(uitem, l.root.prev)

}

func (l *LinkedList) move(e, at *LinkedItem) {
	if e == at {
		return
	}
	if e.prev == nil || e.next == nil {
		e.prev = at
		e.next = at.next
		e.prev.next = e
		e.next.prev = e
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

func (l *LinkedList) GetBack() ListItem {
	if len(l.items)-len(l.stack) == 0 {
		return nil
	}
	return l.root.next
}
func (l *LinkedList) GetFront() ListItem {
	if len(l.items)-len(l.stack) == 0 {
		return nil
	}
	return l.root.prev
}

func (l *LinkedList) Remove(e ListItem) {
	uitem, ok := e.(*LinkedItem)
	if ok && uitem.parent == l {
		uitem.prev.next = uitem.next
		uitem.next.prev = uitem.prev
		uitem.next = nil
		uitem.prev = nil
		l.stack = append(l.stack, uitem.idx)
	}
}

func (l *LinkedList) Clear() {

	l.root.next = &l.root
	l.root.prev = &l.root
	l.stack = l.stack[:0]
	for i := range l.items {
		l.items[i].next = nil
		l.items[i].prev = nil
		l.stack = append(l.stack, i)
	}
}

func (l *LinkedList) Len() int {
	return len(l.items) - len(l.stack)
}

func (l *LinkedList) NewItem(clr [2]tcell.Color, main string, sec string) ListItem {
	if len(l.stack) == 0 {
		return nil
	}
	ls := l.stack[len(l.stack)-1]
	l.stack = l.stack[:len(l.stack)-1]
	li := &l.items[ls]
	li.Color = clr
	li.MainText = main
	li.SecondaryText = sec
	return li
}

type LinkedItem struct {
	idx         int
	parent      *LinkedList
	Color       [2]tcell.Color
	MainText    string
	MainTextBuf *bytes.Buffer /* by default is nil,
	only used if item changing often*/
	SecondaryText string
	next, prev    *LinkedItem
}

func (l *LinkedItem) GetParent() ListItems {
	return l.parent
}
func (l *LinkedItem) GetMainText() string {
	if l.MainTextBuf != nil {
		return l.MainTextBuf.String()
	}
	return l.MainText

}

func (l *LinkedItem) GetSecondaryText() string {
	return l.SecondaryText
}

func (l *LinkedItem) GetColor(idx int) tcell.Color {
	if idx < 2 && idx >= 0 {
		return l.Color[idx]
	}
	return tcell.ColorWhite
}

func (l *LinkedItem) SetMainText(str string, mode uint8) {
	switch mode {
	case 0:
		l.MainText = str
	case 1:
		if l.MainTextBuf != nil {
			l.MainTextBuf.WriteString(str)
			return
		}
		l.MainTextBuf = bytes.NewBufferString(str)
	case 2:
		if l.MainTextBuf != nil {
			_, size := utf8.DecodeLastRuneInString(l.MainTextBuf.String())
			l.MainTextBuf.Truncate(len(l.MainTextBuf.String()) - size)
		}
	}
}

func (l *LinkedItem) SetSecondaryText(str string) {
	l.SecondaryText = str
}

func (l *LinkedItem) SetColor(clr tcell.Color, idx int) {
	if idx < 2 && idx >= 0 {
		l.Color[idx] = clr
	}
}

func (l *LinkedItem) Next() ListItem {
	if p := l.next; l.parent != nil && p != &l.parent.root {
		return p
	}
	return nil
}

func (l *LinkedItem) Prev() ListItem {
	if p := l.prev; l.parent != nil && p != &l.parent.root {
		return p
	}
	return nil
}

// val.parent.front == val
func (l *LinkedItem) IsNil() bool { // if interface is not nil but interface value is
	return l == nil
}
