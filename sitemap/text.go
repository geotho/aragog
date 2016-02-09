package sitemap
import (
	"github.com/geotho/crawler/resource"
	"net/url"
	"sort"
	"fmt"
)

type TextSiteMap struct{}

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


func (t *TextSiteMap) SiteMap(crawled map[url.URL]resource.Resource) string {
	pages := make(Resources, 0, len(crawled))

	for _, v := range crawled {
		pages = append(pages, v)
	}

	sort.Sort(pages)

	for _, p := range pages {
		fmt.Println(p.URL.String())
		fmt.Println("\tLinks:")
		for _, s:= range URLMapToStringSlice(p.Links) {
			fmt.Println("\t\t", s)
		}
		fmt.Println("\tAssets:")
		for _, s:= range URLMapToStringSlice(p.Assets) {
			fmt.Println("\t\t", s)
		}
	}

	return ""
}

func URLMapToStringSlice(urlMap map[url.URL]bool) []string {
	s := make([]string, 0, len(urlMap))
	for k := range urlMap {
		s = append(s, k.String())
	}

	sort.Strings(s)
	return s
}
