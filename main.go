package main

import (
	"flag"
	"os"

	"github.com/jarifibrahim/webcrawler/crawler"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	baseURL := flag.String("baseurl", "http://jarifibrahim.github.io", "Base URL to crawl")
	maxDepth := flag.Int("max-depth", 2, "Max Depth to crawl")
	output := flag.String("output-file", "Stdout", "File to write output")

	var fileToWrite *os.File
	if *output == "Stdout" {
		fileToWrite = os.Stdout
	} else {
		var err error
		fileToWrite, err = os.Create(*output)
		if err != nil {
			log.Fatal(err)
		}
	}
	flag.Parse()

	crawler.StartCrawling(*maxDepth, *baseURL, fileToWrite)
}
