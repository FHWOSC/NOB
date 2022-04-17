package main

import (
	"context"
	"discord-messenger/event"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("running...")

	broker := event.NewBroker(
		os.Getenv("MESSAGE_BROKER_ADDR"),
		os.Getenv("MESSAGE_BROKER_PASS"),
	)
	defer broker.Close()

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := broker.Subscribe(ctx, "splan.timestamp.changed", func(_, _, payload string) {
			fmt.Println("TIMESTAMP CHANGED =>", payload)
		})
		if err != nil {
			log.Fatalln(err)
		}
	}()

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := broker.Subscribe(ctx, "*logged*", func(channel, pattern, payload string) {
			fmt.Printf("LOG > [%s|%s] %s\n", channel, pattern, payload)
		})
		if err != nil {
			log.Fatalln(err)
		}
	}()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
