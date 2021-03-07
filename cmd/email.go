package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/jordan-wright/email"
	"github.com/mmcdole/gofeed"
)

func saveEmail(article *gofeed.Item, saves map[string]string) (filename string, err error) {
	eml := email.NewEmail()

	replaces := make(map[string]string)
	for src, localFile := range saves {
		mime, err := mimetype.DetectFile(localFile)
		if err != nil {
			return "", err
		}

		file, err := os.Open(localFile)
		if err != nil {
			return "", err
		}

		contentId := generateContentID()
		attach, attachErr := eml.Attach(file, contentId, mime.String())

		_ = file.Close()

		if attachErr != nil {
			return "", attachErr
		}

		attach.HTMLRelated = true

		replaces[src] = fmt.Sprintf(`cid:%s`, contentId)
	}

	html, err := fixHTML(article.Content, replaces, footer(article))
	if err != nil {
		return
	}

	eml.Subject = article.Title
	eml.HTML = []byte(html)

	data, err := eml.Bytes()
	if err != nil {
		return
	}

	extension := "eml"
	basename := escapeFileName(article.Title)
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

var filenameEscape = strings.NewReplacer(
	"/", "#slash",
)

func escapeFileName(name string) string {
	return filenameEscape.Replace(name)
}
func generateContentID() string {
	return strings.ToUpper(uuid.New().String())
}

var footerTPL = `<br><br>
<a style="display: block; display:inline-block; border-top: 1px solid #ccc; padding-top: 5px; color: #666; text-decoration: none;"
   href="${href}"
>${href}</a>
<p style="color:#999;">
Sent with <a style="color:#666; text-decoration:none; font-weight: bold;" href="https://github.com/gonejack/leve">LEVE</a>
</p>`

func footer(article *gofeed.Item) string {
	return strings.NewReplacer(
		"${href}", article.Link,
		"${pub_time}", article.PublishedParsed.Format("2006-01-02 15:04:05"),
	).Replace(footerTPL)
}
