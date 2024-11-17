package main

import (
	"container/list"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LinkedListUI struct {
	*tview.Box
	offset       int // scroll offset
	SelectedFunc func(*list.Element)
	Items        *list.List
	Current      *list.Element // type Room
}

func (l *LinkedListUI) SetSelectedFunc(handler func(*list.Element)) *LinkedListUI {
	l.SelectedFunc = handler
	return l
}
func (l *LinkedListUI) MoveToFront(e *list.Element) *LinkedListUI {
	l.Items.MoveToFront(e)
	return l
}
func (l *LinkedListUI) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l)
	x, y, width, height := l.GetInnerRect()

	element := l.Items.Front()
	for i := 0; i < l.offset && element != nil; i++ {
		element = element.Next()
	}

	row := 0
	for element != nil && row < height {
		item := element.Value.(*Room)

		color := tcell.ColorWhite
		if element == l.Current {
			color = tcell.ColorYellow
		}

		tview.Print(screen, item.RoomName, x, y+row, width, tview.AlignLeft, color)

		if len(item.RoomType) > 0 && width > len(item.RoomName)+3 {
			secondaryX := x + len(item.RoomName) + 2
			tview.Print(screen, item.RoomType, secondaryX, y+row,
				width-len(item.RoomName)-2, tview.AlignLeft, tcell.ColorGray)
		}

		element = element.Next()
		row++
	}
}

func (l *LinkedListUI) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		_, _, _, height := l.GetInnerRect()

		switch event.Key() {
		case tcell.KeyUp:
			if l.Current != nil && l.Current.Prev() != nil {
				l.Current = l.Current.Prev()
				currentIndex := 0
				for e := l.Items.Front(); e != l.Current; e = e.Next() {
					currentIndex++
				}
				if currentIndex < l.offset {
					l.offset--
				}
			}
		case tcell.KeyDown:
			if l.Current != nil && l.Current.Next() != nil {
				l.Current = l.Current.Next()
				currentIndex := 0
				for e := l.Items.Front(); e != l.Current; e = e.Next() {
					currentIndex++
				}
				if currentIndex >= l.offset+height {
					l.offset++
				}
			}
		case tcell.KeyPgUp:
			for i := 0; i < height && l.Current != nil && l.Current.Prev() != nil; i++ {
				l.Current = l.Current.Prev()
				if l.offset > 0 {
					l.offset--
				}
			}
		case tcell.KeyPgDn:
			for i := 0; i < height && l.Current != nil && l.Current.Next() != nil; i++ {
				l.Current = l.Current.Next()
				if l.offset < l.Items.Len()-height {
					l.offset++
				}
			}
		case tcell.KeyEnter:
			l.SelectedFunc(l.Current)
		}
	})
}
