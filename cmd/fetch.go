package cmd

import (
	"context"
	"github.com/mmcdole/gofeed"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

var weighted = semaphore.NewWeighted(5)
var client = http.Client{}

func fetchFeed(url string) (*gofeed.Feed, error) {
	timeout, cancel := timeout10s()
	defer cancel()
	return gofeed.NewParser().ParseURLWithContext(url, timeout)
}
func fetchResources(feed *gofeed.Feed, article *gofeed.Item) (map[string]string, error) {
	log := logrus.WithFields(logrus.Fields{
		"feed":    feed.Title,
		"article": article.Title,
	})

	var wg sync.WaitGroup
	saves := make(map[string]string)
	srcList := parseSources(article.Content)

	log.Debugf("download start")
	for _, src := range srcList {
		log := log.WithField("source", src)
		fp, err := os.CreateTemp("", "level-")
		if err != nil {
			logrus.WithError(err).Fatal("cannot create tempfile")
			return nil, err
		}
		log.Debugf("open file %s", fp.Name())

		saves[src] = fp.Name()

		wg.Add(1)
		go func(fp *os.File, log *logrus.Entry) {
			weighted.Acquire(context.TODO(), 1)
			defer fp.Close()
			defer wg.Done()
			defer weighted.Release(1)

			err = download(fp, src)
			if err != nil {
				log.WithError(err).Error("download failed")
			}
		}(fp, log)
	}
	wg.Wait()

	log.Debugf("download finish")

	return saves, nil
}

func download(file *os.File, imageRef string) (err error) {
	timeout, cancel := timeout(time.Minute * 2)
	defer cancel()

	req, err := http.NewRequestWithContext(timeout, http.MethodGet, imageRef, nil)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if flagVerbose {
		bar := progressbar.NewOptions64(resp.ContentLength,
			progressbar.OptionSetTheme(progressbar.Theme{Saucer: "=", SaucerPadding: ".", BarStart: "|", BarEnd: "|"}),
			progressbar.OptionSetWidth(10),
			progressbar.OptionSpinnerType(11),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetDescription(imageRef),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionClearOnFinish(),
		)
		defer bar.Clear()
		_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	} else {
		_, err = io.Copy(file, resp.Body)
	}

	return
}

func timeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), duration)
}
func timeout10s() (context.Context, context.CancelFunc) {
	return timeout(time.Second * 10)
}
