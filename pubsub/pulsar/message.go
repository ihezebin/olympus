package pulsar

import (
	"github.com/apache/pulsar-client-go/pulsar"

	"github.com/ihezebin/olympus/pubsub"
)

type pulsarConsumerMessage struct {
	pulsar.Message
	ack func(msg pulsar.Message) error
}

var _ pubsub.ConsumerMessage = (*pulsarConsumerMessage)(nil)

func (m *pulsarConsumerMessage) ID() string {
	return m.Message.ID().String()
}

func (m *pulsarConsumerMessage) Value(v interface{}) error {
	return m.Message.GetSchemaValue(v)
}

func (m *pulsarConsumerMessage) Ack() error {
	return m.ack(m.Message)
}
