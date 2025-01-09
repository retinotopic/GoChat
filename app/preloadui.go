package main

var NavText = []string{"Menu", "Create Duo Room", "Create Group Room", "Unblock User", "Change Username", "Current Room Actions", "Change Privacy",
	"Current Room Actions", "Block", "Leave Room", "Show users", "Add Users To Room", "Delete Users From Room", "Change Room Name",
	"Events", "Change Privacy", "for Duo Rooms", "for Group Rooms", "Unblock User", "Get Blocked Users", "Unblock User"}

var NavigateEventtMap map[string]NavigateEvent

type NavigateEvent struct {
	From  int // slicing NavText from -> to, like NavText[From:To]
	To    int
	Lists []int // lists for main *tview.Flex
}

var SendEventMap map[string]SendEvent

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
