package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"time"
	"tvinfo/event"
	"tvinfo/tvinfo"
)

func main() {
	log.Println("running...")

	//p := parser.New()

	broker := event.NewBroker(
		os.Getenv("MESSAGE_BROKER_ADDR"),
		os.Getenv("MESSAGE_BROKER_PASS"),
	)
	defer broker.Close()

	hashes := make(map[string]bool)

	var timestamp *time.Time
	timestamp = nil

	for {
		log.Println("start")
		ts := runScan(broker, hashes, timestamp)
		if timestamp == nil && ts != nil {
			timestamp = ts
		}

		log.Println("waiting")
		time.Sleep(5 * time.Minute)
	}

}

func runScan(broker *event.Broker, hashes map[string]bool, timestamp *time.Time) *time.Time {
	ts, urls, err := tvinfo.GetTvInfo()
	if err != nil {
		return nil
	}

	firstRun := timestamp == nil

	if firstRun || ts.After(*timestamp) {
		buffers, err := tvinfo.GetImages(urls)
		if err == nil {
			log.Println("sending new images")

			broker.Publish("tvinfo.updated", ts.Format(time.RFC3339))

			for _, buf := range buffers {
				hash := fmt.Sprintf("%x", md5.Sum(buf.Bytes()))

				_, exists := hashes[hash]

				log.Printf("Found (new=%t) Image: %s\n", !exists, hash)
				if !exists {
					broker.Publish("tvinfo.image.downloaded", buf.String())
					hashes[hash] = true
				} else {
					log.Println("img is known")
				}
			}
		}
	} else {
		log.Println("timestamp didn't change")
	}

	if firstRun {
		log.Println("first run", ts)
		return ts
	}
	return nil
}
