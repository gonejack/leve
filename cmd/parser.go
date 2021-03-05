package cmd

import (
	"context"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"time"
)

func process(url string) (err error) {
	timeout, cancel := timeout10s()
	defer cancel()

	logrus.Debugf("fetching %s", url)
	feed, err := gofeed.NewParser().ParseURLWithContext(url, timeout)
	if err != nil {
		return
	}

	logger := logrus.WithField("feed", feed.Title)
	{
		for _, item := range feed.Items {
			logger := logger.WithField("article", item.Title)
			logger.Info("processing start")

			downloaded := make(map[string]string)
			{
				sourceList := parseSources(item.Content)
				for _, src := range sourceList {
					logger := logger.WithField("source", src)
					logger.Info("download start")
					fp, err := os.CreateTemp("", "level-")
					if err != nil {
						logrus.WithError(err).Fatal("cannot create tempfile")
						return err
					}
					logger.Debugf("open: %s", fp.Name())
					err = download(fp, src)
					if err != nil {
						logger.WithError(err).Error("download failed")
						continue
					}
					logger.Infof("download done. %s", fp.Name())
					downloaded[src] = fp.Name()
				}
			}

			err = packEmail(item, downloaded)
			if err != nil {
				logger.WithError(err).Error("pack email failed")
			} else {
				logger.Info("processing done.")
			}
		}
	}

	return
}

var srcRegexp = regexp.MustCompile(`src="([^"]+)"`)

func parseSources(html string) (list []string) {
	unique := map[string]struct{}{}

	matches := srcRegexp.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		u := match[1]
		_, exist := unique[u]
		if !exist {
			list = append(list, u)
			unique[u] = struct{}{}
		}
	}

	return
}

func timeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), duration)
}
func timeout10s() (context.Context, context.CancelFunc) {
	return timeout(time.Second * 10)
}
