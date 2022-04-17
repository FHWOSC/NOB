package main

import (
	"log"
	"os"
	"splan/event"
	"splan/parser"
	"time"
)

const (
	UpdateChannel       = "splan.timestamp.changed"
	DefaultScanInterval = "5m"
)

func main() {
	log.Println("running...")

	p := parser.New()
	//ts, err := time.Parse("02.01.2006, 15:04", "13.03.2022, 08:24")
	//if err != nil {
	//	panic(err)
	//}
	//p.SetTimestamp(ts)

	broker := event.NewBroker(
		os.Getenv("MESSAGE_BROKER_ADDR"),
		os.Getenv("MESSAGE_BROKER_PASS"),
	)
	defer broker.Close()

	Log := func(v ...any) {
		log.Println(v...)

		//broker.Publish("splan.logged", fmt.Sprint(v...))
	}

	for {
		func() {
			defer sleep()
			log.Println("scanning splan for changes")

			ts, changed, err := p.GetTimestamp()
			if err != nil {
				Log("error while :", err)
				return
			}

			if !changed {
				Log("timestamp didn't change")
				return
			}

			Log("timestamp changed!", ts)
			err = broker.Publish(UpdateChannel, ts.Format(time.RFC3339))
			if err != nil {
				Log("error while trying to publish message:", err)
				return
			}
		}()
	}
}

func sleep() {
	env := os.Getenv("SPLAN_SCAN_INTERVAL")
	if env == "" {
		env = DefaultScanInterval
	}

	duration, err := time.ParseDuration(env)
	if err != nil {
		log.Panicln(err)
	}

	log.Println("Sleeping for", duration)
	time.Sleep(duration)
}
