package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	PubSub *redis.PubSub
	Client *redis.Client
}

func (r *Redis) Subscribe(ctx context.Context, channels ...string) error {
	return r.PubSub.Subscribe(ctx, channels...)
}

func (r *Redis) Unsubscribe(ctx context.Context, channels ...string) error {
	return r.PubSub.Unsubscribe(ctx, channels...)
}

func (r *Redis) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.Client.Publish(ctx, channel, message).Err()
}