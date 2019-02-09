package crawler

import (
	"io"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/alecthomas/template"
	"github.com/jarifibrahim/webcrawler/fetchers"
	"github.com/jarifibrahim/webcrawler/tree"
	log "github.com/sirupsen/logrus"
)

type urlCache struct {
	urlMap          map[string]struct{} // urlMap is used for fast lookup. It is used to ensure we don't crawl a URL twice
	urls            []string            // urls stores the actual list of URLs seen
	seenURLCount    int                 // seenURLCount stores the number of URLs. seenURLCount will always be less than or equal to crawledURLCoun
	crawledURLCount int                 // The number of seen URLs is not equal to the number of crawled URLs. This variable stores the value of crawled URLs
	sync.Mutex
}

func NewURLCache() *urlCache {
	return &urlCache{
		urlMap: make(map[string]struct{}),
	}
}

func (c *urlCache) IncrementCrawledCount() {
	c.Lock()
	c.crawledURLCount++
	c.Unlock()
}

func (c *urlCache) Add(url string) bool {
	c.Lock()
	if _, ok := c.urlMap[url]; ok {
		// URL already present. Return false indicating the new url was already present
		c.Unlock()
		return false
	}
	c.urlMap[url] = struct{}{}
	c.seenURLCount++
	c.urls = append(c.urls, url)
	c.Unlock()
	return true
}

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
func crawl(baseURL string, depth int, fetcher fetchers.Fetcher, urlNode *tree.URLNode, cache *urlCache) {
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

// WriteSiteMapToFile generetes sitemap from the given list of URLs
// The sitemap is minimal and contains only the mandatory <loc> field
// Sample sitemap
//
// <?xml version="1.0" encoding="UTF-8"?>
//
// <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
//   <url>
//     <loc>http://foo.com</loc>
//   </url>
// </urlset>
func (c *urlCache) WriteSiteMapToFile(f io.Writer) {
	xmlTemplate := `
<?xml version="1.0" encoding="UTF-8"?>

<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
{{range $element := . }}
  <url>
    <loc>{{$element}}</loc>
  </url>{{end}}
</urlset>
`
	tmpl, err := template.New("test").Parse(xmlTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	err = tmpl.Execute(f, c.urls)
	if err != nil {
		log.Error(err)
	}
}
