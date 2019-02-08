package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/net/html/atom"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

type fakeClient struct {
	responseCache map[string]string
}

func (fc fakeClient) Get(url string) (*http.Response, error) {
	if res, ok := fc.responseCache[url]; ok {
		return &http.Response{
			Body: ioutil.NopCloser(strings.NewReader(res)),
		}, nil
	}
	return nil, fmt.Errorf("not found: %s", url)
}
func TestSimpleFetcher(t *testing.T) {
	fakeClient := fakeClient{
		responseCache: map[string]string{
			"http://localhost:8000/foobar": `
			<html>
			<body>
				<a href="/hello">Hello, World!</a>
				<p>Lorem Ipsum</p>
			    <div>
				    <a href="/bye">Bye, World!</a>
				</div>
				<A HREF="/BYE">BYE, WORLD!</a>
            	<a href="#">$$$</a>
			</body>
		    </html>`,
		},
	}
	// Ensure fakeClient conforms to the Client interface
	var _ Client = fakeClient

	testFetcher := NewSimpleFetcher("http://localhost:8000/foobar")
	testFetcher.client = fakeClient
	t.Run("success", func(t *testing.T) {
		result, err := testFetcher.Fetch(testFetcher.baseURL)
		assert.Nil(t, err)
		assert.Equal(t, result, []string{"http://localhost:8000/hello", "http://localhost:8000/bye", "http://localhost:8000/BYE"})
	})
	t.Run("client error", func(t *testing.T) {
		result, err := testFetcher.Fetch("my/random/url")
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
func TestBuildURL(t *testing.T) {
	testData := []struct {
		name        string
		baseURL     string
		childURL    string
		result      string
		expectError bool
	}{
		{
			"relative URL",
			"http://foo.com",
			"/bar",
			"http://foo.com/bar",
			false,
		}, {
			"absolute URL",
			"http://foo.com",
			"http://foo.com/bar",
			"http://foo.com/bar",
			false,
		}, {
			"invalid base URL",
			"foo.....com",
			"/bar",
			"/bar",
			false,
		},
	}
	for _, tt := range testData {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			builtURL, err := buildURL(tt.baseURL, tt.childURL)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.result, builtURL)
		})
	}
}
func TestFindHref(t *testing.T) {
	Validtoken := html.Token{
		Type:     html.StartTagToken,
		DataAtom: atom.Lookup([]byte("a")),
		Data:     "",
	}
	t.Run("non-empty href", func(t *testing.T) {
		expectedHref := "http://foo.com"
		Validtoken.Attr = []html.Attribute{html.Attribute{Key: "href", Val: expectedHref}}

		assert.Equal(t, expectedHref, *findHrefValue(Validtoken))
	})
	t.Run("non-empty HREF", func(t *testing.T) {
		expectedHref := "http://foo.com"
		Validtoken.Attr = []html.Attribute{html.Attribute{Key: "HREF", Val: expectedHref}}

		assert.Equal(t, expectedHref, *findHrefValue(Validtoken))
	})
	t.Run("empty href", func(t *testing.T) {
		expectedHref := ""
		Validtoken.Attr = []html.Attribute{html.Attribute{Key: "HREF", Val: expectedHref}}

		assert.Equal(t, expectedHref, *findHrefValue(Validtoken))
	})
	t.Run("missing href", func(t *testing.T) {
		assert.Equal(t, "", *findHrefValue(Validtoken))
	})
}

func TestConvertResponseToURLList(t *testing.T) {
	baseURL := "http://site.com"
	testData := []struct {
		testName        string
		response        string
		expectedURLList []string
	}{
		{"empty string", "", nil},
		{"no <a> tags", "<html></html>", nil},
		{"one <a> tag", "<a href='/bar/foo'></a>", []string{"http://site.com/bar/foo"}},
		{"multiple <a> tags", "<a href='/foo'></a><A HREF='/bar'></A>", []string{"http://site.com/foo", "http://site.com/bar"}},
		{"<a> tag without href", "<a class='btn'></a>", nil},
		{"<a> tag with absolute URL", "<a href='http://foo.com'></a><a href='/bar'></a>", []string{"http://foo.com", "http://site.com/bar"}},
		{"<a> tag with # href", "<a href='#content'></a>", nil},
		{"duplicate <a> tags", "<a href='/foo'></a><a href='/foo'></a>", []string{"http://site.com/foo"}},
	}

	for _, tt := range testData {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			body := strings.NewReader(tt.response)
			actualURLList := ConvertResponseToURLList(baseURL, body)
			assert.Equal(t, tt.expectedURLList, actualURLList)
		})
	}
}
