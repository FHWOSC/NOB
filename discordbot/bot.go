package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session
	adminId string

	__userChannels map[string]*discordgo.Channel
}

func NewBot(discordToken, adminUserId string) (*Bot, error) {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return nil, err
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		return nil, err
	}

	bot := new(Bot)
	bot.session = dg
	bot.adminId = adminUserId
	bot.__userChannels = make(map[string]*discordgo.Channel)

	return bot, nil
}

func (b *Bot) Log(v ...interface{}) {
	log.Debug(v...)
	err := b.SendDirectMessage(b.adminId, fmt.Sprintf("%s", v))
	if err != nil {
		log.Critical(err)
		panic(err)
	}
}

func (b *Bot) SendMessage(channelId string, msg interface{}) error {
	_, err := b.session.ChannelMessageSend(
		channelId,
		fmt.Sprintf("%s", msg))
	return err
}

func (b *Bot) SendDirectMessage(userId string, msg interface{}) error {
	channel, err := b.openUserChannel(userId)
	if err != nil {
		log.Error("error opening DM channel:", err)
		return err
	}

	return b.SendMessage(channel.ID, msg)
}

func (b *Bot) Close() {
	b.session.Close()
}

func (b *Bot) openUserChannel(userId string) (*discordgo.Channel, error) {
	// Open channel to user
	channel, err := b.session.UserChannelCreate(userId)
	if err != nil {
		// If an error occurred, we failed to create the channel.
		//
		// Some common causes are:
		// 1. We don't share a server with the user (not possible here).
		// 2. We opened enough DM channels quickly enough for Discord to
		//    label us as abusing the endpoint, blocking us from opening
		//    new ones.
		return nil, err
	}

	return channel, nil
}
