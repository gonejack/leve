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
	for src, path := range saves {
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}

		contentId := generateContentID()
		a, e := eml.Attach(file, contentId, mime.TypeByExtension(filepath.Ext(src)))
		if e != nil {
			return "", e
		}

		a.HTMLRelated = true
		replaces = append(replaces, src, fmt.Sprintf("cid:%s", contentId))
	}
	replacer := strings.NewReplacer(replaces...)
	html := replacer.Replace(article.Content)

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
