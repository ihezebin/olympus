package redisbroadcast

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"github.com/ihezebin/soup/logger"
	"github.com/ihezebin/soup/pubsub"
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
	sub        *redis.PubSub
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

	if opts.ErrSleepTime == 0 {
		opts.ErrSleepTime = 1 * time.Second
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
		if s.sub != nil {
			err = s.sub.Close()
			if err != nil {
				return
			}
		}
		err = s.UniversalClient.Close()
	})
	return err
}

func (s *subscriber) Receive(ctx context.Context, handler pubsub.MessageHandler) error {
	s.sub = s.UniversalClient.Subscribe(ctx, s.Options.Topics...)

	s.wg.Add(1)

	defer func() {
		s.wg.Done()
		s.sub.Close()
	}()

	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	ch := s.sub.Channel()
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		case msg := <-ch:
			smsg := &message{}
			err := json.Unmarshal([]byte(msg.Payload), smsg)
			if err != nil {
				return errors.Wrap(err, "unmarshal err")
			}
			smsg.topic = msg.Channel

			err = handler(ctx, smsg)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				logger.Warningf(ctx, "failed to receive redisbroadcast message")
				time.Sleep(s.Options.ErrSleepTime)
			}
		}
	}

}
