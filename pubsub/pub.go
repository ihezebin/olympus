package pubsub

import (
	"context"
	"time"
)

type ProducerMessage struct {
	// Payload for the message
	Payload []byte `json:"payload"`
	//Value and payload is mutually exclusive, `Value interface{}` for schema message.
	Value interface{} `json:"value"`
	// Sets the key of the message for routing policy
	Key string `json:"key"`
	// Attach application defined properties on the message
	Properties map[string]string `json:"properties"`
	// Set the event time for a given message
	EventTime time.Time `json:"event_time"`
	// Override the replication clusters for this message.
	ReplicationClusters []string `json:"replication_clusters"`
	// Set the sequence id to assign to the current message
	SequenceID int64 `json:"sequence_id"`
	// diliver delay, not all implimentaion support this property
	DeliverAfterTime time.Duration `json:"deliver_after_time"`
}

type Publisher interface {
	Close() error
	Send(ctx context.Context, msg ProducerMessage) (err error)
	SendAsync(ctx context.Context, msg ProducerMessage, callback func(ProducerMessage, error))
}
