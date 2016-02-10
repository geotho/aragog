package resource

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormaliseResource(t *testing.T) {
	testCases := map[string]string{
		"http://google.com/foo.png": "http://google.com/foo.png",
		"#header": "http://google.com/testcase/",
		"/bar.jpg": "http://google.com/bar.jpg",
		"test": "http://google.com/testcase/test",
	}

	for k, v := range testCases {
		actual := &Resource{
			URL:    parseURL("http://google.com/testcase/"),
			Links:  makeURLMap(k),
			Assets: makeURLMap(k),
		}
		actual.Normalise()
		expected := makeURLMap(v)
		assert.Equal(t, expected, actual.Links, "Expected %s, got %s", expected, actual.Links)
		assert.Equal(t, expected, actual.Assets, "Expected %s, got %s", expected, actual.Assets)
	}

	amazon := "http://amazon.com/kindle.jpg"
	actual := &Resource{
		URL:    parseURL("http://google.com/testcase/"),
		Links:  makeURLMap(amazon),
		Assets: makeURLMap(amazon),
	}
	actual.Normalise()
	expected := makeURLMap()
	assert.Equal(t, expected, actual.Links, "Expected %s, got %s", expected, actual.Links)
	assert.Equal(t, expected, actual.Assets, "Expected %s, got %s", expected, actual.Assets)
}

func parseURL(parseMe string) url.URL {
	u, _ := url.Parse(parseMe)
	return *u
}

func makeURLMap(ss ...string) map[url.URL]bool {
	m := make(map[url.URL]bool, len(ss))
	for _, s := range ss {
		m[parseURL(s)] = true
	}
	return m
}
