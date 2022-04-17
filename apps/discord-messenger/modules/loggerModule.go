package modules

import (
	"discord-messenger/discord"
	"discord-messenger/handler"
	"fmt"
)

var LoggerModule = handler.Module{
	Pattern: "*logged*",
	Handler: loggerModule,
}

func loggerModule(bot discord.Bot, channel, pattern, payload string) {
	bot.SendMessage(discord.Announcement, fmt.Sprintf("LOG > [%s|%s] %s\n", channel, pattern, payload))
}
