package e2e_test

import (
	"os"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app"
	"github.com/retinotopic/GoChat/app/list"
)

func E2E_test(t *testing.T) {
	user1, err := app.NewChat("CoolUser1", os.Getenv("CHAT_CONNECT_ADDRESS"), 1)
	if err != nil {
		t.Error(err)
	}
	_, err = app.NewChat("WonderfulUser2", os.Getenv("CHAT_CONNECT_ADDRESS"), 1)

	if err != nil {
		t.Error(err)
	}
	_, err = app.NewChat("AmazingUser3", os.Getenv("CHAT_CONNECT_ADDRESS"), 1)

	if err != nil {
		t.Error(err)
	}

	// e := user1.EventMap[list.Content{SecondaryText: "Find Users"}]
	// user1.OptionEvent()
	l := list.NewArrayList(1) // mock list for one single item
	item := l.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "", "")
	var li list.ListItem = item

	user1.Lists[3].Items.MoveToFront(list.ArrayItem{MainText: "WonderfulUser2"})
	li.SetSecondaryText("Find Users")
	user1.OptionEvent(li)
}
