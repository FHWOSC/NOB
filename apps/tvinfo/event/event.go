package event

import (
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

func (b *Broker) Close() {
	b.client.Close()
}
