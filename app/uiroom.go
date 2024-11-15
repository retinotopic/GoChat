package main

import (
	"container/list"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (r *Chat) SetSelectedFunc(handler func(*list.Element)) *Chat {
	r.selectedFunc = handler
	return r
}

func (r *Chat) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()

	element := r.items.Front()
	for i := 0; i < r.offset && element != nil; i++ {
		element = element.Next()
	}

	row := 0
	for element != nil && row < height {
		item := element.Value.(*Room)

		color := tcell.ColorWhite
		if element == r.current {
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

func (r *Chat) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		_, _, _, height := r.GetInnerRect()

		switch event.Key() {
		case tcell.KeyUp:
			if r.current != nil && r.current.Prev() != nil {
				r.current = r.current.Prev()
				currentIndex := 0
				for e := r.items.Front(); e != r.current; e = e.Next() {
					currentIndex++
				}
				if currentIndex < r.offset {
					r.offset--
				}
			}
		case tcell.KeyDown:
			if r.current != nil && r.current.Next() != nil {
				r.current = r.current.Next()
				currentIndex := 0
				for e := r.items.Front(); e != r.current; e = e.Next() {
					currentIndex++
				}
				if currentIndex >= r.offset+height {
					r.offset++
				}
			}
		case tcell.KeyPgUp:
			for i := 0; i < height && r.current != nil && r.current.Prev() != nil; i++ {
				r.current = r.current.Prev()
				if r.offset > 0 {
					r.offset--
				}
			}
		case tcell.KeyPgDn:
			for i := 0; i < height && r.current != nil && r.current.Next() != nil; i++ {
				r.current = r.current.Next()
				if r.offset < r.items.Len()-height {
					r.offset++
				}
			}
		case tcell.KeyEnter:
			r.selectedFunc(r.current)
		}
	})
}
