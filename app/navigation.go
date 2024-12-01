package main

import (
	"github.com/retinotopic/GoChat/app/list"
)

func (r *Chat) NavigationOptions(item list.ListItem) {
	switch item.GetMainText() {
	case "Event logs":
	case "Create Duo Room":
	case "Create Group Room":
	case "Unblock User":
	case "Change Username":
	case "Current Room Actions":
	case "Change Privacy for Duo Rooms":
	case "Change Privacy for Group Rooms":
	case "Add Users To Room":
	case "Delete Users From Room":
	case "Show users":
	case "Change Room Name":
	case "Block":
	case "Leave Room":
	case "Menu":
	}
}
