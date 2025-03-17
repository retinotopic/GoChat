package chat

import (
	"reflect"
	"strconv"
	"unicode"

	"github.com/gdamore/tcell/v2"
	// "github.com/mattn/go-isatty"
	"github.com/rivo/tview"

	"github.com/retinotopic/GoChat/app/list"
)

type Event struct {
	targets []int // slice of ints to Kind funcs
	content []list.Content
	Kind    func([]list.Content, ...int)
}

var SendEventNames = []string{"Add Users To Room", "Delete Users From Room",
	"Unblock User", "Block User", "Change Privacy Direct", "Change Privacy Group",
	"Change Username", "Find Users"}

var InitMapText = []string{
	"Events", "true", "4",
	"Menu", "true", "Create Duo Room", "true", "Create Group Room", "true",
	"Unblock User", "true", "Change Username", "true", "Change Privacy", "true",
	"Find User", "true", "2",
	"This Group Room(Admin)", "true", "Delete Users From Room", "true",
	"Add Users To Room", "true", "Change Room Name", "true", "Show Users", "true",
	"Leave Room", "true", "2",
	"This Group Room", "true", "Show Users", "true",
	"Leave Room", "true", "2",
	"This Duo Room", "true", "Block", "false", "2",
	"Create Duo Room", "true", "Create Duo Room", "false", "2", "5",
	"Create Group Room", "true", "Create Group Room", "false", "2", "6", "3",
	"Unblock Users", "true", "Unblock User", "false", "2", "7",
	"Add Users To Room", "true", "Add Users", "false", "2", "6",
	"Delete Users From Room", "true", "Delete Users", "false", "2", "6",
	"Change Room Name", "true", "Change Roomname", "false", "2", "3",
	"Change Username", "true", "Change Username", "false", "2", "3",
	"Find Users", "true", "Find Users", "false", "2", "5", "3",
	"Change Privacy", "true", "Change Duo Room Policy", "false",
	"Change Group Room Policy", "false", "2", "10",
	"Show Users", "true", "8",
	"Send Message", "false", "Message", "SendMessage", "-1",
	"Add Users", "false", "Room", "AddDeleteUsersInRoom", "6", "0",
	"Delete Users", "false", "Room", "AddDeleteUsersInRoom", "8", "1",
	"Unblock User", "false", "Room", "BlockUnblockUser", "7", "2",
	"Block", "false", "Room", "BlockUnblockUser", "-1", "3",
	"Change Duo Room Policy", "false", "User", "ChangePrivacy", "10", "4",
	"Change Group Room Policy", "false", "User", "ChangePrivacy", "10", "5",
	"Change Username", "false", "User", "ChangeUsernameFindUsers", "3", "6",
	"Find Users", "false", "User", "ChangeUsernameFindUsers", "3", "6",
	"Get Messages From Room", "false", "Message", "GetMessagesFromRoom", "-1",
	"Change Room Name", "false", "Room", "ChangeRoomname", "3",
	"Create Duo Room", "false", "Room", "CreateDuoRoom", "5",
	"Create Group Room", "false", "Room", "CreateGroupRoom", "6",
	"",
}

// parse InitMapText and fill c.EventMap
func (c *Chat) ParseAndInitUI() {

	SendEventKind := map[string]interface{}{
		"Room":    &Room{UserIds: make([]uint64, 0, 100), RoomIds: make([]uint64, 0, 100)},
		"User":    &User{},
		"Message": &Message{},
	}

	target := make([]int, 5)
	targetStr := make([]string, 20)
	c.EventMap = make(map[list.Content]Event)
	laststr := ""
	for _, v := range InitMapText {
		if !unicode.IsNumber([]rune(v)[0]) && unicode.IsNumber([]rune(laststr)[0]) {

			mode := targetStr[:2]
			targetStr = targetStr[2:]
			ev := Event{}
			key := list.Content{}

			if mode[1] == "true" {
				key.MainText = mode[0]
				for i, _ := range targetStr {
					if i%2 == 0 {
						evv := list.Content{}
						b, err := strconv.ParseBool(targetStr[i+1])
						if err != nil {
							panic("Parse Bool Error")
						}
						if b {
							evv.MainText = targetStr[i]
						} else {
							evv.SecondaryText = targetStr[i]
						}
						ev.content = append(ev.content, evv)
					}
				}
				ev.targets = target
				ev.Kind = c.EventUI
			} else {
				key.SecondaryText = mode[0]
				val := SendEventKind[targetStr[0]]
				raw := reflect.ValueOf(val).MethodByName(targetStr[1]).Interface()
				fn := raw.(func([]list.Content, ...int))
				ev.Kind = fn
				ev.targets = append(ev.targets, target...)
			}
			c.EventMap[key] = ev

		} else if unicode.IsNumber([]rune(v)[0]) {
			n, err := strconv.Atoi(v)
			if err != nil {
				panic("strconv Atoi error")
			}
			target = append(target, n)
		} else {
			targetStr = append(targetStr, v)
		}
		laststr = v
	}
}

func (c *Chat) EventUI(cnt []list.Content, trgt ...int) {
	lists := make([]tview.Primitive, 0, 5)
	c.Lists[trgt[0]].Items.Clear()
	ll, ok := c.Lists[trgt[0]].Items.(*list.ArrayList)
	if ok {
		for i := range cnt {
			a := list.ArrayItem{}
			a.ArrList = ll
			a.Color = [2]tcell.Color{tcell.ColorWhite, tcell.ColorWhite}

			a.MainText = cnt[i].MainText

			a.SecondaryText = cnt[i].SecondaryText

			c.Lists[trgt[0]].Items.MoveToFront(a)
		}
	}
	for i := range trgt {
		lists = append(lists, c.Lists[trgt[i]])
	}
	c.AddItemMainFlex(lists...)
}

type LoadingState struct {
	message         string
	spinner         []string
	color           string
	InProgressCount int
}

var state = LoadingState{
	message:         "In Progress",
	spinner:         []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	color:           "yellow",
	InProgressCount: 0,
}
