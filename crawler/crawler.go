package crawler

import (
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/jarifibrahim/webcrawler/fetchers"
	"github.com/jarifibrahim/webcrawler/tree"
	log "github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

/*
crawl fetches the list of URLs using the fetcher and crawls the new list of
URLs in a depth-first-search manner. Returns the list of URLs crawled.
Params:
	baseURL - The URL to crawl
	fetcher - A fetcher used to fetch the list of URLs to crawl next
	urlNode - urlNode is used to build the tree when --show-tree flag is set.
      	      urlNode stores a tree of link. Where root of the tree is the base URL and
			  all the links reachable from root are stored at it's children
	state   - state stores the global state of crawling. It contains the set of URLs
		      already crawled.
*/
func crawl(baseURL string, depth int, fetcher fetchers.Fetcher, urlNode *tree.URLNode, state *CrawlerState) {
	contextLogger := log.WithFields(log.Fields{
		"base_url": baseURL,
		"depth":    depth,
	})

	defer wg.Done()

	// state.AddURL() returns false if the URL was already seen.
	if !state.AddURL(baseURL) {
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

	state.IncrementCrawledCount()

	if err != nil {
		contextLogger.Infof("failed to fetch URL")
		return
	}

	for _, url := range urlList {
		// Add new URL as child of the current node.
		childNode := urlNode.AddChild(url)

		if !isPartOfDomain(baseURL, url) {
			// even if we're not crawling the URL, mark it as seen
			state.AddURL(url)
			contextLogger.WithField("child_url", url).Info("Child URL not part of the domain. Skipping.")
			continue
		}
		wg.Add(1)
		go crawl(url, depth-1, fetcher, childNode, state)
	}
}

// isPartOfDomain checks if the baseURL and urlToCheck belong to the same domain
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

// StartCrawling is the main entry point for crawling. It crawls the given URL
// in depth first search manner.
//
// params:
//	BaseURL: 		It is the starting URL for the crawler
// 	MaxDepth: 		It determines the depth of depth-first-search
// 	ShowTree: 		It determines if the tree should be generated for the crawled pages
//  TreeWriter:		If showTree is true, the tree is written to treeWriter
// 	SiteMapWriter:	The xml sitemap is written to the sitemapwriter
func StartCrawling(baseURL string, maxDepth int, showTree bool, treeWriter, siteMapWriter io.Writer) {
	start := time.Now()
	var root *tree.URLNode
	if showTree {
		root = tree.NewNode(baseURL)
	}

	crawlerState := NewCrawlerState()

	wg.Add(1)
	go crawl(baseURL, maxDepth, fetchers.NewSimpleFetcher(baseURL), root, crawlerState)
	wg.Wait()

	log.Info("Total URLs found:", crawlerState.seenURLCount)
	log.Info("Total URLs crawled:", crawlerState.crawledURLCount)
	log.Info("Total time taken:", time.Since(start))

	crawlerState.WriteSiteMap(siteMapWriter)

	if showTree {
		root.WriteTree(treeWriter)
	}

}
