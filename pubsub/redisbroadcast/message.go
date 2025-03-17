package redisbroadcast

import (
	"errors"
	"time"

	"github.com/ihezebin/olympus/pubsub"
)

type message struct {
	pubsub.ProducerMessage
	P     time.Time `json:"publish_time"`
	topic string
}

var _ pubsub.ConsumerMessage = (*message)(nil)

func (m *message) Topic() string {
	return m.topic
}

func (m *message) Properties() map[string]string {
	return m.ProducerMessage.Properties
}

func (m *message) Payload() []byte {
	return m.ProducerMessage.Payload
}

func (m *message) ID() string {
	return ""
}

func (m *message) PublishTime() time.Time {
	return m.P
}

func (m *message) EventTime() time.Time {
	return m.ProducerMessage.EventTime
}

func (m *message) Key() string {
	return ""
}

func (m *message) Value(v interface{}) error {
	return errors.New("not implemented")
}

func (m *message) Ack() error {
	return nil
}
