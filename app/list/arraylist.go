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

func (a *ArrayList) MoveToBack(e ListItem) { // for arrays is the same as movetofront
	a.MoveToFront(e)
}
func (a *ArrayList) MoveToFront(e ListItem) {
	uitem, ok := e.(ArrayItem)
	uitem.ArrList = a
	if ok {
		uitem.Idx = len(a.Items)
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
	if len(a.Items) != 0 {
		a.Items[len(a.Items)-1].ArrList = nil
		a.Items = a.Items[:len(a.Items)-1] // removing the last element. argument is unused, only to satisfy the interface
	}
}
func (a *ArrayList) Clear() {
	for i := range a.Items {
		a.Items[i].ArrList = nil
	}
	a.Items = a.Items[:0]
}
func (a *ArrayList) Len() int {
	return len(a.Items)
}

type ArrayItem struct {
	ArrList       *ArrayList
	Idx           int
	Color         [2]tcell.Color
	MainText      string
	MainTextBuf   *bytes.Buffer
	SecondaryText string
}

func (a *ArrayList) NewItem(clr [2]tcell.Color, main string, sec string) ListItem {
	return ArrayItem{
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
func (a ArrayItem) GetColor(Idx int) tcell.Color {
	if Idx < 2 && Idx >= 0 {
		return a.ArrList.Items[a.Idx].Color[Idx]

	}
	return tcell.ColorWhite
}
func (a ArrayItem) SetMainText(str string, mode uint8) {
	a.ArrList.Items[a.Idx].MainText = str
}
func (a ArrayItem) SetSecondaryText(str string) {
	a.ArrList.Items[a.Idx].SecondaryText = str
}
func (a ArrayItem) SetColor(clr tcell.Color, Idx int) {
	if Idx < 2 && Idx >= 0 {
		a.ArrList.Items[a.Idx].Color[Idx] = clr
	}
}
func (a ArrayItem) Next() ListItem {
	if a.Idx+1 < len(a.ArrList.Items) {
		return a.ArrList.Items[a.Idx+1]
	}
	return nil
}

func (a ArrayItem) Prev() ListItem {
	if a.Idx-1 >= 0 && a.Idx-1 < len(a.ArrList.Items) {
		return a.ArrList.Items[a.Idx-1]
	}
	return nil
}
func (a ArrayItem) IsNil() bool {
	return false
}
