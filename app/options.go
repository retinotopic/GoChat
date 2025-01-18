package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
)

func MultOption(item list.ListItem) {
	color := item.GetColor(0)

	if color == tcell.ColorYellow {
		item.SetColor(tcell.ColorWhite, 0)
	} else {
		item.SetColor(tcell.ColorYellow, 0)
	}
}

func OneOption(item list.ListItem) {
	item.SetColor(tcell.ColorYellow, 0)
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
