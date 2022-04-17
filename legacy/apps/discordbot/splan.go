package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func StartSplanChecks(onChange func(interface{}), onError func(interface{}), cooldown time.Duration) error {
	for {
		changed, ts, err := CheckTimestamp()
		if err != nil {
			onError(err)
		}
		if changed {
			onChange(fmt.Sprintf(
				"@everyone\n Neuer Vorlesungsplan: %s\n(%s)", //
				splanUrl,
				ts.Format("02 Jan 2006 15:04"),
			))
		} else {
			log.Println("splan check completed: splan didn't change")
		}

		time.Sleep(cooldown)
	}
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

//SPLAN_GENERATED_AT=2021-01-04T13:00:00Z
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
