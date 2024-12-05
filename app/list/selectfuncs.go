package list

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type SelectHandler struct {
	MultOpts     map[string]struct{}
	OneOpt       string
	mtx          sync.RWMutex
	DefaultColor tcell.Color
	ChangeColor  tcell.Color
}

func (h *SelectHandler) Clear() {
	clear(h.MultOpts)
	h.OneOpt = ""
}
func (h *SelectHandler) MultOptions(item ListItem) {
	id := item.GetSecondaryText()
	h.mtx.RLock()
	_, ok := h.MultOpts[id]
	h.mtx.RUnlock()
	if ok {
		item.SetColor(h.DefaultColor)
		h.mtx.Lock()
		delete(h.MultOpts, id)
		h.mtx.Unlock()
	} else {
		item.SetColor(h.ChangeColor)
		h.mtx.Lock()
		h.MultOpts[id] = struct{}{}
		h.mtx.Unlock()
	}
}

func (h *SelectHandler) OneOption(item ListItem) {
	item.SetColor(h.ChangeColor)
	h.OneOpt = item.GetSecondaryText()
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
				bi.SetColor(h.DefaultColor)
			}
		}
		if forward {
			fi = fi.Next()
			if fi == nil || fi.IsNil() {
				forward = false
			} else {
				fi.SetColor(h.DefaultColor)
			}
		}
	}
}
