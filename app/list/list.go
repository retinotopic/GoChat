package list

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type List struct {
	*tview.Box
	offset       int // scroll offset
	SelectedFunc func(ListItem)
	Items        ListItems
	Current      ListItem // type Room
}
type ListItem interface {
	GetMainText() string
	GetSecondaryText() string
	GetColor() tcell.Color
	SetMainText(string)
	SetSecondaryText(string)
	SetColor(tcell.Color)
	Next() ListItem
	Prev() ListItem
	IsNil() bool
}
type ListItems interface {
	MoveToFront(ListItem)
	MoveToBack(ListItem)
	GetFront() ListItem
	Remove(ListItem)
	Clear()
}

func (l *List) SetSelectedFunc(handler func(ListItem)) *List {
	l.SelectedFunc = handler
	return l
}

func (l *List) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l)
	x, y, width, height := l.GetInnerRect()

	element := l.Items.GetFront()
	for i := 0; i < l.offset && element != nil && !element.IsNil(); i++ {
		element = element.Next()
	}

	row := 0
	for element != nil && !element.IsNil() && row < height {

		tview.Print(screen, element.GetMainText(), x, y+row, width, tview.AlignLeft, element.GetColor())

		if len(element.GetSecondaryText()) > 0 && width > len(element.GetMainText())+3 {
			secondaryX := x + len(element.GetMainText()) + 2
			tview.Print(screen, element.GetSecondaryText(), secondaryX, y+row,
				width-len(element.GetMainText())-2, tview.AlignLeft, tcell.ColorGray)
		}

		element = element.Next()
		row++
	}
}

func (l *List) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		_, _, _, height := l.GetInnerRect()

		switch event.Key() {
		case tcell.KeyUp:
			if l.Current != nil && l.Current.Prev() != nil && !l.Current.Prev().IsNil() {
				l.Current = l.Current.Prev()
				currentIndex := 0
				for e := l.Items.GetFront(); e != l.Current; e = e.Next() {
					currentIndex++
				}
				if currentIndex < l.offset {
					l.offset--
				}
			}
		case tcell.KeyDown:
			if l.Current != nil && l.Current.Next() != nil && !l.Current.Next().IsNil() {
				l.Current = l.Current.Next()
				currentIndex := 0
				for e := l.Items.GetFront(); e != l.Current; e = e.Next() {
					currentIndex++
				}
				if currentIndex >= l.offset+height {
					l.offset++
				}
			}
		case tcell.KeyEnter:
			l.SelectedFunc(l.Current)
		}
	})
}
