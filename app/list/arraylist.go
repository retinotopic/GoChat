package list

import (
	"bytes"

	"github.com/gdamore/tcell/v2"
)

func NewArrayList() *ArrayList {
	return &ArrayList{Items: make([]ArrayItem, 0, 40)}
}

type ArrayList struct {
	Items        []ArrayItem
	CurrentIndex int
	Length       int
}

func (a *ArrayList) MoveToFront(e ListItem) {
	uitem, ok := e.(ArrayItem)
	uitem.ArrList = a
	if ok {
		a.Items[0] = uitem
	}
}
func (a *ArrayList) MoveToBack(e ListItem) {
	uitem, ok := e.(ArrayItem)
	if ok {
		a.Items = append(a.Items, uitem)
	}
}
func (a *ArrayList) GetFront() ListItem {
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
	Color         [2]tcell.Color
	MainText      string
	MainTextBuf   *bytes.Buffer
	SecondaryText string
}

func NewArrayItem(arr *ArrayList, clr [2]tcell.Color, main string, sec string) ArrayItem {
	return ArrayItem{
		ArrList:       arr,
		Color:         clr,
		MainText:      main,
		MainTextBuf:   nil,
		SecondaryText: sec,
	}
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
		return a.ArrList.Items[a.ArrList.CurrentIndex].Color[idx]

	}
	return tcell.ColorWhite
}
func (a ArrayItem) SetMainText(str string, mode uint8) {
	switch mode {
	case 0:
		a.ArrList.Items[a.ArrList.CurrentIndex].MainText = str
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
	a.ArrList.Items[a.ArrList.CurrentIndex].SecondaryText = str
}
func (a ArrayItem) SetColor(clr tcell.Color, idx int) {
	if idx < 2 && idx >= 0 {
		a.ArrList.Items[a.ArrList.CurrentIndex].Color[idx] = clr
	}
}
func (a ArrayItem) Next() ListItem {
	if a.ArrList.CurrentIndex == len(a.ArrList.Items)-1 {
		return nil
	}
	a.ArrList.CurrentIndex++
	return a.ArrList.Items[a.ArrList.CurrentIndex]
}
func (a ArrayItem) Prev() ListItem {
	if a.ArrList.CurrentIndex == 0 {
		return nil
	}
	a.ArrList.CurrentIndex--
	return a.ArrList.Items[a.ArrList.CurrentIndex]
}
func (a ArrayItem) IsNil() bool {
	return false
}
