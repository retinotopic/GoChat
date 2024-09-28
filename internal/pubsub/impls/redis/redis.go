package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/internal/logger"
)

type Redis struct {
	PubSub *redis.PubSub
	Client *redis.Client
	Log    logger.Logger
	Chan   <-chan interface{}
}

func (r *Redis) Publish(ctx context.Context, channel string, subscription bool, kind string, message string) error {
	var msg interface{}
	if subscription {
		msg = redis.Subscription{Kind: kind, Channel: channel}
	} else {
		msg = redis.Message{Channel: channel, Payload: message}
	}
	return r.Client.Publish(ctx, channel, msg).Err()
}
func (r *Redis) Channel() <-chan []byte {
	ch := r.PubSub.ChannelWithSubscriptions()
	resultCh := make(chan []byte, 500)
	go func() {
		defer close(resultCh)
		for msg := range ch {
			switch v := msg.(type) {
			case redis.Message:
				resultCh <- []byte(v.Payload)
			case redis.Subscription:
				if len(v.Kind) == 0 {
					err := r.PubSub.Unsubscribe(context.TODO(), v.Channel)
					if err != nil {
						return
					}
				} else {
					err := r.PubSub.Subscribe(context.TODO(), v.Channel)
					if err != nil {
						return
					}
				}
			default:
				r.Log.Error("Message from redis msg channel", errors.New("undefined message"))
			}
		}
	}()

	return resultCh
}
