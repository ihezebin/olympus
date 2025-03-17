package pulsar

import (
	"context"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/pkg/errors"

	"github.com/ihezebin/olympus/logger"
	"github.com/ihezebin/olympus/pubsub"
)

type SubOptions struct {
	pulsar.ClientOptions
	pulsar.ConsumerOptions
	NotAutoAck   bool
	ErrSleepTime time.Duration
}

type subscriber struct {
	Client     pulsar.Client
	Consumer   pulsar.Consumer
	Options    SubOptions
	Topic      string
	once       sync.Once
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

var _ pubsub.Subscriber = (*subscriber)(nil)

func NewSubscriber(opts SubOptions) (pubsub.Subscriber, error) {
	if opts.ClientOptions.Logger == nil {
		opts.ClientOptions.Logger = log.DefaultNopLogger()
	}

	client, err := pulsar.NewClient(opts.ClientOptions)
	if err != nil {
		return nil, err
	}

	if opts.ErrSleepTime == 0 {
		opts.ErrSleepTime = 1 * time.Second
	}

	consumer, err := client.Subscribe(opts.ConsumerOptions)
	if err != nil {
		return nil, err
	}

	return &subscriber{
		Client:   client,
		Consumer: consumer,
		Options:  opts,
		Topic:    opts.ConsumerOptions.Topic,
	}, nil
}

// Close 关闭前等待所有消息处理完成
func (s *subscriber) Close() error {
	var err error

	s.once.Do(func() {
		if s.cancelFunc != nil {
			s.cancelFunc()
		}
		s.wg.Wait()

		s.Consumer.Close()
		s.Client.Close()
	})

	return err
}

func (s *subscriber) Receive(ctx context.Context, handler pubsub.MessageHandler) error {
	s.wg.Add(1)

	defer func() {
		s.wg.Done()
		s.Close()
	}()

	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	receiveFunc := func() error {
		msg, err := s.Consumer.Receive(ctx)
		if err != nil {
			return errors.Wrap(err, "receive err")
		}

		// 自动 ack
		if !s.Options.NotAutoAck {
			ackErr := s.Consumer.Ack(msg)
			if ackErr != nil {
				logger.Warnf(ctx, "pulsar ack message err: %v", ackErr)
			}
		}

		cmsg := &pulsarConsumerMessage{
			Message: msg,
			ack:     s.Consumer.Ack,
		}
		err = handler(ctx, cmsg)
		if err != nil {
			return errors.Wrap(err, "handle err")
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		default:
			err := receiveFunc()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				logger.Debugf(ctx, "pulsar receive message err: %v", err)
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					continue
				}
				logger.Warningf(ctx, "failed to receive pulsar message")
				time.Sleep(s.Options.ErrSleepTime)
			}
		}

	}
}
