package modules

import (
	"bytes"
	"discord-messenger/discord"
	"discord-messenger/handler"
	"encoding/json"
	"log"
)

var CatchAllModule = handler.Module{
	Pattern: "*",
	Handler: catchAllModule,
}

type message struct {
	Channel string `json:"channel"`
	Pattern string `json:"pattern"`
	Payload string `json:"payload"`
}

func catchAllModule(bot discord.Bot, channel, pattern, payload string) {
	m := message{
		Channel: channel,
		Pattern: pattern,
		Payload: payload,
	}
	err := bot.SendMessage(discord.Log, m.String())
	if err != nil {
		log.Println("ERR - catchAllModule:", err)
	}
}

func (m *message) String() string {
	byts, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}

	var buf bytes.Buffer
	err = json.Compact(&buf, byts)
	if err != nil {
		return "{}"
	}

	return string(buf.Bytes())
}
