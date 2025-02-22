package redisbroadcast

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ihezebin/soup/pubsub"
)

/*
	SUBSCRIBE test

1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCAw\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:35.096919+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:35.09692+08:00\"}"
1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCAx\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:37.098804+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:37.098813+08:00\"}"
1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCAy\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:39.100307+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:39.100308+08:00\"}"
1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCAz\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:41.10173+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:41.101732+08:00\"}"
1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCA0\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:43.104307+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:43.104312+08:00\"}"
1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCA1\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:45.106882+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:45.106883+08:00\"}"
1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCA2\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:47.107786+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:47.107788+08:00\"}"
1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCA3\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:49.109418+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:49.109421+08:00\"}"
1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCA4\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:51.11089+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:51.110893+08:00\"}"
1) "message"
2) "test"
3) "{\"payload\":\"dGVzdCA5\",\"value\":null,\"key\":\"\",\"properties\":{},\"event_time\":\"2025-02-03T19:12:53.115079+08:00\",\"replication_clusters\":null,\"sequence_id\":0,\"deliver_after_time\":0,\"publish_time\":\"2025-02-03T19:12:53.115081+08:00\"}"
*/
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
