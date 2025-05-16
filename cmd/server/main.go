package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	db "github.com/retinotopic/GoChat/server/db/postgres"
	"github.com/retinotopic/GoChat/server/logger/loggers/zerolog"
	rd "github.com/retinotopic/GoChat/server/pubsub/impls/redis"
	"github.com/retinotopic/GoChat/server/router"
)

func main() {
	log := zerolog.NewZerologLogger(os.Stdout)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	pgHost := os.Getenv("PG_HOST")
	pgPort := os.Getenv("PG_PORT")
	pgUser := os.Getenv("PG_USER")
	pgPassword := os.Getenv("PG_PASSWORD")
	pgDB := os.Getenv("PG_DATABASE")
	pgSSL := os.Getenv("PG_SSLMODE")

	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s ",
		pgUser, pgPassword, pgHost, pgPort, pgDB, pgSSL,
	)
	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})
	dbgrd := false
	if os.Getenv("REDIS_DEBUG") == "true" {
		dbgrd = true
	}
	rds := &rd.Redis{
		Debug:   dbgrd,
		Client:  client,
		Log:     log,
		Limiter: redis_rate.NewLimiter(client),
	}
	pgclient, err := db.NewPgClient(ctx, dsn, rds)
	if err != nil {
		log.Fatal("db new pool:", err)
	}

	FetchUser := func(w http.ResponseWriter, r *http.Request) (string, error) {
		c, err := r.Cookie("username")
		if err != nil {
			return "", err
		}
		return c.Value, nil
	}
	dbs := stdlib.OpenDBFromPool(pgclient.Pool)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("goose set dialect:", err)
	}
	if err := goose.Up(dbs, os.Getenv("MIGRATIONS_DIR")); err != nil {
		log.Fatal("goose up:", err)
	}
	if err := dbs.Close(); err != nil {
		log.Fatal("close db conn for migrations:", err)
	}
	srv := router.NewRouter("0.0.0.0:"+os.Getenv("APP_PORT"), FetchUser, rds, pgclient, log)
	err = srv.Run(ctx)
	if err != nil {
		log.Fatal("server run:", err)
	}
}
