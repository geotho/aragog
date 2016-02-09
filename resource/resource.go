package resource
import "net/url"


type Resource struct {
	URL    url.URL
	Links  map[url.URL]bool
	Assets map[url.URL]bool
}
