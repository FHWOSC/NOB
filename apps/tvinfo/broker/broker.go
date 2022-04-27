package broker

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"os"
)

var Redis *redis.Client

func init() {
	addr := os.Getenv("MESSAGE_BROKER_ADDR")
	if addr == "" {
		log.Println("[WARNING] MESSAGE_BROKER_ADDR is empty")
	}

	pass := os.Getenv("MESSAGE_BROKER_PASS")
	if pass == "" {
		log.Println("[WARNING] MESSAGE_BROKER_PASS is empty")
	}

	Redis = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})
}

func Publish(channel string, message interface{}) {
	messageForLog := fmt.Sprintf("%v", message)
	cut := 32
	if len(messageForLog) < cut {
		cut = len(messageForLog) - 1
	}
	log.Printf("\"%s\" trying to publish: %s...\n", channel, messageForLog[:cut])
	if Redis != nil {
		Redis.Publish(channel, message)
	}
}

func Subscribe(channels ...string) (*redis.PubSub, error) {
	if Redis != nil {
		pubsub := Redis.Subscribe(channels...)
		if pubsub == nil {
			return nil, errors.New("something happened")
		} else {
			return pubsub, nil
		}
	}

	return nil, errors.New("redis client doesn't exist")
}

func PSubscribe(channels ...string) (*redis.PubSub, error) {
	if Redis != nil {
		pubsub := Redis.PSubscribe(channels...)
		if pubsub == nil {
			return nil, errors.New("something happened")
		} else {
			return pubsub, nil
		}
	}

	return nil, errors.New("redis client doesn't exist")
}

func Close() {
	if Redis != nil {
		Redis.Close()
	}
}
