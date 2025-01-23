package main

import "github.com/rivo/tview"

var NavText = []string{"Events", "Menu", "Current Room Actions", "Create Duo Room", "Create Group Room", "Unblock User",
	"Change Username", "Change Privacy Direct", "Change Privacy Group", "Find User", "Get Blocked Users", "Unblock User",
	"Block", "Leave Room", "Show users", "Add Users To Room", "Delete Users From Room", "Change Room Name"}
var NavEventText = [][6]int{
	// name entry in map, from nav, to nav,input msg, lists... (no more than 2)
	{0, 0, 0, 0, 4, 0},
	{1, 2, 9, 0, 2, 0},
	{3, 3, 3, 0, 2, 5},
	{4, 4, 4, 0, 2, 8},
	{5, 5, 5, 0, 2, 7},
	{6, 6, 6, 6, 2, 3},
}

type NavigateEvent struct {
	From     int // slicing NavText from -> to, like NavText[From:To]
	To       int
	InputMsg string
	Lists    []tview.Primitive // lists for main *tview.Flex
}

func (c *Chat) InitUI() {

}

type SendEvent struct {
	ListIdx int
	Event   func([]string)
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
