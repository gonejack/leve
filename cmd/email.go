package cmd

import (
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"regexp"
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
		replaces = append(replaces, fmt.Sprintf(` src="%s"`, src), fmt.Sprintf(` src="cid:%s"`, contentId))
	}
	replacer := strings.NewReplacer(replaces...)
	html := cleanHTML(article.Content)
	html = replacer.Replace(html)
	html = html + footer(article.Link)

	eml.Subject = article.Title
	eml.HTML = []byte(html)

	data, err := eml.Bytes()
	if err != nil {
		return
	}

	extension := "eml"
	basename := escapeFileName(article.Title)
	filename = fmt.Sprintf("%s.%s", basename, extension)
	for i := 1; fileExists(filename); i++ {
		filename = fmt.Sprintf("%s#%d.%s", basename, i, extension)
	}
	err = ioutil.WriteFile(filename, data, 0666)

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

var srcsetRegExp = regexp.MustCompile(` srcset="[^"]*?"`)
var loadingRegExp = regexp.MustCompile(` loading="[^"]*?"`)

func cleanHTML(html string) (cleaned string) {
	cleaned = srcsetRegExp.ReplaceAllLiteralString(html, "")
	cleaned = loadingRegExp.ReplaceAllLiteralString(cleaned, "")
	return
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
