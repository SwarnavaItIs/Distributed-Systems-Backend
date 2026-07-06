package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	notificationws "github.com/swarnava/dmb/services/notification/internal/ws"
)

type Subscriber struct {
	client  *redis.Client
	channel string
	manager *notificationws.Manager
}

func NewSubscriber(
	redisAddr string,
	channel string,
	manager *notificationws.Manager,
) *Subscriber {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &Subscriber{
		client:  client,
		channel: channel,
		manager: manager,
	}
}

func (s *Subscriber) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *Subscriber) Close() error {
	return s.client.Close()
}

func (s *Subscriber) Start(ctx context.Context) error {
	pubsub := s.client.Subscribe(ctx, s.channel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("failed to subscribe to redis channel %s: %w", s.channel, err)
	}

	fmt.Println("Subscribed to Redis channel:", s.channel)

	messageChannel := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case message, ok := <-messageChannel:
			if !ok {
				return nil
			}

			s.handleMessage(message.Payload)
		}
	}
}

func (s *Subscriber) handleMessage(payload string) {
	var event map[string]any

	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		s.manager.BroadcastJSON(map[string]any{
			"type":      "redis.message",
			"channel":   s.channel,
			"message":   payload,
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	event["source"] = "redis_pubsub"
	event["channel"] = s.channel

	s.manager.BroadcastJSON(event)
}