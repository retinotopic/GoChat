package main

import (
	"context"
	"flag"
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
	addr := flag.String("addr", "0.0.0.0:8080", "app address")
	rdaddr := flag.String("rdaddr", "redis:6379", "redis address")
	pgaddr := flag.String("pgaddr", "postgres:5432", "postgres address")
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	flag.Parse()
	client := redis.NewClient(&redis.Options{
		Addr: *rdaddr,
	})
	rds := &rd.Redis{
		Client:  client,
		Log:     log,
		Limiter: redis_rate.NewLimiter(client),
	}
	pgclient, err := db.NewPgClient(ctx, *pgaddr, rds)
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
	if err := goose.Up(dbs, "migrations"); err != nil {
		log.Fatal("goose up:", err)
	}
	if err := dbs.Close(); err != nil {
		log.Fatal("close db conn for migrations:", err)
	}
	srv := router.NewRouter(*addr, FetchUser, rds, pgclient, log)

	err = srv.Run(ctx)
	if err != nil {
		log.Fatal("server run:", err)
	}
}
