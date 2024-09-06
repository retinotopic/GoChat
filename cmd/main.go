package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	db "github.com/retinotopic/GoChat/internal/dbrepo/postgres"
	"github.com/retinotopic/GoChat/internal/logger/loggers/zerolog"
	"github.com/retinotopic/GoChat/internal/models"
	"github.com/retinotopic/GoChat/internal/pubsub"
	rd "github.com/retinotopic/GoChat/internal/pubsub/impls/redis"
	"github.com/retinotopic/GoChat/internal/router"
	"github.com/retinotopic/tinyauth/provider"
	"github.com/retinotopic/tinyauth/providers/firebase"
	"github.com/retinotopic/tinyauth/providers/google"
)

func main() {
	log := zerolog.NewZerologLogger(os.Stdout)
	addr := flag.String("addr", "localhost:8080", "address to listen on")
	addrpb := flag.String("addrpb", "localhost:6379", "redis pubsub address")
	addrdb := flag.String("addrdb", "localhost:5432", "postgres address")
	mp := make(map[string]provider.Provider)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	google, err := google.New(ctx, os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("REDIRECT"), "/refresh")
	if err != nil {
		log.Fatal("google provider:", err)
	}
	firebase, err := firebase.New(ctx, os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("REDIRECT"), "refresh")
	if err != nil {
		log.Fatal("firebase provider:", err)
	}

	// with the issuer claim, we know which provider issued token.
	mp[os.Getenv("ISSUER_FIREBASE")] = firebase
	mp[os.Getenv("ISSUER_GOOGLE")] = google
	mp[os.Getenv("ISSUER2_GOOGLE")] = google

	flag.Parse()
	client := redis.NewClient(&redis.Options{
		Addr: *addrpb,
	})

	pool, err := db.NewPool(ctx, *addrdb)
	if err != nil {
		log.Fatal("db new pool:", err)
	}
	fndb := func(ctx context.Context, sub string) (pubsub.Databaser, error) {
		pg, err := pool.GetClient(ctx, sub)
		return pg, err
	}
	fnps := func(ctx context.Context) (pubsub.PubSuber, error) {
		ps := client.Subscribe(ctx)
		return &rd.Redis{PubSub: ps, Client: client, Chan: make(chan models.Flowjson, 500)}, nil
	}
	dbs := stdlib.OpenDBFromPool(pool.Pool)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("goose set dialect:", err)
	}
	if err := goose.Up(dbs, "migrations"); err != nil {
		log.Fatal("goose up:", err)
	}
	if err := dbs.Close(); err != nil {
		log.Fatal("close db conn for migrations:", err)
	}
	srv := router.NewRouter(*addr, mp, pubsub.Connector{GetDB: fndb, GetPS: fnps, Log: log}, log)

	err = srv.Run(ctx)
	if err != nil {
		log.Fatal("server run:", err)
	}
}
