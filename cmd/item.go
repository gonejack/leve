package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/jordan-wright/email"
	"github.com/mmcdole/gofeed"
)

var filenameEscape = strings.NewReplacer(
	"/", "#slash",
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

func (o *leveItem) render(cids map[string]string) (output string, err error) {
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
func (o *leveItem) saveEmail(saves map[string]string) (filename string, err error) {
	eml := email.NewEmail()

	cids := make(map[string]string)
	for remoteRef, localFile := range saves {
		mime, err := mimetype.DetectFile(localFile)
		if err != nil {
			return "", err
		}

		file, err := os.Open(localFile)
		if err != nil {
			return "", err
		}

		contentId := o.newRandomContentId()
		attach, attachErr := eml.Attach(file, contentId, mime.String())

		_ = file.Close()

		if attachErr != nil {
			return "", attachErr
		}

		attach.HTMLRelated = true

		cids[remoteRef] = fmt.Sprintf(`cid:%s`, contentId)
	}

	html, err := o.render(cids)
	if err != nil {
		return
	}

	if *to != "" {
		eml.To = append(eml.To, *to)
	}
	eml.From = *from
	eml.Subject = o.Title
	eml.HTML = []byte(html)

	data, err := eml.Bytes()
	if err != nil {
		return
	}

	extension := "eml"
	basename := o.filename()
	filename = fmt.Sprintf("%s.%s", basename, extension)

	var file *os.File
	var inc = 1
	for {
		file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err == nil {
			break
		}
		if os.IsExist(err) {
			filename = fmt.Sprintf("%s#%d.%s", basename, inc, extension)
			inc++
			continue
		} else {
			return
		}
	}
	defer file.Close()

	_, err = file.Write(data)

	return
}

func (o *leveItem) filename() string {
	return filenameEscape.Replace(o.Title)
}
func (o *leveItem) newRandomContentId() string {
	return strings.ToUpper(uuid.New().String())
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
