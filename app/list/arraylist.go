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
	idx    int
	Length int
}

func (a *ArrayList) MoveToBack(e ListItem) { // for arrays is the same as movetofront
	a.MoveToFront(e)
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
func (a *ArrayList) GetBack() ListItem {
	if len(a.Items) == 0 {
		return nil
	}
	return a.Items[0]
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
		idx:           len(a.Items),
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
	if a.ArrList.idx+1 < len(a.ArrList.Items) {
		a.ArrList.idx += 1
		return a.ArrList.Items[a.ArrList.idx]
	}
	return nil
}

func (a ArrayItem) Prev() ListItem {
	if a.ArrList.idx-1 >= 0 {
		a.ArrList.idx -= 1
		return a.ArrList.Items[a.ArrList.idx]
	}
	return nil
}
func (a ArrayItem) IsNil() bool {
	return false
}
