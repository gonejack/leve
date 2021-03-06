package cmd

import (
	"github.com/mmcdole/gofeed"
	"net/url"
	"strings"
)

func articleFixes(article *gofeed.Item) *gofeed.Item {
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
