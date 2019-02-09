package crawler

import (
	"io"
	"net/url"
	"os"
	"time"

	"github.com/alecthomas/template"
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
	seenURL - Set of URLs already crawled. This ensures we do not crawl URLs which are
			  already crawled.
*/
func crawl(baseURL string, depth int, fetcher fetchers.Fetcher, urlNode *tree.URLNode, seenURL map[string]bool) *[]string {
	contextLogger := log.WithFields(log.Fields{
		"base_url": baseURL,
		"depth":    depth,
	})
	// URL is already crawled. Skip crawling it.
	if seenURL[baseURL] {
		contextLogger.Info("URL already crawled. Skipping")
		return nil
	}

	// Mark current URL as crawled
	seenURL[baseURL] = true

	// Create crawledURLs to store new list of crawled URLs
	crawledURLs := &[]string{baseURL}

	if depth < 1 {
		contextLogger.Info("Max depth reached. Skipping")
		return crawledURLs
	}

	contextLogger.Infof("Started crawling page")
	defer contextLogger.Info("Finished crawling page")

	// Get list of URLs on the given page
	urlList, err := fetcher.Fetch(baseURL, fetchers.SimpleLinkExtractor)

	if err != nil {
		contextLogger.Infof("failed to fetch URL")
		return crawledURLs
	}

	for _, url := range urlList {
		// Add new URL as child of the current node.
		childNode := urlNode.AddChild(url)

		if !isPartOfDomain(baseURL, url) {
			contextLogger.WithField("child_url", url).Info("Child URL not part of the domain. Skipping.")
			continue
		}
		newlyCrawledURLs := crawl(url, depth-1, fetcher, childNode, seenURL)
		if newlyCrawledURLs != nil {
			// Add newly crawled URLs to the list of crawled URLs
			*crawledURLs = append(*crawledURLs, *newlyCrawledURLs...)
		}
	}
	return crawledURLs
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

// StartCrawling is the main entry point for crawling.
func StartCrawling(maxDepth int, baseURL string, file *os.File) {
	start := time.Now()

	root := tree.NewNode(baseURL)
	visited := make(map[string]bool)

	crawl(baseURL, maxDepth, fetchers.NewSimpleFetcher(baseURL), root, visited)

	log.Info("Total time taken:", time.Since(start))

	root.WriteTreeToFile(file)
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
func WriteSiteMapToFile(crawledURLs []string, f io.Writer) {
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
	err = tmpl.Execute(f, crawledURLs)
	if err != nil {
		log.Error(err)
	}
}
