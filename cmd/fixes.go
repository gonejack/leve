package cmd

import (
	"github.com/mmcdole/gofeed"
	"net/url"
	"regexp"
	"strings"
)

func articleFixes(article *gofeed.Item) *gofeed.Item {
	if article.GUID == "" {
		article.GUID = article.Link
	}
	if article.Content == "" {
		article.Content = article.Description
	}

	return article
}

func srcFixes(article *gofeed.Item, src string) string {
	if !strings.HasPrefix(src, "http") {
		u, err := url.Parse(article.Link)
		if err == nil {
			u.Path = src
			src = u.String()
		}
	}

	return src
}

func cleanHTML(html string) (cleaned string) {
	cleaned = html
	cleaned = removeImageAttrs(cleaned)
	cleaned = removeIframe(cleaned)
	return
}

var srcsetRegExp = regexp.MustCompile(` srcset="[^"]*?"`)
var loadingRegExp = regexp.MustCompile(` loading="[^"]*?"`)

func removeImageAttrs(html string) (cleaned string) {
	cleaned = srcsetRegExp.ReplaceAllLiteralString(html, "")
	cleaned = loadingRegExp.ReplaceAllLiteralString(cleaned, "")
	return
}

var iframeRegExp = regexp.MustCompile(`<iframe.+?src="([^"]+)"[^>]*?>.*?(</iframe>)?`)

func removeIframe(html string) (cleaned string) {
	return iframeRegExp.ReplaceAllString(html, "<a src=$1>$1</a>")
}
