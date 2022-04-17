package tvinfo

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"time"
	"tvinfo/parser"
)

func GetTvInfo() (*time.Time, []string, error) {
	doc, err := parser.GetDoc("https://intern.fh-wedel.de/scala/?no_cache=1")
	if err != nil {
		return nil, nil, err
	}

	var ts time.Time
	doc.Find(".tx-fhwscala-pi1 div").First().Each(func(i int, s *goquery.Selection) {
		text := strings.Trim(s.Text(), " \n")
		ts, err = time.Parse("Die Inhalte des CampusInfo-Systems wurden zuletzt am 02.04.2006 um 15:04 Uhr geändert.", text)
		if err != nil {
			return
		}
	})

	urls := make([]string, 0)
	doc.Find("#col2 .clearfix img").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("src")
		if exists {
			urls = append(urls, "https://intern.fh-wedel.de/"+url)
		}
	})

	return &ts, urls, nil
}

func GetImages(urls []string) ([]*bytes.Buffer, error) {
	imgs := make([]*bytes.Buffer, 0)

	for _, url := range urls {
		buf, err := parser.Get(url)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if buf.Len() > 5000 {
			log.Println("downloaded img")
			imgs = append(imgs, buf)
			//fmt.Printf("IMGBUF size=%d | hash=%x\n", buf.Len(), md5.Sum(buf.Bytes()))
		}
	}

	return imgs, nil
}
