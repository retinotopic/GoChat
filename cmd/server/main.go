package main

import (
	"context"
	"flag"
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
	pgHost := os.Getenv("POSTGRES_HOST")
	pgPort := os.Getenv("POSTGRES_PORT")
	pgUser := os.Getenv("POSTGRES_USER")
	pgPassword := os.Getenv("POSTGRES_PASSWORD")
	pgDB := os.Getenv("POSTGRES_DB")
	pgSSLMode := os.Getenv("POSTGRES_SSLMODE")

	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		pgUser, pgPassword, pgHost, pgPort, pgDB, pgSSLMode,
	)
	flag.Parse()
	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
	})
	rds := &rd.Redis{
		Client:  client,
		Log:     log,
		Limiter: redis_rate.NewLimiter(client),
	}
	pgclient, err := db.NewPgClient(ctx, dsn, rds)
	if err != nil {
		log.Fatal("db new pool:", err)
	}

	//FetchUser(http.ResponseWriter, *http.Request) (string, error)
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
	if err := goose.Up(dbs, "/bin/maindir/migrations"); err != nil {
		log.Fatal("goose up:", err)
	}
	if err := dbs.Close(); err != nil {
		log.Fatal("close db conn for migrations:", err)
	}
	srv := router.NewRouter(os.Getenv("SERVER_CONNECT_ADDR"), FetchUser, rds, pgclient, log)

	err = srv.Run(ctx)
	if err != nil {
		log.Fatal("server run:", err)
	}
}
