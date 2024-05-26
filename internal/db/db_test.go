package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/retinotopic/GoChat/internal/db"
)

func TestDb(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	time.Sleep(time.Second * 5)
	pool, err := db.ConnectToDB(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = ctx.Err()
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = db.NewUser(ctx, "34hb534ihg5", "test1", pool)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
