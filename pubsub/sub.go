package pubsub

import (
	"context"
	"time"
)

// Identifier for a particular message
type MessageID interface {
	// Serialize the message id into a sequence of bytes that can be stored somewhere else
	Serialize() []byte
}

type ConsumerMessage interface {
	Topic() string
	Properties() map[string]string
	Payload() []byte
	ID() string
	PublishTime() time.Time
	EventTime() time.Time
	Key() string
	Value(v interface{}) error
	Ack() error
}

type MessageHandler func(ctx context.Context, message ConsumerMessage) error

type Subscriber interface {
	//Close stop receive message, and close any connection to the queue
	Close() error
	//Start start a loop to receive message from queue, the loop do not stop when handler return err
	Receive(ctx context.Context, handler MessageHandler) error
}
