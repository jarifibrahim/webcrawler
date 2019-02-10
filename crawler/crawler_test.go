package crawler

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jarifibrahim/webcrawler/fetchers"

	"github.com/jarifibrahim/webcrawler/tree"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Disable logger. We don't want noisy logs when running tests
func init() {
	logrus.SetLevel(logrus.PanicLevel)
}

func TestStartCrawling(t *testing.T) {
	baseURL := "http://jarifibrahim.github.io"
	maxDepth := 3

	t.Run("with show tree", func(t *testing.T) {
		treeBuffer := bytes.Buffer{}
		sitemapBuffer := bytes.Buffer{}
		showTree := true
		StartCrawling(baseURL, maxDepth, showTree, &treeBuffer, &sitemapBuffer)
		// both the buffers should not be empty
		assert.NotEqual(t, treeBuffer, bytes.Buffer{})
		assert.NotEqual(t, sitemapBuffer, bytes.Buffer{})
	})
	t.Run("without show tree", func(t *testing.T) {
		treeBuffer := bytes.Buffer{}
		sitemapBuffer := bytes.Buffer{}
		showTree := false
		StartCrawling(baseURL, maxDepth, showTree, &treeBuffer, &sitemapBuffer)
		// tree butter should be empty
		assert.Equal(t, treeBuffer, bytes.Buffer{})
		// sitemap buffer should not be empty
		assert.NotEqual(t, sitemapBuffer, bytes.Buffer{})

	})
}
func TestWriteSiteMap(t *testing.T) {
	var writeBuffer bytes.Buffer
	cache := NewURLCache()
	cache.urls = []string{"/foo", "/bar", "/helloWorld"}
	expectedOutput := `<?xml version="1.0" encoding="UTF-8"?>

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
	cache.WriteSiteMap(&writeBuffer)
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
		cache := NewURLCache()
		wg.Add(1)
		go crawl("https://g.org/", 0, ffetcher, nil, cache)
		wg.Wait()
		assert.Equal(t, expectedURLList, cache.urls)
	})
	t.Run("with Tree", func(t *testing.T) {
		cache := NewURLCache()
		rootNode := tree.NewNode("https://g.org/")
		wg.Add(1)
		go crawl("https://g.org/", 0, ffetcher, rootNode, cache)
		wg.Wait()
		assert.Equal(t, expectedURLList, cache.urls)
		// A tree with depth 0 is only the root node
		expectedTree := tree.NewNode("https://g.org/")
		assert.Equal(t, expectedTree, rootNode)
	})
}
func TestCrawlDepth1(t *testing.T) {
	expectedURLList := []string{"https://g.org/", "https://g.org/pkg/", "https://g.org/cmd/"}
	t.Run("without tree", func(t *testing.T) {
		cache := NewURLCache()
		wg.Add(1)
		go crawl("https://g.org/", 1, ffetcher, nil, cache)
		wg.Wait()
		assert.ElementsMatch(t, expectedURLList, cache.urls)

	})
	t.Run("with tree", func(t *testing.T) {
		cache := NewURLCache()
		rootNode := tree.NewNode("https://g.org/")
		wg.Add(1)
		go crawl("https://g.org/", 1, ffetcher, rootNode, cache)
		wg.Wait()
		assert.ElementsMatch(t, expectedURLList, cache.urls)

		expectedTree := tree.NewNode("https://g.org/")
		for _, url := range expectedURLList[1:] {
			expectedTree.AddChild(url)
		}
		assert.Equal(t, expectedTree, rootNode)
	})
	t.Run("non existent URL", func(t *testing.T) {
		t.Run("without tree", func(t *testing.T) {
			cache := NewURLCache()
			wg.Add(1)
			go crawl("https://foo.org/", 1, ffetcher, nil, cache)
			wg.Wait()
			assert.Equal(t, []string{"https://foo.org/"}, cache.urls)
		})
	})
}
func TestCrawlDepth2(t *testing.T) {
	expectedURLs := []string{
		"https://g.org/", "https://g.org/cmd/", "https://g.org/net/html",
		"https://g.org/pkg/", "https://g.org/pkg/os/", "https://g.org/x/tools",
		"https://g.org/net/http", "https://g.org/pkg/fmt/",
	}
	t.Run("without tree", func(t *testing.T) {
		cache := NewURLCache()
		wg.Add(1)
		go crawl("https://g.org/", 2, ffetcher, nil, cache)
		wg.Wait()
		assert.ElementsMatch(t, expectedURLs, cache.urls)
	})
	t.Run("with tree", func(t *testing.T) {
		cache := NewURLCache()
		rootNode := tree.NewNode("https://g.org/")
		wg.Add(1)
		go crawl("https://g.org/", 2, ffetcher, rootNode, cache)
		wg.Wait()
		assert.ElementsMatch(t, expectedURLs, cache.urls)

		expectedTree := tree.NewNode("https://g.org/")
		child := expectedTree.AddChild("https://g.org/pkg/")
		child.AddChild("https://g.org/")
		child.AddChild("https://g.org/cmd/")
		child.AddChild("https://g.org/pkg/fmt/")
		child.AddChild("https://g.org/pkg/os/")
		child1 := expectedTree.AddChild("https://g.org/cmd/")
		child1.AddChild("https://g.org/x/tools")
		child1.AddChild("https://g.org/net/http")
		child1.AddChild("https://g.org/net/html")
		assert.Equal(t, expectedTree, rootNode)
	})
}
func TestCrawlDepth3(t *testing.T) {
	depth3ExpectedURLs := []string{
		"https://g.org/", "https://g.org/pkg/", "https://g.org/cmd/",
		"https://g.org/x/tools", "https://g.org/net/http", "https://g.org/net/html",
		"https://g.org/pkg/fmt/", "https://g.org/pkg/os/"}

	t.Run("without tree", func(t *testing.T) {
		cache := NewURLCache()
		wg.Add(1)
		go crawl("https://g.org/", 3, ffetcher, nil, cache)
		wg.Wait()
		assert.ElementsMatch(t, depth3ExpectedURLs, cache.urls)
	})
	t.Run("with tree", func(t *testing.T) {
		cache := NewURLCache()
		rootNode := tree.NewNode("https://g.org/")
		wg.Add(1)
		go crawl("https://g.org/", 3, ffetcher, rootNode, cache)
		wg.Wait()
		assert.ElementsMatch(t, depth3ExpectedURLs, cache.urls)

		expectedTree := tree.NewNode("https://g.org/")

		child1 := expectedTree.AddChild("https://g.org/pkg/")
		child1.AddChild("https://g.org/")
		child1.AddChild("https://g.org/cmd/")

		child2 := expectedTree.AddChild("https://g.org/cmd/")

		child2.AddChild("https://g.org/x/tools")
		child2.AddChild("https://g.org/net/http")
		child2.AddChild("https://g.org/net/html")

		child3 := child1.AddChild("https://g.org/pkg/fmt/")
		child3.AddChild("https://g.org/")
		child3.AddChild("https://g.org/pkg/")

		child4 := child1.AddChild("https://g.org/pkg/os/")
		child4.AddChild("https://g.org/")
		child4.AddChild("https://g.org/pkg/")

		assert.Equal(t, expectedTree, rootNode)

		// The output of crawledURLs depth 3 onwards should be same since there aren't
		// anymore URLs below dept 3
		t.Run("depth 4", func(t *testing.T) {
			cache := NewURLCache()
			rootNode := tree.NewNode("https://g.org/")
			wg.Add(1)
			go crawl("https://g.org/", 4, ffetcher, rootNode, cache)
			wg.Wait()
			assert.ElementsMatch(t, depth3ExpectedURLs, cache.urls)
			assert.Equal(t, expectedTree, rootNode)
		})
		t.Run("depth 5", func(t *testing.T) {
			cache := NewURLCache()
			rootNode := tree.NewNode("https://g.org/")
			wg.Add(1)
			go crawl("https://g.org/", 5, ffetcher, rootNode, cache)
			wg.Wait()
			assert.ElementsMatch(t, depth3ExpectedURLs, cache.urls)
			assert.Equal(t, expectedTree, rootNode)
		})
	})

}

func TestActualWebsite(t *testing.T) {
	expectedURLs := []string{"http://jarifibrahim.github.io",
		"https://github.com/jarifibrahim",
		"https://facebook.com/jarifibrahim2",
		"https://twitter.com/_iamibrahim",
		"mailto:jarifibrahim@gmail.com",
		"https://linkedin.com/in/jarifibrahim",
		"http://jarifibrahim.github.io/about/",
		"http://jarifibrahim.github.io/blog/golden-files-why-you-should-use-them/",
		"http://jarifibrahim.github.io/blog/",
		"http://jarifibrahim.github.io/blog/test-cleanup-with-gorm-hooks/",
		"http://jarifibrahim.github.io/blog/protractor-and-page-synchronization/",
		"https://github.com/fabric8-services/fabric8-wit/blob/master/controller/test-files/label/update/ok.label.golden.json",
		"https://github.com/fabric8-services/fabric8-wit/blob/master/controller/golden_files_test.go",
		"https://github.com/kwk",
		"https://github.com/fabric8-services/fabric8-wit/tree/master/controller/test-files",
		"http://vincent.demeester.fr/posts/2017-04-22-golang-testing-golden-file/",
		"https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859",
		"http://jarifibrahim.github.io/tags/golang/",
		"http://jarifibrahim.github.io/tags/software-testing/",
		"http://jarifibrahim.github.io/tags/golden-files/",
		"https://github.com/jinzhu/gorm",
		"http://gorm.io/docs/hooks.html",
		"http://jarifibrahim.github.io/tags/go/",
		"http://jarifibrahim.github.io/tags/gorm/",
		"https://unsplash.com/@joaosilas",
		"https://unsplash.com/",
		"https://github.com/fabric8-ui/fabric8-planner",
		"https://github.com/fabric8-ui/fabric8-ui",
		"https://openshift.io/",
		"http://nodejs.org/",
		"https://github.com/SeleniumHQ/selenium/wiki/WebDriverJs",
		"http://seleniumhq.github.io/selenium/docs/api/javascript/module/selenium-webdriver/lib/promise.html",
		"https://medium.com/@MertzAlertz/what-the-hell-is-zone-js-and-why-is-it-in-my-angular-2-6ff28bcf943e",
		"https://github.com/SeleniumHQ/selenium/issues/2969",
		"https://github.com/fabric8-ui/fabric8-planner/tree/master/src/tests",
		"https://christianliebel.com/2016/11/angular-2-protractor-timeout-heres-fix/",
		"https://github.com/angular/protractor/blob/master/docs/async-await.md",
		"https://stackoverflow.com/questions/44691940/explain-about-async-await-in-protractor",
		"http://jarifibrahim.github.io/tags/protractor-framework/",
		"http://jarifibrahim.github.io/tags/javascript/",
		"http://jarifibrahim.github.io/tags/ui-testing/",
		"http://jarifibrahim.github.io/tags/async-await/"}
	cache := NewURLCache()
	wg.Add(1)
	root := tree.NewNode("http://jarifibrahim.github.io")
	go crawl("http://jarifibrahim.github.io", 3, fetchers.NewSimpleFetcher("http://jarifibrahim.github.io"), root, cache)
	wg.Wait()
	assert.ElementsMatch(t, expectedURLs, cache.urls)
}

func BenchmarkCrawl(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cache := NewURLCache()
		wg.Add(1)
		go crawl("http://golang.org/", 4, fetchers.NewSimpleFetcher("http://golang.org/"), nil, cache)
		wg.Wait()
	}

}

// Inspired from https://tour.golang.org/concurrency/10
// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string][]string

func (f fakeFetcher) Fetch(url string, noOpExtractor fetchers.LinksExtractor) ([]string, error) {
	if res, ok := f[url]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var ffetcher = fakeFetcher{
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
