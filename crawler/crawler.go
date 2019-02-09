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

func crawl(baseUrl string, depth int, fetcher fetchers.Fetcher, urlNode *tree.URLNode, seenURL map[string]bool) *[]string {
	contextLogger := log.WithFields(log.Fields{
		"base_url": baseUrl,
		"depth":    depth,
	})
	if seenURL[baseUrl] {
		contextLogger.Info("URL already crawled. Skipping")
		return nil
	}

	// Mark current URL as visited
	seenURL[baseUrl] = true

	// Create crawledURLs to store returned URLs
	crawledURLs := &[]string{baseUrl}

	if depth < 1 {
		contextLogger.Info("Max depth reached. Skipping")
		return crawledURLs
	}

	contextLogger.Infof("Started crawling page")
	defer contextLogger.Info("Finished crawling page")

	urlList, err := fetcher.Fetch(baseUrl, fetchers.SimpleLinkExtractor)

	if err != nil {
		contextLogger.Infof("failed to fetch URL")
		return crawledURLs
	}

	for _, url := range urlList {
		// Add new URL as child of the current node.
		childNode := urlNode.AddChild(url)

		if !isPartOfDomain(baseURL, url) {
			contextLogger.WithField("child_url", url).Info("Child URL not part of domain. Skipping.")
			continue
		}
		newlyCrawledURLs := crawl(url, depth-1, fetcher, childNode, seenURL)
		if newlyCrawledURLs != nil {
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
		return true
	}
	return base.Host == testURL.Host

}

func StartCrawling(maxDepth int, baseURL string, file *os.File) {
	start := time.Now()
	root := tree.NewNode(baseURL)
	visited := make(map[string]bool)
	crawl(baseURL, maxDepth, fetchers.NewSimpleFetcher(baseURL), root, visited)
	log.Info("Total time taken:", time.Since(start))
	root.WriteTreeToFile(file)

}

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
	}
	err = tmpl.Execute(f, crawledURLs)
	if err != nil {
		log.Error(err)
	}
}
