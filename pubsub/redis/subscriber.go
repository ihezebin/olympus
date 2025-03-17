package redis

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"github.com/ihezebin/olympus/logger"
	"github.com/ihezebin/olympus/pubsub"
)

type SubOptions struct {
	redis.UniversalOptions
	Topics       []string
	ErrSleepTime time.Duration
}

type subscriber struct {
	redis.UniversalClient
	Options    SubOptions
	once       sync.Once
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

var _ pubsub.Publisher = (*publisher)(nil)

func NewSubscriber(opts SubOptions) (pubsub.Subscriber, error) {
	client := redis.NewUniversalClient(&opts.UniversalOptions)
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &subscriber{
		UniversalClient: client,
		Options:         opts,
	}, nil
}

func (s *subscriber) Close() error {
	var err error
	s.once.Do(func() {
		if s.cancelFunc != nil {
			s.cancelFunc()
		}
		s.wg.Wait()
		err = s.UniversalClient.Close()
	})
	return err
}

func (s *subscriber) Receive(ctx context.Context, handler pubsub.MessageHandler) error {
	s.wg.Add(1)

	defer func() {
		s.wg.Done()
		_ = s.UniversalClient.Close()
	}()

	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		default:
			val, err := s.UniversalClient.BRPop(ctx, time.Second*3, s.Options.Topics...).Result()
			if err != nil && !errors.Is(err, redis.Nil) {
				return errors.Wrap(err, "brpop err")
			}

			if len(val) != 2 {
				return errors.New("bad values len")
			}

			smsg := &message{}
			err = json.Unmarshal([]byte(val[1]), smsg)
			if err != nil {
				return errors.Wrap(err, "unmarshal err")
			}
			smsg.topic = val[0]

			err = handler(ctx, smsg)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				logger.Warningf(ctx, "failed to receive redis message")
				time.Sleep(s.Options.ErrSleepTime)
			}
		}
	}

}
