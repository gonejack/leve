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

func packEmail(item *gofeed.Item, downloads map[string]string) (err error) {
	e := email.NewEmail()

	var replaces []string
	for src, fpath := range downloads {
		f, err := os.Open(fpath)
		if err != nil {
			return err
		}

		contentId := newContentId()
		a, e := e.Attach(f, contentId, mime.TypeByExtension(filepath.Ext(src)))
		if e != nil {
			return e
		}

		a.HTMLRelated = true
		replaces = append(replaces, src, fmt.Sprintf("cid:%s", contentId))
	}
	replacer := strings.NewReplacer(replaces...)

	e.To = []string{"youi.note@qq.com"}
	e.From = "youi.note@qq.com"
	e.Sender = "youi.note"
	e.Subject = item.Title
	e.HTML = []byte(replacer.Replace(item.Content))

	bytes, err := e.Bytes()
	if err != nil {
		return
	}

	return ioutil.WriteFile(fmt.Sprintf("%s.eml", item.Title), bytes, 0666)
}

func newContentId() string {
	return strings.ToUpper(uuid.New().String())
}
