package list

import (
	"bytes"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

func NewLinkedList(lenl int) *LinkedList {
	ll := &LinkedList{}
	ll.items = make([]LinkedItem, lenl)
	ll.stack = make([]int, lenl)

	for i := range ll.items {
		ll.items[i].idx = i
		ll.items[i].parent = ll
	}
	ll.Clear()
	return ll
}

type LinkedList struct {
	front *LinkedItem
	back  *LinkedItem
	stack []int
	items []LinkedItem
}

func (l *LinkedList) MoveToFront(e ListItem) {
	uitem, ok := e.(*LinkedItem)
	if ok && uitem != nil && uitem.parent == l && uitem != l.front {

		if uitem.prev != nil {
			uitem.prev.next = uitem.next
		}
		if uitem.next != nil {
			uitem.next.prev = uitem.prev
		}

		if uitem == l.back {
			l.back = uitem.next
		}

		uitem.prev = l.front
		uitem.next = nil
		if l.front != nil {
			l.front.next = uitem
		} else {
			l.back = uitem
		}
		l.front = uitem
	}
}

func (l *LinkedList) MoveToBack(e ListItem) {
	uitem, ok := e.(*LinkedItem)
	if ok && uitem != nil && uitem.parent == l && uitem != l.back {

		if uitem.prev != nil {
			uitem.prev.next = uitem.next
		}
		if uitem.next != nil {
			uitem.next.prev = uitem.prev
		}

		if uitem == l.front {
			l.front = uitem.prev
		}

		uitem.next = l.back
		uitem.prev = nil
		if l.back != nil {
			l.back.prev = uitem
		} else {
			l.front = uitem
		}
		l.back = uitem
	}
}

func (l *LinkedList) GetFront() ListItem {
	return l.front
}
func (l *LinkedList) GetBack() ListItem {
	return l.back
}

func (l *LinkedList) Remove(e ListItem) {
	uitem, ok := e.(*LinkedItem)
	if ok && uitem != nil && uitem.parent == l {
		if uitem != l.back && uitem != l.front {
			pr := uitem.prev
			nxt := uitem.next
			pr.next = nxt
			nxt.prev = pr
		}
		if uitem == l.front {
			pr := uitem.prev
			l.front = pr
		} // item can be both front and back simultaneously
		if uitem == l.back {
			nxt := uitem.next
			l.back = nxt
		}
		uitem.prev = nil
		uitem.next = nil
		l.stack = append(l.stack, uitem.idx)
	}
}

func (l *LinkedList) Clear() {
	l.front = nil
	l.back = nil
	l.stack = l.stack[:0]
	for i := range l.items {
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
	next          *LinkedItem
	prev          *LinkedItem
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
	val := l.next
	if val == nil || val.parent == nil || val.parent.back == val {
		return nil
	}
	return val
}

func (l *LinkedItem) Prev() ListItem {
	val := l.prev
	if val == nil || val.parent == nil || val.parent.front == val {
		return nil
	}
	return val
}

// val.parent.front == val
func (l *LinkedItem) IsNil() bool { // if interface is not nil but interface value is
	return l == nil
}
