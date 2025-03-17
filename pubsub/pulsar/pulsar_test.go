package pulsar

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"

	"github.com/ihezebin/olympus/pubsub"
)

/*
pulsar-admin topics create-partitioned-topic persistent://public/default/test --partitions 1
*/
func TestPulsar(t *testing.T) {
	ctx := context.Background()

	sub, err := NewSubscriber(SubOptions{
		ClientOptions: pulsar.ClientOptions{
			URL: "pulsar://localhost:6650",
		},
		ConsumerOptions: pulsar.ConsumerOptions{
			Topic:            "persistent://public/default/test",
			SubscriptionName: "test",
			Type:             pulsar.Exclusive,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		pub, err := NewPublisher(PubOptions{
			ClientOptions: pulsar.ClientOptions{
				URL: "pulsar://localhost:6650",
			},
			ProducerOptions: pulsar.ProducerOptions{
				Topic: "persistent://public/default/test",
			},
		})
		if err != nil {
			sub.Close()
			t.Fatal(err)
		}

		for i := 0; i < 10; i++ {
			msg := pubsub.ProducerMessage{
				Payload: []byte(fmt.Sprintf("test %d", i)),
			}
			pub.Send(ctx, msg)

			t.Logf("unit test send msg: %d", i)
			time.Sleep(2 * time.Second)
		}

		pub.Close()
		time.Sleep(5 * time.Second)
		sub.Close()
	}()

	err = sub.Receive(ctx, func(ctx context.Context, message pubsub.ConsumerMessage) error {
		t.Logf("unit test receive msg: %s", string(message.Payload()))
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Log("unit test success")
}
