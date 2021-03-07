package cmd

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/mmcdole/gofeed"
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var dlLock = semaphore.NewWeighted(5)
var client = http.Client{}

func fetchFeed(url string) (*gofeed.Feed, error) {
	timeout, cancel := timeout10s()
	defer cancel()
	return gofeed.NewParser().ParseURLWithContext(url, timeout)
}
func fetchArticle(article *gofeed.Item) (map[string]string, error) {
	saves := make(map[string]string)

	logrus.Debugf("download start")

	var group errgroup.Group
	for _, src := range parseSources(article.Content) {
		target := filepath.Join(cacheDir, fmt.Sprintf("%s%s", md5str(src), filepath.Ext(src)))

		saves[src] = target

		link := srcFixes(article, src)
		func(link string, file string, log *logrus.Entry) {
			group.Go(func() error {
				dlLock.Acquire(context.TODO(), 1)
				defer dlLock.Release(1)

				err := download(file, link, log)
				if err != nil {
					log.WithError(err).Error("download failed")
				}
				return err
			})
		}(link, target, logrus.WithFields(logrus.Fields{
			"link": link,
			"file": target,
		}))
	}

	err := group.Wait()
	if err == nil {
		logrus.Debugf("download finish")
		return saves, nil
	} else {
		return nil, err
	}
}

func download(path string, src string, log *logrus.Entry) (err error) {
	timeout, cancel := timeout(time.Minute * 2)
	defer cancel()

	info, err := os.Stat(path)
	if err == nil {
		headReq, headErr := http.NewRequestWithContext(timeout, http.MethodHead, src, nil)
		if headErr != nil {
			return headErr
		}
		resp, headErr := client.Do(headReq)
		if headErr == nil && info.Size() == resp.ContentLength {
			log.Debugf("use cache")
			return // skip download
		}
	}

	req, err := http.NewRequestWithContext(timeout, http.MethodGet, src, nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("response status code %d invalid", resp.StatusCode)
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.WithError(err).Fatal("cannot open file")
		return
	}
	defer file.Close()

	var written int64
	if flagVerbose {
		bar := progressbar.NewOptions64(resp.ContentLength,
			progressbar.OptionSetTheme(progressbar.Theme{Saucer: "=", SaucerPadding: ".", BarStart: "|", BarEnd: "|"}),
			progressbar.OptionSetWidth(10),
			progressbar.OptionSpinnerType(11),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetDescription(filepath.Base(src)),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionClearOnFinish(),
		)
		defer bar.Clear()
		written, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	} else {
		written, err = io.Copy(file, resp.Body)
	}

	if err == nil && written < resp.ContentLength {
		err = fmt.Errorf("expected %s but downloaded %s", humanize.Bytes(uint64(resp.ContentLength)), humanize.Bytes(uint64(written)))
	}

	return
}

func md5str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
func timeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), duration)
}
func timeout10s() (context.Context, context.CancelFunc) {
	return timeout(time.Second * 10)
}
