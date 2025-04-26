package list_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/retinotopic/GoChat/app/list"
)

func Test_ArrayList(t *testing.T) {
	li := list.List{}
	listLength := 30
	arr := list.NewArrayList(listLength)
	li.Items = arr

	clrs := [2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}

	l0 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToBack(l0)
	l0 = arr.GetFront()

	l1 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToBack(l1)
	l1 = arr.GetFront()

	l2 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToBack(l2)
	l2 = arr.GetFront()

	l3 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToBack(l3)
	l3 = arr.GetFront()

	arrcheck := []list.ListItem{l0, l1, l2, l3}
	item := arr.GetBack()
	i := 0
	for item != nil && !item.IsNil() {
		if item != arrcheck[i] {
			t.Errorf("consistency list validation failed")
		}
		item = item.Next()
		i += 1
	}
	if arr.Len() != 4 {
		t.Errorf("Length mismatch")
	}

	arr.Clear()

	if arr.Len() != 0 {
		t.Errorf("Length mismatch")
	}
	AssertItems(arr.GetFront(), nil, t)

	AssertItems(l0.Next(), nil, t)

	AssertItems(l3.Prev(), nil, t)

}
