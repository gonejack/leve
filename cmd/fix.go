package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

func fixArticle(article *gofeed.Item) *gofeed.Item {
	if article.GUID == "" {
		article.GUID = article.Link
	}
	if article.Content == "" {
		article.Content = article.Description
	}

	return article
}
func fixURL(article *gofeed.Item, src string) string {
	if !strings.HasPrefix(src, "http") {
		u, err := url.Parse(article.Link)
		if err == nil {
			u.Path = src
			src = u.String()
		}
	}

	return src
}
func fixHTML(html string, replaces map[string]string, footer string) (output string, err error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return
	}

	doc.Find("img").Each(func(i int, selection *goquery.Selection) {
		src, _ := selection.Attr("src")
		if src != "" && replaces[src] != "" {
			selection.SetAttr("src", replaces[src])
		}
		selection.RemoveAttr("loading")
		selection.RemoveAttr("srcset")
	})
	doc.Find("iframe").Each(func(i int, selection *goquery.Selection) {
		src, _ := selection.Attr("src")
		if src != "" {
			selection.ReplaceWithHtml(fmt.Sprintf("<a src=%s>%s</a>", src, src))
		}
	})
	doc.Find("script").Each(func(i int, selection *goquery.Selection) {
		selection.Remove()
	})
	doc.Find("body").AppendHtml(footer)

	return doc.Html()
}
