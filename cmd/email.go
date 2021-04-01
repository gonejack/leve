package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/jordan-wright/email"
)

func saveEmail(item *leveItem, saves map[string]string) (filename string, err error) {
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

		contentId := randomContentID()
		attach, attachErr := eml.Attach(file, contentId, mime.String())

		_ = file.Close()

		if attachErr != nil {
			return "", attachErr
		}

		attach.HTMLRelated = true

		cids[remoteRef] = fmt.Sprintf(`cid:%s`, contentId)
	}

	html, err := item.renderHTML(cids)
	if err != nil {
		return
	}

	if *to != "" {
		eml.To = append(eml.To, *to)
	}
	eml.From = *from
	eml.Subject = item.Title
	eml.HTML = []byte(html)

	data, err := eml.Bytes()
	if err != nil {
		return
	}

	extension := "eml"
	basename := escapeFileName(item.Title)
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
func randomContentID() string {
	return strings.ToUpper(uuid.New().String())
}
