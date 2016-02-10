package main

import (
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/geotho/aragog/parse"
	"github.com/geotho/aragog/resource"
	"github.com/geotho/aragog/sitemap"
)

var (
	Parses         = make(chan resource.Resource, 1000)
	MaxCrawlers    = flag.Int("crawlers", 20, "Maximum number of crawlers to use.")
	ActiveCrawlers = make(chan bool, 1000)
	Crawled        = make(map[url.URL]resource.Resource)
	Start          = flag.String("url", "", "URL to start crawling from. Usernames etc. will be ignored.")
	RootURL        url.URL
)

func main() {
	flag.Parse()
	Root := *Start
	if Root == "" {
		fmt.Println("--url flag not specified: using http://news.ycombinator.com/")
		Root = "http://news.ycombinator.com/"
	}

	var err error
	rootURL, err := url.Parse(Root)
	RootURL = *rootURL
	if err != nil || !RootURL.IsAbs() {
		fmt.Printf("Unable to parse given url %s ", RootURL)
		return
	}

	Crawl(RootURL)
	sm := sitemap.TextSiteMap{}
	sm.SiteMap(Crawled)

	(&sitemap.GraphvizSiteMap{}).SiteMap(Crawled)
	fmt.Println("DONE")
}

// Crawl ranges over the Parses channel and spawns goroutines to parse
// previously-unseen URLs. It halts once no new URLs are discovered.
func Crawl(start url.URL) {
	for i := 0; i < *MaxCrawlers; i++ {
		ActiveCrawlers <- true
	}
	<-ActiveCrawlers
	parse.Fetch(start, Parses, ActiveCrawlers)
	for r := range Parses {
		r := r
		fmt.Printf("Crawled %s\n", r.URL.String())

		Crawled[r.URL] = r
		for l := range r.Links {
			l := l
			if shouldCrawl(l) {
				<-ActiveCrawlers
				Crawled[l] = resource.Resource{}
				go parse.Fetch(l, Parses, ActiveCrawlers)
			}
		}

		for a := range r.Assets {
			a := a
			if shouldCrawl(a) && isCSS(a) {
				<-ActiveCrawlers
				Crawled[a] = resource.Resource{}
				go parse.Fetch(a, Parses, ActiveCrawlers)
			}
		}

		if len(ActiveCrawlers) == *MaxCrawlers {
			return
		}
	}
}

func isCSS(url url.URL) bool {
	return strings.HasSuffix(url.Path, ".css")
}

func shouldCrawl(url url.URL) bool {
	url.Fragment = ""
	_, alreadyCrawled := Crawled[url]
	sameHost := RootURL.Host == url.Host
	return !alreadyCrawled && sameHost
}
