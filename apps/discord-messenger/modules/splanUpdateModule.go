package modules

import (
	"discord-messenger/discord"
	"discord-messenger/handler"
	"fmt"
	"log"
	"time"
)

const (
	splanUrl = "https://intern.fh-wedel.de/~splan/"
)

var SplanUpdateModule = handler.Module{
	Pattern: "splan.timestamp.changed",
	Handler: splanTimestampHandler,
}

func splanTimestampHandler(bot discord.Bot, channel, pattern, payload string) {
	ts, err := time.Parse(time.RFC3339, payload)
	if err != nil {
		log.Println("ERR - splanTimestamp:", err)
	}

	err = bot.SendMessage(discord.Announcement, fmt.Sprintf("\n Neuer Vorlesungsplan: %s\n(%s)",
		splanUrl,
		ts.Format("02 Jan 2006 15:04"),
	))
	if err != nil {
		log.Println("ERR - splanTimestamp:", err)
	}
}
