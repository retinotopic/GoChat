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
	Id      string `json:"Id"`
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
func (r *Redis) PublishWithSubscriptions(ctx context.Context, UserChs []string, PublishChs []string, kind string) (err error) {
	for i := range UserChs {
		for j := range PublishChs {
			id := "$"
			info, err := r.Client.XInfoStream(ctx, PublishChs[j]).Result()
			if err != nil {
				if err.Error() != "ERR no such key" {
					return err
				}
			} else {
				id = info.LastEntry.ID
			}
			act := Action{SubKind: kind, Ch: PublishChs[j], Id: id}
			b, err := json.MarshalString(act)
			if err != nil {
				return err
			}
			err = r.Client.XAdd(ctx, &redis.XAddArgs{
				Stream: UserChs[i],
				Values: []string{"Action", b},
				ID:     "*",
			}).Err()
			if err != nil {
				return err
			}
		}
	}
	return
}

func (r *Redis) PublishWithMessage(ctx context.Context, PublishChs []string, message string) (err error) {
	for i := range PublishChs {
		err := r.Client.XAdd(ctx, &redis.XAddArgs{
			Stream: PublishChs[i],
			Values: []string{"Data", message},
			ID:     "*",
		}).Err()
		if err != nil {
			return err
		}
	}
	return
}

func (r *Redis) Channel(user string) <-chan []byte {
	m := map[string]string{user: "$"}
	resultCh := make(chan []byte, 10)
	strmnids := make([]string, 0, 100)
	r.Log.Error("here we are chan", errors.New(user))
	go func() (err error) {
		defer func() {
			close(resultCh)
			r.Log.Error("here we are::::::", err)
		}()
		for {
			strmnids = strmnids[:0]
			for ch := range m {
				strmnids = append(strmnids, ch)
			}
			slen := len(strmnids)
			for i := range slen {
				strmnids = append(strmnids, m[strmnids[i]])
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*25)
			res, err := r.Client.XRead(ctx, &redis.XReadArgs{
				Streams: strmnids,
				Block:   time.Millisecond * 100,
				Count:   10,
			}).Result()
			cancel()
			if err != nil && err != redis.Nil {
				return err
			}
			if len(res) != 0 {
				r.Log.Error(user, errors.New("WAT"))
				for i := range res {
					for _, v := range res[i].Messages {
						switch res[i].Stream {
						case user:
							i, ok := v.Values["Action"]
							r.Log.Error("val"+user, errors.New("333"))
							if ok {
								b, ok := i.(string)
								if ok {
									r.Log.Error("val indeed"+user, errors.New("333"))
									a := Action{}
									err = json.UnmarshalString(b, &a)
									if err != nil {
										r.Log.Error("redis unmarshal", err)
										return err
									}
									switch a.SubKind {
									case "0":
										delete(m, a.Ch)
									case "1":
										m[a.Ch] = a.Id
									}
								}
							}
						default:
							i, ok := v.Values["Data"]
							r.Log.Error("val"+user, errors.New("123"))
							if ok {
								r.Log.Error("val indeed"+user, errors.New("123"))
								b, ok := i.(string)
								if ok {
									resultCh <- []byte(b)
								}
							}
						}
						m[res[i].Stream] = v.ID
					}
				}
			}
		}
	}()
	return resultCh
}
