package event

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
)

type Broker struct {
	client *redis.Client
}

func NewBroker(address, password string) *Broker {
	b := new(Broker)
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	b.client = r
	return b
}

func (b *Broker) Publish(channel, payload string) error {
	if b == nil {
		return fmt.Errorf("broker doesn't exist (nil)")
	}

	cmd := b.client.Publish(channel, payload)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	return nil
}

func (b *Broker) Subscribe(ctx context.Context, pattern string, f func(channel, pattern, payload string)) error {
	if b == nil {
		return fmt.Errorf("broker doesn't exist (nil)")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		sub := b.client.PSubscribe(pattern)
		defer sub.PUnsubscribe(pattern)

		for message := range sub.Channel() {
			f(message.Channel, message.Pattern, message.Payload)
		}

		return nil
	}
}

func (b *Broker) Close() {
	b.client.Close()
}
