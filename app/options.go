package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
)

func (c *Chat) MultOption(item list.ListItem) {
	if item == nil {
		return
	}
	color := item.GetColor(0)

	if color == tcell.ColorRed {
		item.SetColor(tcell.ColorWhite, 0)
	} else {
		item.SetColor(tcell.ColorRed, 0)
	}
}

func (c *Chat) OneOption(item list.ListItem) {
	if item == nil {
		return
	}

	item.SetColor(tcell.ColorRed, 0)
	bi := item
	fi := item
	back := true
	forward := true
	for back || forward {
		if back {
			bi = bi.Prev()
			if bi == nil || bi.IsNil() {
				back = false
			} else {
				bi.SetColor(tcell.ColorWhite, 0)
			}
		}
		if forward {
			fi = fi.Next()
			if fi == nil || fi.IsNil() {
				forward = false
			} else {
				fi.SetColor(tcell.ColorWhite, 0)
			}
		}
	}
}
