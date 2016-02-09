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
	"github.com/geotho/crawler/resource"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// FetchHTML downloads HTML from given URI, extracts URIs to links and assets
// and sends this on the parses channel.
func FetchHTML(u url.URL, parses chan<- resource.Resource, done chan<- bool) {
	defer func() {
		done <- true
	}()

	respC := make(chan *http.Response, 1)

	err := backoff.Retry(func() error {
		resp, err := http.Get(u.String())
		if err != nil {
			log.Printf("RETRYING: %s", err.Error())
			return err
		}
		respC <- resp
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		log.Printf("[FetchHTML] %s", err.Error())
		return
	}

	resp := <-respC
	defer resp.Body.Close()

	parse, err := ParseHTML(resp.Body)
	if err != nil {
		log.Printf("[FetchHTML] Failed to parse HTML: %s\n", err.Error())
	}
	parse.URL = u
	parse = normaliseResource(parse)
	parses <- parse
}

// FetchCSS downloads CSS from given URI, extracts @imports and other assets
// and sends this on the parses channel.
func FetchCSS(u url.URL, parses chan<- resource.Resource, done chan<- bool) {
	defer func() {
		done <- true
	}()

	respC := make(chan *http.Response, 1)

	err := backoff.Retry(func() error {
		resp, err := http.Get(u.String())
		if err != nil {
			log.Printf("RETRYING: %s", err.Error())
			return err
		}
		respC <- resp
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		log.Printf("[FetchCSS] %s", err.Error())
		return
	}

	resp := <-respC
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[FetchCSS] Could not read body of respose to %s: %s\n", u, err.Error())
		return
	}

	parse := ParseCSS(string(body))

	r := resource.Resource{
		URL:    u,
		Links:  make(map[url.URL]bool),
		Assets: parse,
	}

	r = normaliseResource(r)
	parses <- r
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
		case html.StartTagToken:
			t := z.Token()
			switch t.DataAtom {
			case atom.A:
				// <a> tags link to other pages.
				if attr, ok := extractAttrToURL(t, atom.Href); ok {
					r.Links[*attr] = true
				}
			case atom.Link:
				// <link> tags load js and css assets.
				if attr, ok := extractAttrToURL(t, atom.Href); ok {
					// RSS is not a static asset.
					if typeAttr := extractAttr(t, atom.Type); typeAttr != "application/rss+xml" {
						r.Assets[*attr] = true
					}
				}
			case atom.Img:
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
		case html.SelfClosingTagToken:
			// <link /> and <img /> are valid in XHTML and HTML5.
			t := z.Token()
			switch t.DataAtom {
			case atom.Link:
				if attr, ok := extractAttrToURL(t, atom.Href); ok {
					r.Assets[*attr] = true
				}
			case atom.Img:
				if attr, ok := extractAttrToURL(t, atom.Src); ok {
					r.Assets[*attr] = true
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

// normaliseResource returns a new Resource with all the Links and Assets
// replaced with absolute URLs. Invalid URLs, or those not HTTP and HTTPs, are removed.
// It also removes URL fragments (e.g. google.com#stuff).
func normaliseResource(p resource.Resource) resource.Resource {
	newLinks := make(map[url.URL]bool, len(p.Links))
	newAssets := make(map[url.URL]bool, len(p.Assets))

	for k := range p.Links {
		absoluteURL := p.URL.ResolveReference(&k)
		//		if absoluteURL.Scheme != "http" {
		//			continue
		//		}
		if absoluteURL.Host != p.URL.Host {
			continue
		}
		absoluteURL.Fragment = ""
		newLinks[*absoluteURL] = true
	}

	for k := range p.Assets {
		absoluteURL := p.URL.ResolveReference(&k)
		//		if absoluteURL.Scheme != "http" {
		//			continue
		//		}
		if absoluteURL.Host != p.URL.Host {
			continue
		}
		absoluteURL.Fragment = ""
		newAssets[*absoluteURL] = true
	}

	r := resource.Resource{
		URL:    p.URL,
		Links:  newLinks,
		Assets: newAssets,
	}
	r.URL.Fragment = ""
	return r
}
