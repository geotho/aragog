package main
import (
	"testing"
	"net/url"
	"github.com/geotho/aragog/resource"
)

func TestShouldCrawl(t *testing.T) {
	initial := parseURL("http://google.com/cat.php")
	Crawled[initial] = resource.Resource{}
	RootURL = parseURL("http://google.com")

	testCases := map[string]bool{
		"http://google.com/cat.php": false,
		"http://amazon.com/cat.php": false,
		"http://google.com": true,
		"http://google.com/": true,
		"http://google.com/cat.php?args": true,
		"http://google.com/cat.php#fragment": false,
		"https://google.com/cat.php": true,
		"https://google.com/cat.php#fragment": true,
	}

	for k, v := range testCases {
		if actual := shouldCrawl(parseURL(k)); actual != v {
			t.Errorf("%s: Expected %v, got %v",k, v, actual)
		}
	}


}

func parseURL(s string) url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return *u
}
