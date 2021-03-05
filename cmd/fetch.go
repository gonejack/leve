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
		log := logrus.WithField("source", src)

		target := filepath.Join(tempDir, fmt.Sprintf("%s%s", md5str(src), filepath.Ext(src)))
		saves[src] = target
		if fileExists(target) {
			continue
		}

		file, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			logrus.WithError(err).Fatal("cannot create tempfile")
			return nil, err
		}
		log.Debugf("open file %s", file.Name())

		func(src string, file *os.File, log *logrus.Entry) {
			group.Go(func() error {
				dlLock.Acquire(context.TODO(), 1)
				defer file.Close()
				defer dlLock.Release(1)

				err := download(file, src)
				if err != nil {
					log.WithError(err).Error("download failed")
				}
				return err
			})
		}(src, file, log)
	}

	err := group.Wait()
	if err == nil {
		logrus.Debugf("download finish")
		return saves, nil
	} else {
		return nil, err
	}
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

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("response status code %d invalid", resp.StatusCode)
	}

	var written int64
	if flagVerbose {
		bar := progressbar.NewOptions64(resp.ContentLength,
			progressbar.OptionSetTheme(progressbar.Theme{Saucer: "=", SaucerPadding: ".", BarStart: "|", BarEnd: "|"}),
			progressbar.OptionSetWidth(10),
			progressbar.OptionSpinnerType(11),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetDescription(filepath.Base(imageRef)),
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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
func md5str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
func isDirEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true
	}
	return false
}
func timeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), duration)
}
func timeout10s() (context.Context, context.CancelFunc) {
	return timeout(time.Second * 10)
}
