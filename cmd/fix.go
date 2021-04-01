package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type leveItem struct {
	*gofeed.Item
}

func (o *leveItem) key() string {
	return md5str(fmt.Sprintf("%s.%d", o.GUID, len(o.Content)))
}
func (o *leveItem) fixUUID() {
	if o.Item.GUID == "" {
		o.Item.GUID = o.Item.Link
	}
}
func (o *leveItem) fixContent() {
	if o.Item.Content == "" {
		o.Item.Content = o.Item.Description
	}
}
func (o *leveItem) fixElementRef(ref string) string {
	if strings.HasPrefix(ref, "http") {
		return ref
	}

	u, err := url.Parse(o.Item.Link)
	if err != nil {
		return ref
	}
	srcu, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	if srcu.Host == "" {
		srcu.Host = u.Host
	}
	if srcu.Scheme == "" {
		srcu.Scheme = u.Scheme
	}
	return srcu.String()
}
func (o *leveItem) renderHTML(replaces map[string]string, footer string) (output string, err error) {
	htm := o.Item.Content

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htm))
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
			selection.ReplaceWithHtml(fmt.Sprintf(`<a href="%s">%s</a>`, src, src))
		}
	})
	doc.Find("script").Each(func(i int, selection *goquery.Selection) {
		selection.Remove()
	})
	doc.Find("body").AppendHtml(footer)

	return doc.Html()
}

func newLeveItem(item *gofeed.Item) *leveItem {
	it := &leveItem{Item: item}
	it.fixUUID()
	it.fixContent()

	return it
}
