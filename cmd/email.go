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

func saveEmail(item *gofeed.Item, downloads map[string]string) (filename string, err error) {
	e := email.NewEmail()

	var replaces []string
	for src, fpath := range downloads {
		f, err := os.Open(fpath)
		if err != nil {
			return "", err
		}

		contentId := generateContentID()
		a, e := e.Attach(f, contentId, mime.TypeByExtension(filepath.Ext(src)))
		if e != nil {
			return "", e
		}

		a.HTMLRelated = true
		replaces = append(replaces, src, fmt.Sprintf("cid:%s", contentId))
	}
	replacer := strings.NewReplacer(replaces...)

	e.Subject = item.Title
	e.HTML = []byte(replacer.Replace(item.Content))

	bytes, err := e.Bytes()
	if err != nil {
		return
	}

	filename = fmt.Sprintf("%s.eml", item.Title)
	err = ioutil.WriteFile(filename, bytes, 0666)

	return
}

func generateContentID() string {
	return strings.ToUpper(uuid.New().String())
}
