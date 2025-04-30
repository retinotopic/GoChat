package redis

import (
	"context"
	"errors"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/server/logger"
)

type Redis struct {
	Client  *redis.Client
	Log     logger.Logger
	Limiter *redis_rate.Limiter
}
type Action struct {
	SubKind string `json:"SubKind"`
	Ch      string `json:"Ch"`
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
			act := Action{SubKind: kind, Ch: SubForPub[j]}
			b, err := json.Marshal(act)
			if err != nil {
				return err
			}
			err = r.Client.Publish(ctx, PubForSub[i], b).Err()
			if err != nil {
				return err
			}
		}
	}
	return
}

func (r *Redis) PublishWithMessage(ctx context.Context, SubForPub []string, message string) (err error) {
	for i := range SubForPub {
		err := r.Client.Publish(ctx, SubForPub[i], message).Err()
		if err != nil {
			return err
		}
	}
	return
}

func (r *Redis) Channel(ctx context.Context, closech <-chan bool, user string) <-chan []byte {
	PubSub := r.Client.Subscribe(ctx, user)
	ch := PubSub.Channel()
	resultCh := make(chan []byte, 50)
	go func() {
		defer func() {
			close(resultCh)
			PubSub.Unsubscribe(context.TODO())
			PubSub.Close()
		}()
		for {
			select {
			case msg, ok := <-ch:
				if ok {
					a := Action{}
					err := json.Unmarshal([]byte(msg.Payload), &a)
					if err != nil {
						r.Log.Error("redis unmarshal", err)
						return
					}
					r.Log.Error(msg.Payload+" user: "+user, errors.New("sdsosijdfgpoisfjgoidfo AAA"))
					if len(a.SubKind) != 0 {
						if a.SubKind == "0" {
							err := PubSub.Unsubscribe(context.TODO(), a.Ch)
							if err != nil {
								r.Log.Error("redis unmarshal", err)
								return
							}
						} else {
							err := PubSub.Subscribe(context.TODO(), a.Ch)
							if err != nil {
								r.Log.Error("redis unmarshal", err)
								return
							}
						}
					} else {
						resultCh <- []byte(msg.Payload)
					}
				} else {
					r.Log.Error("wtf ?????", errors.New(">????????????????????????????????<"))
				}
			case <-closech:
				return
			}
		}

	}()

	return resultCh
}
