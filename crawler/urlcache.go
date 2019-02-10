package crawler

import (
	"io"
	"sync"

	"github.com/alecthomas/template"
	log "github.com/sirupsen/logrus"
)

// CrawlerState stores the global state of crawling. It is go rountine safe.
type CrawlerState struct {
	urlMap          map[string]struct{} // urlMap is used for fast lookup. It is used to ensure we don't crawl a URL twice
	urls            []string            // urls stores the actual list of URLs seen
	seenURLCount    int                 // seenURLCount stores the number of URLs. seenURLCount will always be less than or equal to crawledURLCoun
	crawledURLCount int                 // The number of seen URLs is not equal to the number of crawled URLs. This variable stores the value of crawled URLs
	sync.Mutex
}

// NewCrawlerState returns a new CrawlerState
func NewCrawlerState() *CrawlerState {
	return &CrawlerState{
		urlMap: make(map[string]struct{}),
	}
}

// IncrementCrawledCount increases the crawled URL count by 1
func (c *CrawlerState) IncrementCrawledCount() {
	c.Lock()
	c.crawledURLCount++
	c.Unlock()
}

// AddURL tries to insert the new url into the global URL cache.
// Returns false if URL was already present and true if not.
func (c *CrawlerState) AddURL(url string) bool {
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

// WriteSiteMap generates sitemap from the given list of URLs
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
func (c *CrawlerState) WriteSiteMap(f io.Writer) {
	xmlTemplate := `<?xml version="1.0" encoding="UTF-8"?>

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
