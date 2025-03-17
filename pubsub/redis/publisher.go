package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"github.com/ihezebin/olympus/pubsub"
)

type PubOptions struct {
	redis.UniversalOptions
	Topic string
}

type publisher struct {
	redis.UniversalClient
	Options PubOptions
}

var _ pubsub.Publisher = (*publisher)(nil)

func NewPublisher(opts PubOptions) (pubsub.Publisher, error) {
	client := redis.NewUniversalClient(&opts.UniversalOptions)
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &publisher{
		UniversalClient: client,
		Options:         opts,
	}, nil
}

func (p *publisher) Close() error {
	return p.UniversalClient.Close()
}

func (p *publisher) Send(ctx context.Context, msg pubsub.ProducerMessage) error {
	if msg.EventTime.IsZero() {
		msg.EventTime = time.Now()
	}

	if msg.Properties == nil {
		msg.Properties = make(map[string]string)
	}

	pmsg := message{
		ProducerMessage: msg,
		P:               time.Now(),
		topic:           p.Options.Topic,
	}

	data, err := json.Marshal(pmsg)
	if err != nil {
		return errors.Wrap(err, "marshal message")
	}

	err = p.UniversalClient.LPush(ctx, p.Options.Topic, data).Err()
	if err != nil {
		return errors.Wrap(err, "redis publish message")
	}

	return nil
}

func (p *publisher) SendAsync(ctx context.Context, msg pubsub.ProducerMessage, callback func(pubsub.ProducerMessage, error)) {
	err := p.Send(ctx, msg)

	if callback != nil {
		callback(msg, err)
	}
}
