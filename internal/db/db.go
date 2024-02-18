package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToDB(connString string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	return db, err
}
