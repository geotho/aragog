package sitemap

import (
	"bytes"
	"net/url"
	"sort"

	"github.com/geotho/crawler/resource"
)

// TextSiteMap writes text sitemaps into the /out folder.
type TextSiteMap struct{}

// Resources is an Resource slice that implements sort.Interface.
type Resources []resource.Resource

var _ sort.Interface = (*Resources)(nil)

func (s Resources) Len() int {
	return len(s)
}

func (s Resources) Less(i, j int) bool {
	return s[i].URL.String() < s[j].URL.String()
}

func (s Resources) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SiteMap writes a text sitemap into the /out folder.
func (t *TextSiteMap) SiteMap(crawled map[url.URL]resource.Resource) {
	pages := make(Resources, 0, len(crawled))

	for _, v := range crawled {
		pages = append(pages, v)
	}

	sort.Sort(pages)

	b := bytes.Buffer{}

	for _, p := range pages {
		b.WriteString(p.URL.String())
		b.WriteString("\tLinks:")
		for _, s := range URLMapToStringSlice(p.Links) {
			b.WriteString("\t\t")
			b.WriteString(s)
		}
		b.WriteString("\tAssets:")
		for _, s := range URLMapToStringSlice(p.Assets) {
			b.WriteString("\t\t")
			b.WriteString(s)
		}
	}

}

// URLMapToStringSlice converts a map of urls into a sorted string slice.
func URLMapToStringSlice(urlMap map[url.URL]bool) []string {
	s := make([]string, 0, len(urlMap))
	for k := range urlMap {
		s = append(s, k.String())
	}

	sort.Strings(s)
	return s
}
