package resource

import "net/url"

type Resource struct {
	URL    url.URL
	Links  map[url.URL]bool
	Assets map[url.URL]bool
}

// Normalise returns a new Resource with all the Links and Assets
// replaced with absolute URLs. Invalid URLs, or those not HTTP and HTTPs, are removed.
// It also removes URL fragments (e.g. google.com#stuff).
func (r *Resource) Normalise() {
	newLinks := make(map[url.URL]bool, len(r.Links))
	newAssets := make(map[url.URL]bool, len(r.Assets))

	for k := range r.Links {
		absoluteURL := r.URL.ResolveReference(&k)
		if absoluteURL.Host != r.URL.Host {
			continue
		}
		absoluteURL.Fragment = ""
		newLinks[*absoluteURL] = true
	}

	for k := range r.Assets {
		absoluteURL := r.URL.ResolveReference(&k)
		if absoluteURL.Host != r.URL.Host {
			continue
		}
		absoluteURL.Fragment = ""
		newAssets[*absoluteURL] = true
	}

	r.URL.Fragment = ""
	r.Links = newLinks
	r.Assets = newAssets
}
