package event

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"log"
)

type Broker struct {
	client       *redis.Client
	unsubscriber []func()
}

func NewBroker(address, password string) *Broker {
	b := new(Broker)
	b.unsubscriber = make([]func(), 0)
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
		b.appendCloser(func() {
			log.Println("unsubscribing", pattern)
			sub.PUnsubscribe(pattern)
		})

		for message := range sub.Channel() {
			f(message.Channel, message.Pattern, message.Payload)
		}

		return nil
	}
}

func (b *Broker) appendCloser(f func()) {
	b.unsubscriber = append(b.unsubscriber, f)
}

func (b *Broker) Close() {
	for _, cancel := range b.unsubscriber {
		cancel()
	}

	b.client.Close()

}
