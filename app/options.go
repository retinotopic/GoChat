package main

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app/list"
)

type MultOptions struct {
	MultOpts     map[string]struct{}
	mtx          sync.RWMutex
	DefaultColor tcell.Color
	ChangeColor  tcell.Color
}

func (m *MultOptions) Option(item list.ListItem) {
	id := item.GetSecondaryText()
	m.mtx.RLock()
	_, ok := m.MultOpts[id]
	m.mtx.RUnlock()
	if ok {
		item.SetColor(m.DefaultColor)
		m.mtx.Lock()
		delete(m.MultOpts, id)
		m.mtx.Unlock()
	} else {
		item.SetColor(m.ChangeColor)
		m.mtx.Lock()
		m.MultOpts[id] = struct{}{}
		m.mtx.Unlock()
	}
}

func (m *MultOptions) Clear() {
	clear(m.MultOpts)
}
func (m *MultOptions) GetItems() []string {
	sli := make([]string, len(m.MultOpts))
	i := 0
	for s := range m.MultOpts {
		sli[i] = s
		i++
	}
	return sli
}

type OneOption struct {
	OneOpt       string
	DefaultColor tcell.Color
	ChangeColor  tcell.Color
}

func (o *OneOption) Option(item list.ListItem) {
	item.SetColor(o.ChangeColor)
	o.OneOpt = item.GetSecondaryText()
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
				bi.SetColor(o.DefaultColor)
			}
		}
		if forward {
			fi = fi.Next()
			if fi == nil || fi.IsNil() {
				forward = false
			} else {
				fi.SetColor(o.DefaultColor)
			}
		}
	}
}
func (o *OneOption) Clear() {
	o.OneOpt = ""
}
func (o *OneOption) GetItems() []string {
	return []string{o.OneOpt}
}
