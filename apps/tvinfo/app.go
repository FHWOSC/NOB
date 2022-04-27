package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"time"
	"tvinfo/broker"
	"tvinfo/persistance"
	"tvinfo/tvinfo"
)

var TimeStamp *time.Time

func main() {
	log.Println("running...")

	if os.Getenv("TVINFO_TIMESTAMP") != "" {
		ts, err := time.Parse(time.RFC3339, os.Getenv("TVINFO_TIMESTAMP"))
		if err != nil {
			TimeStamp = nil
		}

		TimeStamp = &ts
	}

	for {
		func() {
			defer Sleep()
			log.Println("Checking if campus-info timestamp has changed")
			hasChanged, timestamp, imgUrls, err := HasChanged()
			if err != nil {
				log.Println("Check failed:", err)
				return
			}

			if !hasChanged {
				log.Println("campus-info timestamp hasn't changed")
				return
			} else {
				log.Println("campus-info timestamp has changed. New timestmap:", timestamp.Format(time.RFC3339))
			}

			err = PublishChangedImgs(timestamp, imgUrls...)
			if err != nil {
				log.Println("error while trying to publish images:", err)
			}
		}()
	}

}

func Sleep() {
	sleepDuration, err := time.ParseDuration(os.Getenv("INTERVAL"))
	if err != nil {
		sleepDuration = 5 * time.Minute
	}

	log.Println("Going to sleep for", sleepDuration)
	time.Sleep(sleepDuration)
}

func HasChanged() (bool, *time.Time, []string, error) {
	ts, urls, err := tvinfo.GetTvInfo()
	if err != nil {
		return false, nil, nil, err
	}

	// first scan
	if TimeStamp == nil {
		TimeStamp = ts
		return false, ts, []string{}, nil
	}

	if ts.After(*TimeStamp) {
		TimeStamp = ts
		return true, ts, urls, nil
	} else {
		if !ts.Equal(*TimeStamp) {
			log.Printf("WARNING: Timestamp mismatch: [ts=%v] [TimeStamp=%v]\n", ts, TimeStamp)
		}
	}
	return false, ts, []string{}, nil
}

func PublishChangedImgs(newTimestamp *time.Time, urls ...string) error {
	sendUpdatedTimestamp := false

	log.Println("downloading campus-info images")
	images, err := tvinfo.GetImages(urls)
	if err != nil {
		return err
	}

	for i, image := range images {
		hash := Hash(image)

		contains, err := persistance.Contains(hash)
		if err != nil {
			log.Fatal(err)
		}

		if contains {
			log.Printf("%2d. Hash is known: %s\n", i, hash)
			continue
		}
		log.Printf("%2d. Hash is UNKNOWN: %s\n", i, hash)

		err = persistance.Append(hash)
		if err != nil {
			log.Fatal(err)
		}

		if !sendUpdatedTimestamp {
			broker.Publish("tvinfo.updated", newTimestamp.Format(time.RFC3339))
			sendUpdatedTimestamp = true
		}

		broker.Publish("tvinfo.image.downloaded", image.String())
	}

	if !sendUpdatedTimestamp {
		log.Println("suppressed timestamp updated message, because no new img(hash) was detected")
	}

	return nil
}

func Hash(buf *bytes.Buffer) string {
	return fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
}
