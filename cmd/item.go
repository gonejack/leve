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
func (o *leveItem) fixReference(ref string) string {
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
func (o *leveItem) renderHTML(cids map[string]string) (output string, err error) {
	htm := o.Item.Content

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htm))
	if err != nil {
		return
	}

	doc.Find("img").Each(func(i int, selection *goquery.Selection) {
		src, _ := selection.Attr("src")
		if src != "" && cids[src] != "" {
			selection.SetAttr("src", cids[src])
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
	doc.Find("body").AppendHtml(o.footer())

	return doc.Html()
}

func (o *leveItem) footer() string {
	const footerTPL = `<br><br>
<a style="display: block; display:inline-block; border-top: 1px solid #ccc; padding-top: 5px; color: #666; text-decoration: none;"
   href="${href}"
>${href}</a>
<p style="color:#999;">
Sent with <a style="color:#666; text-decoration:none; font-weight: bold;" href="https://github.com/gonejack/leve">LEVE</a>
</p>`

	return strings.NewReplacer(
		"${href}", o.Link,
		"${pub_time}", o.PublishedParsed.Format("2006-01-02 15:04:05"),
	).Replace(footerTPL)
}

func newLeveItem(item *gofeed.Item) *leveItem {
	it := &leveItem{Item: item}
	it.fixUUID()
	it.fixContent()

	return it
}
