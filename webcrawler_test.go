package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jarifibrahim/webcrawler/tree"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Disable logger. We don't want noisy logs when running tests
func init() {
	logrus.SetLevel(logrus.PanicLevel)

}
func TestWriteSiteMapToFile(t *testing.T) {
	var writeBuffer bytes.Buffer

	seenLinks := []string{"/foo", "/bar", "/helloWorld"}
	expectedOutput := `
<?xml version="1.0" encoding="UTF-8"?>

<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">

  <url>
    <loc>/foo</loc>
  </url>
  <url>
    <loc>/bar</loc>
  </url>
  <url>
    <loc>/helloWorld</loc>
  </url>
</urlset>
`
	// Write sitemap to writeBuffer
	WriteSiteMapToFile(seenLinks, &writeBuffer)
	assert.Equal(t, expectedOutput, writeBuffer.String())
}

func TestIsPartOfDomain(t *testing.T) {
	testData := []struct {
		name           string
		baseURL        string
		urlToCheck     string
		expectedOutput bool
	}{
		{"same domain", "http://foo.com", "http://foo.com", true},
		{"same domain, different path", "http://foo.com", "http://foo.com/bar", true},
		{"different domain 1", "http://foo.com", "http://bar.com", false},
		{"different domain 2", "http://foo.com", "http:/foo.org", false},
		{"different schema", "https://foo.com", "http://foo.com", true},
		{"with userinfo", "http://foo.com", "https://ibrahim@foo.com", true},
		{"with fragment", "http://foo.com", "http://foo.com/#content", true},
		{"with query", "http://foo.com", "http://foo.com/bar?hello=world", true},
	}

	for _, tt := range testData {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, isPartOfDomain(tt.baseURL, tt.urlToCheck), tt.expectedOutput)
		})
	}
}

func TestCrawlDepth0(t *testing.T) {
	expectedURLList := []string{"https://g.org/"}
	t.Run("without Tree", func(t *testing.T) {
		visited := make(map[string]bool)
		crawledURLs := Crawl("https://g.org/", 0, fetcher, nil, visited)
		assert.Equal(t, expectedURLList, *crawledURLs)
	})
	t.Run("with Tree", func(t *testing.T) {
		visited := make(map[string]bool)
		rootNode := tree.NewNode("https://g.org/")
		crawledURLs := Crawl("https://g.org/", 0, fetcher, rootNode, visited)
		assert.Equal(t, expectedURLList, *crawledURLs)
		// A tree with depth 0 is only the root node
		expectedTree := tree.NewNode("https://g.org/")
		assert.Equal(t, expectedTree, rootNode)
	})
}
func TestCrawlDepth1(t *testing.T) {
	expectedURLList := []string{"https://g.org/", "https://g.org/pkg/", "https://g.org/cmd/"}
	t.Run("without tree", func(t *testing.T) {
		visited := make(map[string]bool)
		crawledURLs := Crawl("https://g.org/", 1, fetcher, nil, visited)
		assert.Equal(t, expectedURLList, *crawledURLs)

	})
	t.Run("with tree", func(t *testing.T) {
		visited := make(map[string]bool)
		rootNode := tree.NewNode("https://g.org/")
		crawledURLs := Crawl("https://g.org/", 1, fetcher, rootNode, visited)
		assert.Equal(t, expectedURLList, *crawledURLs)

		expectedTree := tree.NewNode("https://g.org/")
		for _, url := range expectedURLList[1:] {
			expectedTree.AddChild(url)
		}
		assert.Equal(t, expectedTree, rootNode)
	})
	t.Run("non existent URL", func(t *testing.T) {
		t.Run("without tree", func(t *testing.T) {
			visited := make(map[string]bool)
			crawledURLs := Crawl("https://foo.org/", 1, fetcher, nil, visited)
			assert.Equal(t, []string{"https://foo.org/"}, *crawledURLs)
		})
	})
}
func TestCrawlDepth2(t *testing.T) {
	expectedURLs := []string{
		"https://g.org/", "https://g.org/pkg/", "https://g.org/cmd/",
		"https://g.org/pkg/fmt/", "https://g.org/pkg/os/"}
	t.Run("without tree", func(t *testing.T) {
		visited := make(map[string]bool)
		crawledURLs := Crawl("https://g.org/", 2, fetcher, nil, visited)
		assert.Equal(t, expectedURLs, *crawledURLs)
	})
	t.Run("with tree", func(t *testing.T) {
		visited := make(map[string]bool)
		rootNode := tree.NewNode("https://g.org/")
		crawledURLs := Crawl("https://g.org/", 2, fetcher, rootNode, visited)
		assert.Equal(t, expectedURLs, *crawledURLs)

		expectedTree := tree.NewNode("https://g.org/")
		child, err := expectedTree.AddChild("https://g.org/pkg/")
		assert.Nil(t, err)
		child.AddChild("https://g.org/")
		child.AddChild("https://g.org/cmd/")
		child.AddChild("https://g.org/pkg/fmt/")
		child.AddChild("https://g.org/pkg/os/")
		expectedTree.AddChild("https://g.org/cmd/")
		assert.Equal(t, expectedTree, rootNode)
	})
}
func TestCrawlDepth3(t *testing.T) {
	depth3ExpectedURLs := []string{
		"https://g.org/", "https://g.org/pkg/", "https://g.org/cmd/",
		"https://g.org/x/tools", "https://g.org/net/http", "https://g.org/net/html",
		"https://g.org/pkg/fmt/", "https://g.org/pkg/os/"}

	t.Run("without tree", func(t *testing.T) {
		visited := make(map[string]bool)
		crawledURLs := Crawl("https://g.org/", 3, fetcher, nil, visited)
		assert.Equal(t, depth3ExpectedURLs, *crawledURLs)
	})
	t.Run("with tree", func(t *testing.T) {
		visited := make(map[string]bool)
		rootNode := tree.NewNode("https://g.org/")
		crawledURLs := Crawl("https://g.org/", 3, fetcher, rootNode, visited)
		assert.Equal(t, depth3ExpectedURLs, *crawledURLs)

		expectedTree := tree.NewNode("https://g.org/")

		child1, err := expectedTree.AddChild("https://g.org/pkg/")
		assert.Nil(t, err)
		child1.AddChild("https://g.org/")

		child2, err := child1.AddChild("https://g.org/cmd/")
		assert.Nil(t, err)
		child2.AddChild("https://g.org/x/tools")
		child2.AddChild("https://g.org/net/http")
		child2.AddChild("https://g.org/net/html")

		child3, err := child1.AddChild("https://g.org/pkg/fmt/")
		assert.Nil(t, err)
		child3.AddChild("https://g.org/")
		child3.AddChild("https://g.org/pkg/")

		child4, err := child1.AddChild("https://g.org/pkg/os/")
		assert.Nil(t, err)
		child4.AddChild("https://g.org/")
		child4.AddChild("https://g.org/pkg/")

		expectedTree.AddChild("https://g.org/cmd/")

		assert.Equal(t, expectedTree, rootNode)

		// The output of crawledURLs depth 3 onwards should be same since there aren't
		// anymore URLs below dept 3
		t.Run("depth 4", func(t *testing.T) {
			visited := make(map[string]bool)
			rootNode := tree.NewNode("https://g.org/")
			crawledURLs := Crawl("https://g.org/", 4, fetcher, rootNode, visited)
			assert.Equal(t, depth3ExpectedURLs, *crawledURLs)
			assert.Equal(t, expectedTree, rootNode)
		})
		t.Run("depth 5", func(t *testing.T) {
			visited := make(map[string]bool)
			rootNode := tree.NewNode("https://g.org/")
			crawledURLs := Crawl("https://g.org/", 5, fetcher, rootNode, visited)
			assert.Equal(t, depth3ExpectedURLs, *crawledURLs)
			assert.Equal(t, expectedTree, rootNode)
		})
	})

}

// Inspired from https://tour.golang.org/concurrency/10
// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string][]string

func (f fakeFetcher) Fetch(url string) ([]string, error) {
	if res, ok := f[url]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://g.org/": []string{
		"https://g.org/pkg/",
		"https://g.org/cmd/",
	},
	"https://g.org/pkg/": []string{
		"https://g.org/",
		"https://g.org/cmd/",
		"https://g.org/pkg/fmt/",
		"https://g.org/pkg/os/",
	},
	"https://g.org/cmd/": []string{
		"https://g.org/x/tools",
		"https://g.org/net/http",
		"https://g.org/net/html",
	},
	"https://g.org/pkg/fmt/": []string{
		"https://g.org/",
		"https://g.org/pkg/",
	},
	"https://g.org/pkg/os/": []string{
		"https://g.org/",
		"https://g.org/pkg/",
	},
}
