package list_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/retinotopic/GoChat/app/list"
)

func Test_LinkedList(t *testing.T) {
	li := list.List{}
	listLength := 30
	arr := list.NewLinkedList(listLength)
	li.Items = arr

	clrs := [2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}

	l0 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToBack(l0)

	l1 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToFront(l1)

	AssertItems(l1.Prev(), l0, t)
	AssertItems(l0.Next(), l1, t)

	l2 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToFront(l2)

	l3 := arr.NewItem(clrs, "SampleText", "SampleText")
	arr.MoveToBack(l3)

	AssertItems(li.Items.GetFront(), l2, t)

	AssertItems(l0.Prev(), l3, t)
	AssertItems(l3.Next(), l0, t)

	arr.MoveToFront(l3)

	AssertItems(li.Items.GetFront(), l3, t)

	arr.Remove(l3)

	if arr.Len() != 3 {
		t.Errorf("Length mismatch")
	}

	AssertItems(li.Items.GetFront(), l2, t)

	arr.Remove(l2)

	AssertItems(li.Items.GetFront(), l1, t)

	arr.Remove(l1)

	AssertItems(li.Items.GetFront(), l0, t)

	if arr.Len() != 1 {
		t.Errorf("Length mismatch")
	}
	arr.Clear()

	if arr.Len() != 0 {
		t.Errorf("Length mismatch")
	}

}

func AssertItems(got list.ListItem, want list.ListItem, t *testing.T) {
	if got != want {
		t.Errorf("Assertion failed, got: %v, want: %v ", got, want)
	}
}
