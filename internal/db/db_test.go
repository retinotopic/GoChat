package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/retinotopic/GoChat/internal/db"
)

func TestDb(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// creating user 1
	err := db.NewUser(ctx, "user1sub", "user1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	client1, err := db.NewClient(ctx, "user1sub")
	if err != nil {
		t.Fatalf("%v", err)
	}
	// creating user 2
	err = db.NewUser(ctx, "user2sub", "user2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := db.NewClient(ctx, "user2sub")
	if err != nil {
		t.Fatalf("%v", err)
	}
	// creating user 3
	err = db.NewUser(ctx, "user3sub", "user3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := db.NewClient(ctx, "user3sub")
	if err != nil {
		t.Fatalf("%v", err)
	}
	// user 1 finds user 2
	flowjson := &db.FlowJSON{
		Mode:  "FindUsers",
		Name:  "user2",
		Users: make([]uint32, 10),
	}
	go client1.TxManage(flowjson)
	ch1 := client1.Channel()
	for f := range ch1 {
		*flowjson = f
		break
	}
	client1.ClearChannel()
	got := flowjson.Users[0]
	want := client2.UserID
	if got != want {
		t.Fatalf("%v %v %v assertion failed", got, want, flowjson)
	}
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 1 add to duoroom user 2
	flowjson = &db.FlowJSON{
		Mode:  "CreateDuoRoom",
		Users: []uint32{client1.UserID, flowjson.Users[0]},
	}
	client1.ClearChannel()
	client1.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 2 blocks user 1
	flowjson = &db.FlowJSON{
		Mode:  "BlockUser",
		Users: []uint32{client2.UserID, client1.UserID},
	}
	client2.ClearChannel()
	client2.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 1 tries to add user 2 to duoroom
	flowjson = &db.FlowJSON{
		Mode:  "CreateDuoRoom",
		Users: []uint32{client1.UserID, client2.UserID},
	}
	client1.ClearChannel()
	client1.TxManage(flowjson)
	if flowjson.Err == nil {
		t.Fatalf("%v Expected error, got nil", flowjson)
	}

	// user 2 unblock user 1
	flowjson = &db.FlowJSON{
		Mode:  "UnblockUser",
		Users: []uint32{client2.UserID, client1.UserID},
	}
	client2.ClearChannel()
	client2.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 1 add to duoroom user 2
	flowjson = &db.FlowJSON{
		Mode:  "CreateDuoRoom",
		Users: []uint32{client1.UserID, client2.UserID},
	}
	client1.ClearChannel()
	client1.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 3 privacy settings: group and direct = false
	if err != nil {
		t.Fatalf("%v", err)
	}
	flowjson = &db.FlowJSON{
		Mode: "ChangePrivacyDirect",
		Bool: false,
	}
	client3.ClearChannel()
	client3.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}
	flowjson = &db.FlowJSON{
		Mode: "ChangePrivacyGroup",
		Bool: false,
	}
	client3.ClearChannel()
	client3.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 1 finds user 3
	flowjson = &db.FlowJSON{
		Mode:  "FindUsers",
		Name:  "user3",
		Users: make([]uint32, 10),
	}
	client1.ClearChannel()
	go client1.TxManage(flowjson)
	for f := range ch1 {
		*flowjson = f
		break
	}
	got = flowjson.Users[0]
	want = client3.UserID
	if got != want {
		t.Fatalf("%v %v %v %s assertion failed", got, want, flowjson.Users, flowjson.Name)
	}
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 2 finds user 3
	flowjson = &db.FlowJSON{
		Mode:  "FindUsers",
		Name:  "user3",
		Users: make([]uint32, 10),
	}
	client2.ClearChannel()
	go client2.TxManage(flowjson)
	ch2 := client2.Channel()
	for f := range ch2 {
		*flowjson = f
		break
	}
	got = flowjson.Users[0]
	want = client3.UserID
	if got != want {
		t.Fatalf("assertion failed")
	}
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 2 tries to create duo room with user 3
	flowjson = &db.FlowJSON{
		Mode:  "CreateDuoRoom",
		Users: []uint32{client2.UserID, client3.UserID},
	}
	client2.ClearChannel()
	client2.TxManage(flowjson)
	if flowjson.Err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// user 1 tries to create group with user 3
	flowjson = &db.FlowJSON{
		Mode:  "CreateGroupRoom",
		Users: []uint32{client1.UserID, client3.UserID},
		Name:  "Test Group",
	}
	client1.ClearChannel()
	client1.TxManage(flowjson)
	if flowjson.Err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// user 3 privacy settings: group = false , direct = true
	flowjson = &db.FlowJSON{
		Mode: "ChangePrivacyDirect",
		Bool: true,
	}
	client3.ClearChannel()
	client3.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}
	// user 2 add to duoroom user 3
	flowjson = &db.FlowJSON{
		Mode:  "CreateDuoRoom",
		Users: []uint32{client2.UserID, client3.UserID},
	}
	client2.ClearChannel()
	client2.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 3 privacy settings: all true
	flowjson = &db.FlowJSON{
		Mode: "ChangePrivacyGroup",
		Bool: true,
	}
	client3.ClearChannel()
	client3.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 1 add to group user 3
	flowjson = &db.FlowJSON{
		Mode:  "CreateGroupRoom",
		Users: []uint32{client1.UserID, client3.UserID},
		Name:  "Test Group",
	}
	client1.ClearChannel()
	client1.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}
	groupRoomID := flowjson.Room

	// user 3 finds user 2
	flowjson = &db.FlowJSON{
		Mode:  "FindUsers",
		Name:  "user2",
		Users: make([]uint32, 10),
	}
	client3.ClearChannel()
	go client3.TxManage(flowjson)
	ch3 := client3.Channel()
	for f := range ch3 {
		*flowjson = f
		break
	}
	got = flowjson.Users[0]
	want = client2.UserID
	if got != want {
		t.Fatalf("assertion failed")
	}
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 3 tries to add user 2 to user1 group
	flowjson = &db.FlowJSON{
		Mode:  "AddUsersToRoom",
		Users: []uint32{client2.UserID},
		Room:  groupRoomID,
	}
	client3.ClearChannel()
	client3.TxManage(flowjson)
	if flowjson.Err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// user 1 add to group user 2
	flowjson = &db.FlowJSON{
		Mode:  "AddUsersToRoom",
		Users: []uint32{client2.UserID},
		Room:  groupRoomID,
	}
	client1.ClearChannel()
	client1.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 2 tries to delete user 3 from user1 group
	flowjson = &db.FlowJSON{
		Mode:  "DeleteUsersFromRoom",
		Users: []uint32{client3.UserID},
		Room:  groupRoomID,
	}
	client2.ClearChannel()
	client2.TxManage(flowjson)
	if flowjson.Err == nil {
		t.Fatalf("%v %v Expected error, got nil", flowjson.Users, flowjson.Room)
	}

	// user 2 leaves group
	flowjson = &db.FlowJSON{
		Mode:  "DeleteUsersFromRoom",
		Users: []uint32{client2.UserID},
		Room:  groupRoomID,
	}
	client2.ClearChannel()
	client2.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}

	// user 3 leaves group
	flowjson = &db.FlowJSON{
		Mode:  "DeleteUsersFromRoom",
		Users: []uint32{client3.UserID},
		Room:  groupRoomID,
	}
	client3.ClearChannel()
	client3.TxManage(flowjson)
	if flowjson.Err != nil {
		t.Fatalf("%v", flowjson.Err)
	}
}
