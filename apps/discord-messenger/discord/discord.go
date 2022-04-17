package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/bwmarrin/discordgo"
	"os"
)

type Channel string

const (
	FHW          Channel = "CHID_FHW"
	Announcement Channel = "CHID_ANNOUNCEMENT"
)

type Bot interface {
	SendMessage(channel Channel, message string) error
	Close()
}

type bot struct {
	session    *discordgo.Session
	channelIDs map[Channel]string
}

func New(token string) (*bot, error) {
	b := new(bot)
	b.channelIDs = make(map[Channel]string)

	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	err = dg.Open()
	if err != nil {
		return nil, err
	}

	b.session = dg

	setId(b.channelIDs, FHW, os.Getenv(string(FHW)))
	setId(b.channelIDs, Announcement, os.Getenv(string(Announcement)))

	return b, nil
}

func (b *bot) Close() {
	b.session.Close()
}

func (b *bot) SendMessage(channel Channel, message string) error {
	id, err := b.getId(channel)
	if err != nil {
		return err
	}

	_, err = b.session.ChannelMessageSend(id, message)
	if err != nil {
		return err
	}

	return nil
}

func (b *bot) getId(channel Channel) (string, error) {
	id, exists := b.channelIDs[channel]
	if !exists {
		return "", fmt.Errorf("channel has no associated id")
	}

	return id, nil
}

func setId(channels map[Channel]string, channel Channel, id string) {
	if id == "" {
		return
	}
	if channels == nil {
		return
	}
	if string(channel) == "" {
		return
	}

	channels[channel] = id
}
