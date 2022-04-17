package modules

import (
	"discord-messenger/discord"
	"discord-messenger/handler"
	"fmt"
	"log"
)

var LoggerModule = handler.Module{
	Pattern: "*logged*",
	Handler: loggerModule,
}

func loggerModule(bot discord.Bot, channel, pattern, payload string) {
	err := bot.SendMessage(discord.Log, fmt.Sprintf("LOG > [%s|%s] %s\n", channel, pattern, payload))
	if err != nil {
		log.Println("ERR - loggerModule:", err)
	}
}
