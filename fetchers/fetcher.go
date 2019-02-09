package fetchers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) ([]string, error)
}

type Client interface {
	Get(string) (*http.Response, error)
}

// SimpleFetcher implements Fetcher
type SimpleFetcher struct {
	client  Client
	baseURL string
}

// NewSimpleFetcher creates a new fetcher with the given base URL. It also
// creates a new http.Client with 3 seconds timeout
func NewSimpleFetcher(url string) *SimpleFetcher {
	simpleClient := http.DefaultClient
	simpleClient.Timeout = 5 * time.Second
	return &SimpleFetcher{baseURL: url, client: simpleClient}

}

func (f SimpleFetcher) Fetch(url string) ([]string, error) {
	contextLogger := log.WithField("url", url)

	resp, err := f.client.Get(url)
	if err != nil {
		contextLogger.Errorf("Failed to fetch URL: %s", err)
		return nil, fmt.Errorf("Failed to fetch URL: %s", err)
	}

	defer resp.Body.Close()
	return ConvertResponseToURLList(f.baseURL, resp.Body), nil
}

func ConvertResponseToURLList(baseURL string, body io.Reader) []string {
	contextLogger := log.WithField("base_url", baseURL)

	var urlList []string
	// URLset is used to ensure urlList is always unique
	URLset := make(map[string]struct{})
	tokenizer := html.NewTokenizer(body)
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return urlList
		case html.StartTagToken:
			token := tokenizer.Token()
			// Token isn't a <a> tag. Skip token and continue the loop
			if token.DataAtom.String() != "a" {
				continue
			}
			href := findHrefValue(token)
			if href == nil {
				continue
			}
			// If we've already added this URL to urlList, don't add it again
			if _, ok := URLset[*href]; ok {
				continue
			}

			// add URL to URLset
			URLset[*href] = struct{}{}
			builtURL, err := buildURL(baseURL, *href)
			if err != nil {
				// error occurred while trying to build the URL. Log the error
				// and continue.
				contextLogger.Infof("Failed to build URL: %s", err)
				continue
			}

			// if the new url is equal to the baseURL don't add it
			if builtURL == baseURL {
				continue
			}

			urlList = append(urlList, builtURL)
		}
	}
}

func findHrefValue(t html.Token) *string {
	for _, attr := range t.Attr {
		if attr.Key == "href" || attr.Key == "HREF" {
			return &attr.Val
		}
	}
	return nil
}

func buildURL(baseURL string, href string) (string, error) {
	// Discard invalid href values. Return error
	if href == "" ||
		len(href) == 1 ||
		strings.HasPrefix(href, "#") {
		return baseURL, nil
	}

	u, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	// Remove all query params
	u.RawQuery = ""
	// Remove fragment, if any
	u.Fragment = ""
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(u).String(), nil
}
