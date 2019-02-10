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
	sitemapFileName := flag.String("sitemap-file-name", "sitemap.xml", "File to write sitemap")
	showTree := flag.Bool("show-tree", true, "Show links between pages")
	treeFileName := flag.String("tree-file-name", "url-tree.txt", "File to write the generated tree")
	flag.Parse()

	siteMapFile, err := os.Create(*sitemapFileName)
	if err != nil {
		log.Fatal(err)
	}
	var treeFile *os.File
	if *showTree {
		treeFile, err = os.Create(*treeFileName)
		if err != nil {
			log.Fatal(err)
		}
	}

	crawler.StartCrawling(*baseURL, *maxDepth, *showTree, treeFile, siteMapFile)
}
