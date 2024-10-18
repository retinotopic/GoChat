package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/internal/logger"
)

type Redis struct {
	Client *redis.Client
	Log    logger.Logger
	rl     *redis_rate.Limiter
}
type PubSuber interface {
	Unsubscribe(ctx context.Context, channels ...string) error
	Subscribe(ctx context.Context, channels ...string) error
}

func (r *Redis) Allow(ctx context.Context, key string, rate int, burst int, period time.Duration) (err error) {
	res, err := r.rl.Allow(context.Background(), key, redis_rate.Limit{
		Rate:   rate,
		Burst:  burst,
		Period: period,
	})
	if err != nil {
		return err
	}
	if res.Allowed > 0 {
		return nil
	}
	return errors.New("limit exceeded, retry after " + res.RetryAfter.String())
}
func (r *Redis) PublishWithSubscriptions(ctx context.Context, pubChannels []string, subChannel string, kind string) (err error) {
	msg := redis.Subscription{Kind: kind, Channel: subChannel}
	for i := range pubChannels {
		err := r.Client.Publish(ctx, pubChannels[i], msg).Err()
		if err != nil {
			return err
		}
	}
	return
}
func (r *Redis) Publish(ctx context.Context, channel string, message string) error {
	msg := redis.Message{Channel: channel, Payload: message}
	return r.Client.Publish(ctx, channel, msg).Err()
}
func (r *Redis) Channel(ctx context.Context, closech <-chan bool, user string) <-chan []byte {
	PubSub := r.Client.Subscribe(ctx, user)
	ch := PubSub.ChannelWithSubscriptions()
	resultCh := make(chan []byte, 500)
	go func() {
		defer func() {
			PubSub.Unsubscribe(context.TODO(), user)
			PubSub.Close()
			close(resultCh)
		}()
		select {
		case msg, ok := <-ch:
			if ok {
				switch v := msg.(type) {
				case redis.Message:
					resultCh <- []byte(v.Payload)
				case redis.Subscription:
					if len(v.Kind) == 0 {
						err := PubSub.Unsubscribe(context.TODO(), v.Channel)
						if err != nil {
							return
						}
					} else {
						err := PubSub.Subscribe(context.TODO(), v.Channel)
						if err != nil {
							return
						}
					}
				default:
					r.Log.Error("Message from redis msg channel", errors.New("undefined message"))
				}
			}
		case <-closech:
			return
		}

	}()

	return resultCh
}
