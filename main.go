package main

import (
	"os"

	"github.com/jarifibrahim/webcrawler/crawler"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	crawler.StartCrawling(3, "http://jarifibrahim.github.io", os.Stdout)
}
