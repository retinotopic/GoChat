package redis

import (
	"context"

	"github.com/goccy/go-json"

	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/internal/models"
)

type Redis struct {
	PubSub *redis.PubSub
	Client *redis.Client
	Chan   chan models.Flowjson
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
func (r *Redis) Channel() chan models.Flowjson {
	ch := r.PubSub.Channel()
	go func() {
		for m := range ch {
			fj := models.Flowjson{}
			if err := json.Unmarshal([]byte(m.Payload), &fj); err != nil {
				fj.ErrorMsg = "unmarshall error"
			}
			r.Chan <- fj
		}
		close(r.Chan)
	}()
	return r.Chan
}
