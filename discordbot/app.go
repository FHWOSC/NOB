package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	SuppressStart bool
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&TargetChannel, "c", "", "Target Channel")
	flag.StringVar(&AdminUserId, "a", "", "Admin Channel")
	flag.BoolVar(&SuppressStart, "silent", false, "Suppress start messages")
	flag.Parse()
}

func main() {
	bot, err := NewBot(Token, AdminUserId)
	if err != nil {
		log.Panicln(err)
	}

	if !SuppressStart {
		bot.SendMessage(TargetChannel, "Moin! Ich bin Nob, und ich sage euch bescheid wenn sich der Vorlesungsplan verändert!")
	}

	go func() {
		err := StartSplanChecks(
			func(v interface{}) {
				err := bot.SendMessage(TargetChannel, v)
				if err != nil {

					err := bot.SendDirectMessage(AdminUserId, err)
					if err != nil {
						log.Fatalln(err)
					}
				}
			},
			func(v interface{}) {
				err := bot.SendDirectMessage(AdminUserId, v)
				if err != nil {
					log.Fatalln(err)
				}
			},
			5*time.Minute)
		if err != nil {
			bot.SendDirectMessage(AdminUserId, err)
			log.Panicln(err)
		}
	}()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	err = bot.SendMessage(AdminUserId, "I'll be back")
	if err != nil {
		log.Println(err)
	}
	bot.Close()
}
