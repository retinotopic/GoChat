package list

import (
	"github.com/gdamore/tcell/v2"
)

func NewUnrolledList() *ArrayList {
	return &ArrayList{Items: make([]ArrayItem, 0, 40)}
}

type ArrayList struct {
	Items        []ArrayItem
	CurrentIndex int
	Length       int
}

func (a *ArrayList) MoveToFront(e ListItem) {
	uitem, ok := e.(ArrayItem)
	if ok {
		a.Items[0] = uitem
	}
}
func (a *ArrayList) MoveToBack(e ListItem) {
	uitem, ok := e.(ArrayItem)
	if ok {
		a.Items[len(a.Items)-1] = uitem
	}
}
func (a *ArrayList) GetFront() ListItem {
	return a.Items[0]
}
func (a *ArrayList) Remove(e ListItem) {
	a.Items = a.Items[:len(a.Items)-1] // removing the last element. argument is unused, only to satisfy the interface lol
}
func (a *ArrayList) Clear() {
	a.Items = a.Items[:0]
}
func (a *ArrayList) Len() int {
	return len(a.Items)
}

type ArrayItem struct {
	ArrList       *ArrayList
	Color         tcell.Color
	MainText      string
	SecondaryText string
}

func (a ArrayItem) GetMainText() string {
	return a.MainText
}
func (a ArrayItem) GetSecondaryText() string {
	return a.SecondaryText
}
func (a ArrayItem) GetColor() tcell.Color {
	return a.Color
}
func (a ArrayItem) SetMainText(str string) {
	a.ArrList.Items[a.ArrList.CurrentIndex].MainText = str
}
func (a ArrayItem) SetSecondaryText(str string) {
	a.ArrList.Items[a.ArrList.CurrentIndex].SecondaryText = str
}
func (a ArrayItem) SetColor(clr tcell.Color) {
	a.ArrList.Items[a.ArrList.CurrentIndex].Color = clr
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
