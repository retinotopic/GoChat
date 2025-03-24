package list

import (
	"bytes"

	"github.com/gdamore/tcell/v2"
)

func NewArrayList(length int) *ArrayList {
	return &ArrayList{Items: make([]ArrayItem, 0, length)}
}

type ArrayList struct {
	Items  []ArrayItem
	Length int
}

func (a *ArrayList) MoveToBack(e ListItem) {
	uitem, ok := e.(ArrayItem)
	uitem.ArrList = a
	if ok {
		uitem.idx = 0
		if len(a.Items) == 0 {
			a.Items = append(a.Items, uitem)
			return
		}
		a.Items[0] = uitem
	}
}
func (a *ArrayList) MoveToFront(e ListItem) {
	uitem, ok := e.(ArrayItem)
	uitem.ArrList = a
	if ok {
		uitem.idx = len(a.Items)
		a.Items = append(a.Items, uitem)
	}
}
func (a *ArrayList) GetFront() ListItem {
	if len(a.Items) == 0 {
		return nil
	}
	return a.Items[len(a.Items)-1]
}
func (a *ArrayList) Remove(e ListItem) {
	a.Items = a.Items[:len(a.Items)-1] // removing the last element. argument is unused, only to satisfy the interface
}
func (a *ArrayList) Clear() {
	a.Items = a.Items[:0]
}
func (a *ArrayList) Len() int {
	return len(a.Items)
}

type ArrayItem struct {
	ArrList       *ArrayList
	idx           int
	Color         [2]tcell.Color
	MainText      string
	MainTextBuf   *bytes.Buffer
	SecondaryText string
}

func (a *ArrayList) NewItem(clr [2]tcell.Color, main string, sec string) ListItem {
	return ArrayItem{
		ArrList:       a,
		Color:         clr,
		MainText:      main,
		MainTextBuf:   nil,
		SecondaryText: sec,
	}
}
func (a ArrayItem) GetParent() ListItems {
	return a.ArrList
}
func (a ArrayItem) GetMainText() string {
	if a.MainTextBuf != nil {
		return a.MainTextBuf.String()
	}
	return a.MainText
}
func (a ArrayItem) GetSecondaryText() string {
	return a.SecondaryText
}
func (a ArrayItem) GetColor(idx int) tcell.Color {
	if idx < 2 && idx >= 0 {
		return a.ArrList.Items[a.idx].Color[idx]

	}
	return tcell.ColorWhite
}
func (a ArrayItem) SetMainText(str string, mode uint8) {
	switch mode {
	case 0:
		a.ArrList.Items[a.idx].MainText = str
	case 1:
		if a.MainTextBuf != nil {
			a.MainTextBuf.WriteString(str)
			return
		}
		a.MainTextBuf = bytes.NewBufferString(str)
	case 2:
		if a.MainTextBuf != nil {
			a.MainTextBuf.Truncate(len(a.MainTextBuf.String()) - len(str))
		}
	}
}
func (a ArrayItem) SetSecondaryText(str string) {
	a.ArrList.Items[a.idx].SecondaryText = str
}
func (a ArrayItem) SetColor(clr tcell.Color, idx int) {
	if idx < 2 && idx >= 0 {
		a.ArrList.Items[a.idx].Color[idx] = clr
	}
}
func (a ArrayItem) Next() ListItem {
	if a.idx+1 > len(a.ArrList.Items)-1 || a.idx+1 < 0 {
		return nil
	}
	return a.ArrList.Items[a.idx+1]
}

func (a ArrayItem) Prev() ListItem {
	if a.idx-1 < 0 || a.idx-1 > len(a.ArrList.Items)-1 {
		return nil
	}
	return a.ArrList.Items[a.idx-1]
}
func (a ArrayItem) IsNil() bool {
	return false
}
func (a ArrayItem) GetId() int {
	return a.idx
}
