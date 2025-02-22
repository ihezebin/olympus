package pulsar

import (
	"context"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/log"

	"github.com/ihezebin/soup/pubsub"
)

type PubOptions struct {
	pulsar.ClientOptions
	pulsar.ProducerOptions
}

type publisher struct {
	Client   pulsar.Client
	Producer pulsar.Producer
	Options  PubOptions
	Topic    string
}

var _ pubsub.Publisher = (*publisher)(nil)

func NewPublisher(opts PubOptions) (pubsub.Publisher, error) {
	clientOpts := opts.ClientOptions
	if clientOpts.Logger == nil {
		clientOpts.Logger = log.DefaultNopLogger()
	}

	client, err := pulsar.NewClient(clientOpts)
	if err != nil {
		return nil, err
	}

	producerOpts := opts.ProducerOptions
	if producerOpts.BatchingMaxPublishDelay == 0 {
		producerOpts.BatchingMaxPublishDelay = 10 * time.Millisecond
	}

	if producerOpts.BatchingMaxMessages == 0 {
		producerOpts.BatchingMaxMessages = 1000
	}

	if producerOpts.BatchingMaxSize == 0 {
		producerOpts.BatchingMaxSize = 1024 * 1024 // default 1MB
	}

	producer, err := client.CreateProducer(opts.ProducerOptions)
	if err != nil {
		return nil, err
	}

	return &publisher{
		Client:   client,
		Producer: producer,
		Options:  opts,
		Topic:    opts.ProducerOptions.Topic,
	}, nil
}

func (p *publisher) Close() error {
	p.Producer.Close()
	p.Client.Close()
	return nil
}

func (p *publisher) Send(ctx context.Context, msg pubsub.ProducerMessage) error {
	pmsg := &pulsar.ProducerMessage{
		Payload:             msg.Payload,
		Value:               msg.Value,
		Key:                 msg.Key,
		Properties:          msg.Properties,
		EventTime:           msg.EventTime,
		ReplicationClusters: msg.ReplicationClusters,
		SequenceID:          &msg.SequenceID,
		DeliverAfter:        msg.DeliverAfterTime,
	}

	_, err := p.Producer.Send(ctx, pmsg)

	return err
}

func (p *publisher) SendAsync(ctx context.Context, msg pubsub.ProducerMessage, callback func(pubsub.ProducerMessage, error)) {
	pmsg := &pulsar.ProducerMessage{
		Payload:             msg.Payload,
		Value:               msg.Value,
		Key:                 msg.Key,
		Properties:          msg.Properties,
		EventTime:           msg.EventTime,
		ReplicationClusters: msg.ReplicationClusters,
		SequenceID:          &msg.SequenceID,
		DeliverAfter:        msg.DeliverAfterTime,
	}

	p.Producer.SendAsync(ctx, pmsg, func(_ pulsar.MessageID, _ *pulsar.ProducerMessage, err error) {
		callback(msg, err)
	})
}
