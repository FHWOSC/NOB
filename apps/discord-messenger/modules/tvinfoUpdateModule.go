package modules

import (
	"bytes"
	"discord-messenger/discord"
	"discord-messenger/handler"
	"fmt"
	"image/jpeg"
	"image/png"
	"log"
	"time"
)

//tvinfo.post.downloaded

var TvInfoImageUpdateModule = handler.Module{
	Pattern: "tvinfo.image.downloaded",
	Handler: tvinfoImageHandler,
}

func tvinfoImageHandler(bot discord.Bot, channel, pattern, payload string) {

	imageBytes := bytes.NewBufferString(payload)

	img, err := jpeg.Decode(bytes.NewReader(imageBytes.Bytes()))
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return
	}

	bot.SendImage(discord.TvInfo, "", buf)
}

var TvInfoUpdateModule = handler.Module{
	Pattern: "tvinfo.updated",
	Handler: tvinfoTimestampHandler,
}

func tvinfoTimestampHandler(bot discord.Bot, channel, pattern, payload string) {
	ts, err := time.Parse(time.RFC3339, payload)
	if err != nil {
		ts = time.Now()
	}

	err = bot.SendMessage(discord.TvInfo,
		fmt.Sprintf(
			"Das CampusInfo System wurde geupdated (%s)\nNeue Infos:",
			ts.Format("02 Jan 2006 15:04"),
		),
	)
	if err != nil {
		log.Println("ERR - tvinfoTimestampHandler:", err)
	}
}
