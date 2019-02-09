package crawler

import (
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/jarifibrahim/webcrawler/fetchers"
	"github.com/jarifibrahim/webcrawler/tree"
	log "github.com/sirupsen/logrus"
)

/*
crawl fetches the list of URLs using the fetcher and crawls the new list of
URLs in a depth-first-search manner. Returns the list of URLs crawled.
Params:
	baseURL - The URL to crawl
	fetcher - A fetcher used to fetch the list of URLs to crawl next
	urlNode - urlNode is used to build the tree when --show-tree flag is set.
      	      urlNode stores a tree of link. Where root of the tree is the base URL and
			  all the links reachable from root are stored at it's children
	cache  	- Set of URLs already crawled. This ensures we do not crawl URLs which are
			  already crawled.
*/
func crawl(baseURL string, depth int, fetcher fetchers.Fetcher, urlNode *tree.URLNode, cache *URLCache) {
	contextLogger := log.WithFields(log.Fields{
		"base_url": baseURL,
		"depth":    depth,
	})

	defer wg.Done()

	// cache.Add() returns false if the URL was already seen.
	if !cache.Add(baseURL) {
		contextLogger.Info("URL already crawled. Skipping")
		return
	}

	if depth < 1 {
		contextLogger.Info("Max depth reached. Skipping")
		return
	}

	contextLogger.Infof("Started crawling page")
	defer contextLogger.Info("Finished crawling page")

	// Get list of URLs on the given page
	urlList, err := fetcher.Fetch(baseURL, fetchers.SimpleLinkExtractor)

	cache.IncrementCrawledCount()

	if err != nil {
		contextLogger.Infof("failed to fetch URL")
		return
	}

	for _, url := range urlList {
		// Add new URL as child of the current node.
		childNode := urlNode.AddChild(url)

		if !isPartOfDomain(baseURL, url) {
			// even if we're not crawling the URL, mark it as seen
			cache.Add(url)
			contextLogger.WithField("child_url", url).Info("Child URL not part of the domain. Skipping.")
			continue
		}
		wg.Add(1)
		go crawl(url, depth-1, fetcher, childNode, cache)
	}
}

func isPartOfDomain(baseURL, urlToCheck string) bool {
	base, err := url.Parse(baseURL)
	if err != nil {
		log.Error(err)
		return false
	}
	testURL, err := url.Parse(urlToCheck)
	if err != nil {
		log.Error(err)
		return false
	}
	return base.Host == testURL.Host

}

var wg sync.WaitGroup

// StartCrawling is the main entry point for crawling.
func StartCrawling(maxDepth int, baseURL string, showTree bool, treeFile, siteMapFile *os.File) {
	start := time.Now()
	var root *tree.URLNode
	if showTree {
		root = tree.NewNode(baseURL)
	}

	urlCache := NewURLCache()

	wg.Add(1)
	go crawl(baseURL, maxDepth, fetchers.NewSimpleFetcher(baseURL), root, urlCache)
	wg.Wait()

	log.Info("Total URLs found:", urlCache.seenURLCount)
	log.Info("Total URLs Crawled:", urlCache.crawledURLCount)
	log.Info("Total time taken:", time.Since(start))

	urlCache.WriteSiteMapToFile(siteMapFile)

	if showTree {
		root.WriteTreeToFile(treeFile)
	}

}
