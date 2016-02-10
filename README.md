# aragog
A Go web crawler that produces PDF sitemaps.

![Example PDF sitemap](https://raw.githubusercontent.com/geotho/aragog/master/out/finely.co.png?token=ACFNZ6Pa0iwNLmuAFV4e5uXQRplXfVt2ks5WxKEtwA%3D%3D)

## Install

`go get github.com/geotho/aragog`

To produce the PDF graphs, you'll need Graphviz. On OS X with Homebrew, I think you can do:

`brew install graphviz`

## Usage

Build using: `go build main.go`
Run using: `./main`

Command line flags are:
- `-crawlers int`: Maximum number of crawlers to use. (default 20)
- `-url string`: URL to start crawling from. Usernames etc. will be ignored.

After crawling, a text sitemap, a .dot file and a PDF sitemap will be written into /out
