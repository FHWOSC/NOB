package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/bwmarrin/discordgo"
	"io"
	"log"
	"os"
)

type Channel string

const (
	Log          Channel = "CHID_LOG"
	TvInfo       Channel = "CHID_TVINFO"
	Announcement Channel = "CHID_ANNOUNCEMENT"
)

type Bot interface {
	SendMessage(channel Channel, message string) error
	SendImage(channel Channel, message string, img io.Reader) error
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

	setId(b.channelIDs, Log, os.Getenv(string(Log)))
	setId(b.channelIDs, TvInfo, os.Getenv(string(TvInfo)))
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

	log.Println("trying to send discord message to channel", string(channel), "=>", id)
	_, err = b.session.ChannelMessageSend(id, message)
	if err != nil {
		return err
	}

	return nil
}

func (b *bot) SendImage(channel Channel, message string, img io.Reader) error {
	id, err := b.getId(channel)
	if err != nil {
		return err
	}

	log.Println("trying to send discord message with image to channel", string(channel), "=>", id)
	msg := &discordgo.MessageSend{
		File: &discordgo.File{
			Name:        "image.png",
			ContentType: "image/png",
			Reader:      img,
		},
		Content: message,
	}
	_, err = b.session.ChannelMessageSendComplex(id, msg)
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

	log.Println(string(channel), "=", id)

	channels[channel] = id
}
