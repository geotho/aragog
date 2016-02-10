package parse

import (
	"net/url"
	"strings"
	"testing"

	"github.com/geotho/crawler/resource"
	"github.com/stretchr/testify/suite"
)

const (
	htmlNothing          = `<div>foo</div>`
	htmlA                = `<a href="foo.html"></a><a href="www.google.com/bar.html"></a>`
	htmlLink             = `<link rel="stylesheet" href="style.css"></link>`
	htmlLinkSC           = `<link rel="stylesheet" href="style.css" />`
	htmlScript           = `<script src="backboneangulargruntgulpnode.js"></script>`
	htmlImg              = `<img src="meme.jpg">`
	htmlImgSC            = `<img src="cat.gif" />`
	htmlWithStyleTag     = `<style> @import "style.css"; </style>`
	htmlWithNothingStyle = `<style></style>`
	htmlWithStyleAttr    = `<div style='background: url("cats.bmp");'></div>`

	cssNothing   = `.catvideo {background-color: #0BEEF0;}`
	cssImport    = `@import "style.css";`
	cssImportURL = `@import url("style.css");`
	cssURL       = `#cookiewarning {background: #ffffff url("img_tree.png") no-repeat right top;} #test { background: url("foo.png") no-repeat; }`
)

type ParseTestSuite struct {
	suite.Suite
}

type ParseHTMLTestCase struct {
	html          string
	expectedParse resource.Resource
}

type ParseCSSTestCase struct {
	css           string
	expectedParse map[url.URL]bool
}

func (s *ParseTestSuite) TestParseHTML() {
	tests := []ParseHTMLTestCase{
		ParseHTMLTestCase{
			html: htmlNothing,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links:  map[url.URL]bool{},
				Assets: map[url.URL]bool{},
			},
		},
		ParseHTMLTestCase{
			html: htmlA,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{
					parseURL("foo.html"):                true,
					parseURL("www.google.com/bar.html"): true,
				},
				Assets: map[url.URL]bool{},
			},
		},
		ParseHTMLTestCase{
			html: htmlLink,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{},
				Assets: map[url.URL]bool{
					parseURL("style.css"): true,
				},
			},
		},
		ParseHTMLTestCase{
			html: htmlLinkSC,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{},
				Assets: map[url.URL]bool{
					parseURL("style.css"): true,
				},
			},
		},
		ParseHTMLTestCase{
			html: htmlScript,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{},
				Assets: map[url.URL]bool{
					parseURL("backboneangulargruntgulpnode.js"): true,
				},
			},
		},
		ParseHTMLTestCase{
			html: htmlImg,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{},
				Assets: map[url.URL]bool{
					parseURL("meme.jpg"): true,
				},
			},
		},
		ParseHTMLTestCase{
			html: htmlImgSC,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{},
				Assets: map[url.URL]bool{
					parseURL("cat.gif"): true,
				},
			},
		},
		ParseHTMLTestCase{
			html: htmlNothing + htmlA + htmlLink + htmlLinkSC + htmlScript + htmlImg + htmlImgSC,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{
					parseURL("foo.html"):                true,
					parseURL("www.google.com/bar.html"): true,
				},
				Assets: map[url.URL]bool{
					parseURL("style.css"):                       true,
					parseURL("backboneangulargruntgulpnode.js"): true,
					parseURL("meme.jpg"):                        true,
					parseURL("cat.gif"):                         true,
				},
			},
		},
		ParseHTMLTestCase{
			html: htmlWithStyleTag,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{},
				Assets: map[url.URL]bool{
					parseURL("style.css"): true,
				},
			},
		},
		ParseHTMLTestCase{
			html: htmlWithNothingStyle,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links:  map[url.URL]bool{},
				Assets: map[url.URL]bool{},
			},
		},
		ParseHTMLTestCase{
			html: htmlWithNothingStyle + htmlWithStyleTag,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{},
				Assets: map[url.URL]bool{
					parseURL("style.css"): true,
				},
			},
		},
		ParseHTMLTestCase{
			html: htmlWithStyleAttr,
			expectedParse: resource.Resource{
				// URL: parseURL("http://www.google.com"),
				Links: map[url.URL]bool{},
				Assets: map[url.URL]bool{
					parseURL("cats.bmp"): true,
				},
			},
		},
	}

	for _, t := range tests {
		reader := strings.NewReader(t.html)
		resource, err := ParseHTML(reader)
		s.NoError(err)
		s.Equal(t.expectedParse, resource, "Failed for %s", t.html)
	}
}

func (s *ParseTestSuite) TestParseCSS() {
	tests := []ParseCSSTestCase{
		ParseCSSTestCase{
			css:           cssNothing,
			expectedParse: map[url.URL]bool{},
		},
		ParseCSSTestCase{
			css: cssImport,
			expectedParse: map[url.URL]bool{
				parseURL("style.css"): true,
			},
		},
		ParseCSSTestCase{
			css: cssImportURL,
			expectedParse: map[url.URL]bool{
				parseURL("style.css"): true,
			},
		},
		ParseCSSTestCase{
			css: cssURL,
			expectedParse: map[url.URL]bool{
				parseURL("img_tree.png"): true,
				parseURL("foo.png"):      true,
			},
		},
	}

	for _, t := range tests {
		resource := ParseCSS(t.css)
		s.Equal(t.expectedParse, resource, "Failed for %s", t.css)
	}
}

func parseURL(parseMe string) url.URL {
	u, _ := url.Parse(parseMe)
	return *u
}

func TestParseTestSuite(t *testing.T) {
	suite.Run(t, new(ParseTestSuite))
}
