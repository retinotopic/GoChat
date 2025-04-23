package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/server/logger"
)

type Redis struct {
	Client  *redis.Client
	Log     logger.Logger
	Limiter *redis_rate.Limiter
}

func (r *Redis) Allow(ctx context.Context, key string, rate int, burst int, period time.Duration) (err error) {
	res, err := r.Limiter.Allow(context.Background(), key, redis_rate.Limit{
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
func (r *Redis) PublishWithSubscriptions(ctx context.Context, PubForSub []string, SubForPub []string, kind string) (err error) {

	for i := range PubForSub {
		for j := range SubForPub {
			msg := redis.Subscription{Kind: kind, Channel: SubForPub[j]}
			err := r.Client.Publish(ctx, PubForSub[i], msg).Err()
			if err != nil {
				return err
			}
		}
	}
	return
}
func (r *Redis) PublishWithMessage(ctx context.Context, SubForPub []string, message string) (err error) {
	msg := redis.Message{Payload: message}

	for i := range SubForPub {
		msg.Channel = SubForPub[i]
		err := r.Client.Publish(ctx, msg.Channel, msg).Err()
		if err != nil {
			return err
		}
	}
	return
}
func (r *Redis) Channel(ctx context.Context, closech <-chan bool, user string) <-chan []byte {
	PubSub := r.Client.Subscribe(ctx, user)
	ch := PubSub.ChannelWithSubscriptions()
	resultCh := make(chan []byte, 500)
	go func() {
		defer func() {
			close(resultCh)
			PubSub.Unsubscribe(context.TODO())
			PubSub.Close()
		}()
		select {
		case msg, ok := <-ch:
			if ok {
				switch v := msg.(type) {
				case *redis.Message:
					resultCh <- []byte(v.Payload)
				case *redis.Subscription:
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
