package main

import (
	"discord-messenger/discord"
	"discord-messenger/event"
	"discord-messenger/modules"
	routerpkg "discord-messenger/router"
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

	discordBot, err := discord.New(
		os.Getenv("DISCORD_TOKEN"),
	)
	if err != nil {
		panic(err)
	}
	defer discordBot.Close()

	router := routerpkg.New(broker, discordBot)
	go router.RegisterModule(modules.SplanUpdateModule)
	//go router.RegisterModule(modules.LoggerModule)
	go router.RegisterModule(modules.CatchAllModule)
	go router.RegisterModule(modules.TvInfoUpdateModule)
	go router.RegisterModule(modules.TvInfoImageUpdateModule)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
