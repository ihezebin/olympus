package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ihezebin/olympus/pubsub"
)

func TestRedisBroadcast(t *testing.T) {
	ctx := context.Background()

	sub, err := NewSubscriber(SubOptions{
		UniversalOptions: redis.UniversalOptions{
			Addrs:    []string{"localhost:6379"},
			Password: "root",
		},
		Topics: []string{"test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	go func() {
		pub, err := NewPublisher(PubOptions{
			UniversalOptions: redis.UniversalOptions{
				Addrs:    []string{"localhost:6379"},
				Password: "root",
			},
			Topic: "test",
		})
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 10; i++ {
			pub.Send(ctx, pubsub.ProducerMessage{
				Payload: []byte(fmt.Sprintf("test %d", i)),
			})

			t.Logf("unit test send msg: %d", i)
			time.Sleep(1 * time.Second)
		}

		pub.Close()
		time.Sleep(3 * time.Second)
		sub.Close()
	}()

	sub.Receive(ctx, func(ctx context.Context, message pubsub.ConsumerMessage) error {
		t.Logf("unit test receive msg: %s", string(message.Payload()))
		return nil
	})

	t.Log("unit test success")
}
