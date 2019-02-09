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

type seenURLs struct {
	urlMap map[string]struct{}
	urls   []string

	sync.Mutex
}

func newSeenMap() *seenURLs {
	return &seenURLs{
		urlMap: make(map[string]struct{}),
	}
}
func (s *seenURLs) Add(url string) bool {
	s.Lock()
	if _, ok := s.urlMap[url]; ok {
		// URL already present. Return false indicating the new url was already present
		s.Unlock()
		return false
	}
	s.urlMap[url] = struct{}{}
	s.urls = append(s.urls, url)
	s.Unlock()
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
	seenURL - Set of URLs already crawled. This ensures we do not crawl URLs which are
			  already crawled.
*/
func crawl(baseURL string, depth int, fetcher fetchers.Fetcher, urlNode *tree.URLNode, seenURL *seenURLs) {
	contextLogger := log.WithFields(log.Fields{
		"base_url": baseURL,
		"depth":    depth,
	})
	// seenURL.Add returns false means the URL was already seen.
	if !seenURL.Add(baseURL) {
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

	if err != nil {
		contextLogger.Infof("failed to fetch URL")
		return
	}

	for _, url := range urlList {
		// Add new URL as child of the current node.
		childNode := urlNode.AddChild(url)

		if !isPartOfDomain(baseURL, url) {
			// even if we're not crawling the URLs, mark it as seen
			seenURL.Add(url)
			contextLogger.WithField("child_url", url).Info("Child URL not part of the domain. Skipping.")
			continue
		}
		crawl(url, depth-1, fetcher, childNode, seenURL)
	}
	return
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
	seenURLs := newSeenMap()
	crawl(baseURL, maxDepth, fetchers.NewSimpleFetcher(baseURL), root, seenURLs)

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
