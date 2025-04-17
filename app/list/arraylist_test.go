package list_test

import (
	// "log"
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

	l1 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToFront(l1)

	l2 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToFront(l2)

	l3 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToFront(l3)

	if arr.GetFront().(list.ArrayItem).GetId() == l3.(list.ArrayItem).GetId() {
		t.Errorf("Index mismatch")
	}
	if arr.Len() != 4 {
		t.Errorf("Length mismatch")
	}

	arr.Clear()

	if arr.Len() != 0 {
		t.Errorf("Length mismatch")
	}
	// log.Println(arr.GetFront())
	AssertItems(arr.GetFront(), nil, t)

	AssertItems(l0.Next(), nil, t)

	AssertItems(l3.Prev(), nil, t)

}
