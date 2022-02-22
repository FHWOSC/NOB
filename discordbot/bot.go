package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
const (
	splanUrl              = "https://intern.fh-wedel.de/~splan/"
	splanTimestampEnvName = "SPLAN_GENERATED_AT"
)

var (
	Token         string
	TargetChannel string
	AdminUserId   string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&TargetChannel, "c", "", "Target Channel")
	flag.StringVar(&AdminUserId, "a", "", "Target Channel")
	flag.Parse()
}

func GetDoc(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func GetSPlanTimestamp() (time.Time, error) {
	doc, err := GetDoc("https://intern.fh-wedel.de/~splan/index.html?typ=benutzer_vz")
	if err != nil {
		return time.Time{}, err
	}

	// get timestamp of splan
	generatedAtStr := doc.Find("div.splan_version").First().Text()
	parts := strings.Split(generatedAtStr, "by")
	generatedAt, err := time.Parse("02.01.2006, 15:04 ", parts[0][len("  © generated at "):])
	return generatedAt, err
}

func CheckTimestamp() (bool, *time.Time, error) {
	current, err := GetSPlanTimestamp()
	if err != nil {
		return false, nil, err
	}
	log.Println("current:", current)

	env := os.Getenv(splanTimestampEnvName)
	err = UpdateTimestamp(current)
	if env == "" || err != nil {
		return false, &current, err
	}

	before, err := time.Parse(time.RFC3339, env)
	if err != nil {
		return false, &current, err
	}
	log.Println("before:", before)

	changed := !current.Equal(before)
	log.Println("Vorlesungsplan changed:", changed)
	return changed, &current, nil
}

func UpdateTimestamp(timestamp time.Time) error {
	err := os.Setenv(splanTimestampEnvName, timestamp.Format(time.RFC3339))
	return err
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	sendToAdmin := func(in interface{}) {
		err := SendErrorToDiscordUser(dg, AdminUserId, in)
		if err != nil {
			log.Fatalln(err)
		}
	}

	sendToChannel := func(chId string, msg interface{}) {
		_, err = dg.ChannelMessageSend(
			TargetChannel,
			fmt.Sprintf("%s", msg))
		if err != nil {
			sendToAdmin(err)
		}
	}

	sendToChannel(TargetChannel, "Moin! Ich bin Nob, und ich sage euch bescheid wenn sich der Vorlesungsplan verändert!")

	go func() {
		for {
			changed, ts, err := CheckTimestamp()
			if err != nil {
				sendToAdmin(err)
			}
			if changed {
				sendToChannel(
					TargetChannel,
					fmt.Sprintf(
						"@everyone\n Neuer Vorlesungsplan: %s\n(%s)",
						splanUrl,
						ts.Format("02 Jan 2006 15:04"),
					))
				sendToAdmin("Neuer Vorlesungsplan online!")
			} else {
				log.Println("Vorlesungsplan immer noch der alte")
			}

			time.Sleep(5 * time.Minute)
		}
	}()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	sendToAdmin("I'll be back")
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func SendErrorToDiscordUser(s *discordgo.Session, userId string, in interface{}) error {
	channel, err := s.UserChannelCreate(userId)
	if err != nil {
		// If an error occurred, we failed to create the channel.
		//
		// Some common causes are:
		// 1. We don't share a server with the user (not possible here).
		// 2. We opened enough DM channels quickly enough for Discord to
		//    label us as abusing the endpoint, blocking us from opening
		//    new ones.
		return err
	}

	// Then we send the message through the channel we created.
	_, err = s.ChannelMessageSend(channel.ID, fmt.Sprintf("%s", in))
	if err != nil {
		// If an error occurred, we failed to send the message.
		//
		// It may occur either when we do not share a server with the
		// user (highly unlikely as we just received a message) or
		// the user disabled DM in their settings (more likely).
		log.Println("error sending DM message:", err)
		return err
	}

	return nil
}
