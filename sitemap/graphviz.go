package sitemap

import (
	"io/ioutil"
	"log"
	url "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gv "github.com/awalterschulze/gographviz"
	"github.com/geotho/crawler/resource"
)

// A GraphvizSiteMap produces a .dot file and a .pdf of the crawled site.
type GraphvizSiteMap struct {
	edges map[edge]bool
}

// GraphvizURL wraps url.URL to add some Graphviz features.
type GraphvizURL struct {
	url.URL
}

type edge struct {
	from, to string
}

// String removes special characters from g.URL.String() which cause graphviz to barf.
func (g GraphvizURL) String() string {
	r := strings.NewReplacer("-", "", "/", "", ":", "", ".", "", "?", "", "@", "", "%", "", "=", "", "&", "")
	s := g.URL.String()
	return r.Replace(s)
}

// IsPage is true iff URL has no extension, or is .html or .htm or .php.
func (g GraphvizURL) IsPage() bool {
	ext := filepath.Ext(g.Path)
	switch ext {
	case "", ".html", ".htm", ".php":
		return true
	}
	return false
}

// NodeAttrs returns an attribute map for this URL.
func (g GraphvizURL) NodeAttrs() map[string]string {
	m := make(map[string]string)
	m["style"] = "filled"
	m["fillcolor"] = g.Colour()
	m["label"] = g.BadString()
	if g.IsPage() {
		m["fontsize"] = "20"
		m["shape"] = "box"
	}
	return m
}

// BadString returns the original graphviz-breaking URL.String()
func (g GraphvizURL) BadString() string {
	return g.URL.String()
}

// Colour returns a hex colour for the type of resource this URL represents.
func (g GraphvizURL) Colour() string {
	ext := filepath.Ext(g.Path)
	switch ext {
	case "", ".html", ".htm", ".php":
		return "#DDDDDD"
	case ".gif", ".png", ".jpg", ".jpeg":
		// pink
		return "#FFC6BC"
	case ".js":
		// blue
		return "#A7D3D2"
	case ".css":
		// orange
		return "#F7A541"
	default:
		// green
		return "#A9DA88"
	}
}

// MakeNewEdge creates a new edge between from and to iff it does not already exist.
func (m *GraphvizSiteMap) MakeNewEdge(g *gv.Graph, from, to string, attrs map[string]string) {
	if m.edges == nil {
		m.edges = make(map[edge]bool)
	}
	if ok := m.edges[edge{from, to}]; !ok {
		g.AddEdge(from, to, true, attrs)
		m.edges[edge{from, to}] = true
	}
}

// SiteMap writes a .dot and a .pdf (if you have graphviz installed) to out/siteroot.dot
func (m *GraphvizSiteMap) SiteMap(crawled map[url.URL]resource.Resource) {
	// TODO: treat .html, .php "", different e.g. bigger
	g := gv.NewGraph()
	g.SetName("G")
	g.SetDir(true)
	g.SetStrict(true)
	g.AddAttr("G", "ranksep", "3")
	g.AddAttr("G", "ratio", "auto")
	for k := range crawled {
		k := GraphvizURL{k}
		g.AddNode("G", k.String(), k.NodeAttrs())
	}
	for k, v := range crawled {
		k := GraphvizURL{k}
		for link := range v.Links {
			link := GraphvizURL{link}
			m.MakeNewEdge(g, k.String(), link.String(), map[string]string{"style": "bold"})

		}
		for asset := range v.Assets {
			asset := GraphvizURL{asset}
			g.AddNode("G", asset.String(), asset.NodeAttrs())
			m.MakeNewEdge(g, k.String(), asset.String(), map[string]string{"style": "dashed"})
		}
	}

	var root string
	for k := range crawled {
		root = k.Host
		break
	}

	ioutil.WriteFile("out/"+root+".dot", []byte(g.String()), 0777)

	cmd := exec.Command("dot", "-v", "-Tpdf", "out/"+root+".dot", "-O")
	cmd.Dir, _ = os.Getwd()
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Could not make pdf: %s\n Is dot installed? \n", err.Error())
	}
}
