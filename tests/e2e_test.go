package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/retinotopic/GoChat/app"
	"github.com/retinotopic/GoChat/app/list"
)

func E2E_test2(t *testing.T) {
	user1, err := app.NewChat("CoolUser1", os.Getenv("CHAT_CONNECT_ADDRESS"), 1)
	if err != nil {
		t.Fatal(err)
	}
	user2, err := app.NewChat("WonderfulUser2", os.Getenv("CHAT_CONNECT_ADDRESS"), 1)
	if err != nil {
		t.Fatal(err)
	}
	user3, err := app.NewChat("AmazingUser3", os.Getenv("CHAT_CONNECT_ADDRESS"), 1)
	if err != nil {
		t.Fatal(err)
	}
	menuevent := list.NewList()
	menuevent.Items = list.NewArrayList(1)
	menuevent.Items.NewItem([2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}, "", "Finds Users")

	user1.Lists[3].Items.GetFront().SetMainText("WonderfulUser2", 0)

	user1.OptionEvent(menuevent.Items.GetFront())
	// WaitForCompletion(context.WithTimeout(context.Background(),time.Second *7,user1,,t)

	found := user1.Lists[5].Items.GetFront()

}
func WaitForCompletion(ctx context.Context, user *app.Chat, listidx int,
	recentEventsList *list.List, HasToFail bool, t *testing.T) {

	eventcounter := user.Lists[listidx].Items.Len()
	go func() {
		user1.OptionEvent(menuevent.Items.GetFront())
	}()
	done := ctx.Done()
	for {
		select {
		case <-done:
			t.Error("timeout error")
		}
		if eventcounter < recentEventsList.Items.Len() {
			if len(recentEventsList.Items.GetFront().GetSecondaryText()) != 0 && HasToFail != true {
				t.Error("backend error")
			}
			return
		}
	}
}
