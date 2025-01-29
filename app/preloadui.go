package main

import "github.com/rivo/tview"

var TextUI = []string{"Events", "Menu", "Current Room Actions", "Create Duo Room", "Create Group Room", "Unblock User",
	"Change Username", "Change Privacy", "Adding You To Group Rooms", "Adding You To Duo Rooms", "Find User", "Get Blocked Users", "Unblock User",
	"Block", "Leave Room", "Show users", "Delete Users From Room", "Add Users To Room", "Change Room Name", "Enter Text Here", "Add Users", "Delete Users"}
var NavEventText = [][5]int{
	// name entry in map, from nav, to nav, lists... (no more than 2 and first is target list (text from from:to is here) )
	{0, 0, 0, 4, 0},
	{1, 3, 9, 2, 0},
	{3, 3, 3, 9, 5},
	{4, 4, 4, 9, 8},
	{5, 11, 12, 9, 7},
	{6, 20, 6, 3, 0},
	{7, 8, 9, 9, 10},
	{10, 20, 10, 3, 0},
	{2, 13, 18, 2, 0},
}

type NavigateEvent struct {
	From  int // slicing NavText from -> to, like NavText[From:To]
	To    int
	Lists []tview.Primitive // lists for main *tview.Flex
}

func (c *Chat) InitUI() {
	//c.SendEventMap[]
	//c.Lists
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
