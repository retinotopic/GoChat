package app

import (
	"reflect"
	"strconv"
	"sync/atomic"
	"unicode"

	"golang.org/x/sync/errgroup"

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
	"Unblock Users", "true", "Change Username", "true", "Change Privacy", "true",
	"Find Users", "true", "2",
	"This Group Room(Admin)", "true", "Delete Users From Room", "true",
	"Add Users To Room", "true", "Change Room Name", "true", "Show Users", "true",
	"Leave Room", "true", "2",
	"This Group Room", "true", "Show Users", "true",
	"Leave Room", "true", "2",
	"This Duo Room", "true", "Block", "false", "2",
	"Create Duo Room", "true", "Create Duo Room", "false", "2", "5",
	"Create Group Room", "true", "Create Group Room", "false", "2", "6", "3",
	"Unblock Users", "true", "Unblock User", "false", "Get Blocked Users", "false", "2", "7",
	"Block Users", "true", "Block User", "false", "2", "6",
	"Add Users To Room", "true", "Add Users", "false", "2", "6",
	"Delete Users From Room", "true", "Delete Users", "false", "2", "6",
	"Change Room Name", "true", "Change Roomname", "false", "2", "3",
	"Change Username", "true", "Change Username", "false", "2", "3",
	"Find Users", "true", "Find Users", "false", "2", "5", "3",
	"Change Privacy", "true", "Change Duo Room Policy", "false",
	"Change Group Room Policy", "false", "2", "9",
	"Show Users", "true", "8",
	"Send Message", "false", "Message", "SendMessage", "-1",
	"Add Users", "false", "Room", "AddDeleteUsersInRoom", "6", "0",
	"Delete Users", "false", "Room", "AddDeleteUsersInRoom", "8", "1",
	"Get Blocked Users", "false", "User", "GetBlockedUsers", "-1",
	"Unblock User", "false", "Room", "BlockUnblockUser", "7", "2",
	"Block User", "false", "Room", "BlockUnblockUser", "6", "3",
	"Change Duo Room Policy", "false", "User", "ChangePrivacy", "9", "4",
	"Change Group Room Policy", "false", "User", "ChangePrivacy", "9", "5",
	"Change Username", "false", "User", "ChangeUsernameFindUsers", "-1", "6",
	"Find Users", "false", "User", "ChangeUsernameFindUsers", "-1", "7",
	"Get Messages From Room", "false", "Message", "GetMessagesFromRoom", "-1",
	"Change Room Name", "false", "Room", "ChangeRoomName", "-1",
	"Create Duo Room", "false", "Room", "CreateDuoRoom", "5",
	"Create Group Room", "false", "Room", "CreateGroupRoom", "6",
	"end",
}

// parse InitMapText and fill c.EventMap
func (c *Chat) ParseAndInitUI() {
	c.SendEventCh = make(chan EventInfo, 100)
	c.errch = make(chan error, 2)
	c.errgroup = errgroup.Group{}

	SendEventKind := map[string]any{
		"Room":    &Room{SendCh: c.SendEventCh},
		"User":    &User{SendCh: c.SendEventCh},
		"Message": &Message{SendCh: c.SendEventCh},
	}
	c.BlockedUsers = make(map[uint64]User)
	c.DuoUsers = make(map[uint64]User)
	c.FoundUsers = make(map[uint64]User)
	c.RoomMsgs = make(map[uint64]*RoomInfo)
	target := make([]int, 0, 100)
	targetStr := make([]string, 0, 100)
	c.EventMap = make(map[list.Content]Event)
	laststr := " "
	for _, v := range InitMapText {
		vch := []rune(v)
		laststrch := []rune(laststr)
		if !unicode.IsNumber(vch[len(vch)-1]) && unicode.IsNumber(laststrch[len(laststrch)-1]) {
			mode := targetStr[:2]
			targetstr := targetStr[2:]
			ev := Event{}
			key := list.Content{}

			if mode[1] == "true" {
				key.MainText = mode[0]
				for i, _ := range targetstr {
					if i%2 == 0 {
						evv := list.Content{}
						b, err := strconv.ParseBool(targetstr[i+1])
						if err != nil {
							panic("Parse Bool Error")
						}
						if b {
							evv.MainText = targetstr[i]
						} else {
							evv.SecondaryText = targetstr[i]
						}
						ev.content = append(ev.content, evv)
					}
				}
				ev.targets = append(ev.targets, target...)
				ev.Kind = c.EventUI
			} else {
				key.SecondaryText = mode[0]
				val, ok := SendEventKind[targetstr[0]]
				if !ok {
					panic("this shouldnt happen")
				}
				raw := reflect.ValueOf(val).MethodByName(targetstr[1]).Interface()
				fn := raw.(func([]list.Content, ...int))
				ev.Kind = fn
				ev.targets = append(ev.targets, target...)
			}
			c.EventMap[key] = ev
			targetStr = targetStr[:0]
			target = target[:0]
			target, targetStr = c.isNumber(v, target, targetStr)

		} else {
			target, targetStr = c.isNumber(v, target, targetStr)
		}
		laststr = v
	}
	c.state = LoadingState{
		message: "In Progress",
		spinner: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		color:   "yellow",
	}
	c.state.InProgressCount.Store(0)
}

func (c *Chat) isNumber(v string, target []int, targetStr []string) ([]int, []string) {
	vlen := len([]rune(v))
	if unicode.IsNumber([]rune(v)[vlen-1]) {
		n, err := strconv.Atoi(v)
		if err != nil {
			panic("strconv Atoi error")
		}
		target = append(target, n)
	} else {
		targetStr = append(targetStr, v)
	}
	return target, targetStr
}

type LoadingState struct {
	message         string
	spinner         []string
	color           string
	InProgressCount atomic.Int32
}
