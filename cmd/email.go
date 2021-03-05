package cmd

import (
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jordan-wright/email"
	"github.com/mmcdole/gofeed"
)

func saveEmail(article *gofeed.Item, saves map[string]string) (filename string, err error) {
	eml := email.NewEmail()

	var replaces []string
	for src, localFile := range saves {
		file, err := os.Open(localFile)
		if err != nil {
			return "", err
		}

		contentId := generateContentID()
		attach, attachErr := eml.Attach(file, contentId, mime.TypeByExtension(filepath.Ext(src)))

		_ = file.Close()
		_ = os.Remove(localFile)

		if attachErr != nil {
			return "", attachErr
		}

		attach.HTMLRelated = true
		replaces = append(replaces, src, fmt.Sprintf("cid:%s", contentId))
	}
	replacer := strings.NewReplacer(replaces...)
	html := replacer.Replace(article.Content)
	html = html + footer(article.Link)

	eml.Subject = article.Title
	eml.HTML = []byte(html)

	data, err := eml.Bytes()
	if err != nil {
		return
	}

	filename = fmt.Sprintf("%s.eml", escapeFilename(article.Title))
	err = ioutil.WriteFile(filename, data, 0666)

	return
}

func generateContentID() string {
	return strings.ToUpper(uuid.New().String())
}

var replacer = strings.NewReplacer(
	"/", "#slash",
)

func escapeFilename(name string) string {
	return replacer.Replace(name)
}

var footerTPL = `<br><br><br>
<a style="display: block; display:inline-block; border-top: 1px solid #ccc; padding-top: 5px; color: #666; text-decoration: none;"
   href=""${href}"
>${href}</a>
<p style="color:#999;">
    Sent with <a style="color:#666; text-decoration:none; font-weight: bold;" href="https://github.com/gonejack/leve">LEVE</a>
</p>`

func footer(link string) string {
	return strings.ReplaceAll(footerTPL, "${href}", link)
}
