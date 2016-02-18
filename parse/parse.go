package parse

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cenkalti/backoff"
	"github.com/geotho/aragog/resource"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Fetch(u url.URL, parses chan<- resource.Resource, done chan<- bool) {
	defer func() { done <- true }()

	respC := make(chan *http.Response, 1)

	// If MaxCrawlers is too high, some TCP connections die. Retry them if they fail, but not indefinitely.
	err := backoff.Retry(func() error {
		resp, err := http.Get(u.String())
		if err != nil {
			return err
		}
		respC <- resp
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		log.Printf("[Fetch] %s", err.Error())
		return
	}
	resp := <-respC
	defer resp.Body.Close()
	var parse resource.Resource

	if isCSS(u) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[Fetch] Could not read CSS body of respose to %s: %s\n", u, err.Error())
			return
		}
		assets := ParseCSS(string(body))
		parse = resource.Resource{
			URL:    u,
			Links:  make(map[url.URL]bool),
			Assets: assets,
		}
	} else {
		parse, err = ParseHTML(resp.Body)
		if err != nil {
			log.Printf("[Fetch] Failed to parse HTML: %s\n", err.Error())
		}
	}

	parse.URL = u
	(&parse).Normalise()
	parses <- parse
}

// ParseHTML takes a body of HTML and returns a Resource containing
// (possibly relative) URLs to its Links and Assets.
// <img>, <link>, <style> and <X style=...> assets are all returned.
func ParseHTML(body io.Reader) (resource.Resource, error) {
	r := resource.Resource{
		Links:  make(map[url.URL]bool),
		Assets: make(map[url.URL]bool),
	}

	z := html.NewTokenizer(body)
	for tt := z.Next(); tt != html.ErrorToken; tt = z.Next() {
		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			switch t.DataAtom {
			case atom.A:
				// <a> tags link to other pages.
				if attr, ok := extractAttrToURL(t, atom.Href); ok {
					r.Links[*attr] = true
				}
			case atom.Link:
				// <link> tags loads css assets.
				if attr, ok := extractAttrToURL(t, atom.Href); ok {
					// Ignore alternate links
					if rel := extractAttr(t, atom.Rel); rel == "stylesheet" {
						r.Assets[*attr] = true
					}
				}
			case atom.Img, atom.Script:
				// <img> tags load image assets.
				if attr, ok := extractAttrToURL(t, atom.Src); ok {
					r.Assets[*attr] = true
				}
			case atom.Style:
				// CSS between style tags can load more assets.
				tt = z.Next()
				// Avoid <style></style> by checking for text.
				if tt == html.TextToken {
					style := string(z.Text())
					for url := range ParseCSS(style) {
						r.Assets[url] = true
					}
				}
			default:
				// every element can inline CSS that load more assets e.g. <div style="background: url(...);">.
				if style := extractAttr(t, atom.Style); style != "" {
					for url := range ParseCSS(style) {
						r.Assets[url] = true
					}
				}
			}
		}
	}

	return r, nil
}

// ParseCSS takes CSS and returns a map of URIs of its
// assets from @imports and urls (for e.g. background images).
func ParseCSS(css string) map[url.URL]bool {
	// Gorilla has a css/scanner but the tests are missing @import, it uses regex and we don't need to lex CSS really.

	URLs := map[url.URL]bool{}

	// the following is a bit crufty but only O(n), no regex, and writing a custom CSS lexer is premature optimisation.
	for i := 0; i < len(css); i++ {
		substr := css[i:]
		switch {
		// only care about @import X or url(X)
		case strings.HasPrefix(substr, "@import"):
			// could be @import "style.css"; or @import url("style.css");
			semicolon := strings.IndexRune(substr, ';')
			if semicolon == -1 {
				log.Printf("[ParseCSS] invalid CSS: %s\n", substr)
				return URLs
			}
			urlString := substr[len("@import "):semicolon]
			// u is now of the form X or url(X)
			u, err := extractURL(urlString)
			if err != nil {
				log.Println(err.Error())
				return URLs
			}
			URLs[*u] = true
		case strings.HasPrefix(substr, "url("):
			closingBracket := strings.IndexRune(substr, ')')
			if closingBracket == -1 {
				log.Printf("[ParseCSS] invalid CSS: %s\n", substr)
				return URLs
			}
			urlString := substr[len("url(") : closingBracket+1]
			u, err := extractURL(urlString)
			if err != nil {
				log.Println(err.Error())
				return URLs
			}
			URLs[*u] = true
		}
	}
	return URLs
}

// extractURL extracts an unquoted, trimmed URL from strings of the form url("google.com") or "google.com"
func extractURL(extractMe string) (*url.URL, error) {
	if strings.HasPrefix(extractMe, "url(") {
		if !strings.HasSuffix(extractMe, ")") {
			return nil, fmt.Errorf("[extractURL] no matching brackets in %s", extractMe)
		}
		extractMe = extractMe[len("url(") : len(extractMe)-1]
	}
	extractMe = strings.Trim(extractMe, ` "'()`)
	return url.Parse(extractMe)
}

func isCSS(url url.URL) bool {
	return strings.HasSuffix(url.Path, ".css")
}

func extractAttr(t html.Token, attr atom.Atom) string {
	for _, a := range t.Attr {
		if a.Key == attr.String() {
			return a.Val
		}
	}
	return ""
}

// extractAttrToURL fetches the attr and turns it into a URL for the given token.
// Returns nil if attr missing.
func extractAttrToURL(t html.Token, attr atom.Atom) (*url.URL, bool) {
	for _, a := range t.Attr {
		if a.Key == attr.String() {
			u, err := url.Parse(a.Val)
			if err != nil {
				return nil, false
			}
			return u, true
		}
	}

	return nil, false
}
