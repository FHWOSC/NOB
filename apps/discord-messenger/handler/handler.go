package handler

import "discord-messenger/discord"

type HandlerFunc func(bot discord.Bot, channel, pattern, payload string)

type Module struct {
	Pattern string
	Handler HandlerFunc
}
