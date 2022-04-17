package router

import (
	"context"
	"discord-messenger/discord"
	"discord-messenger/event"
	"discord-messenger/handler"
	"log"
)

type Router struct {
	bot    discord.Bot
	broker *event.Broker
}

func New(broker *event.Broker, bot discord.Bot) *Router {
	r := new(Router)
	r.bot = bot
	r.broker = broker

	return r
}

func (r *Router) RegisterModule(module handler.Module) {
	log.Println("registering handler for", module.Pattern)
	r.Register(module.Pattern, module.Handler)
}

func (r *Router) Register(pattern string, handlerFunc handler.HandlerFunc) (error, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	err := r.broker.Subscribe(ctx, pattern, func(channel, pattern, payload string) {
		log.Println("handling message:", channel)
		handlerFunc(r.bot, channel, pattern, payload)
	})
	if err != nil {
		cancel()
		return err, nil
	}

	return nil, cancel
}
