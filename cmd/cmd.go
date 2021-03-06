package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/antonfisher/nested-logrus-formatter"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	tempDir    = "temp"
	recordFile = "records.txt"
	recordSep  = "#record#"
	recordMax  = 2000

	feedsFile *string

	feedList   []string
	recordList []string
	recordMap  = make(map[string]int)

	flagVerbose = false

	cmd = &cobra.Command{
		Use:   "leve [-f feeds.txt]",
		Short: "Command line tool to save RSS articles as email.",
		Run:   run,
	}
)

func init() {
	cmd.Flags().SortFlags = false
	cmd.PersistentFlags().SortFlags = false
	feedsFile = cmd.PersistentFlags().StringP(
		"feeds",
		"f",
		"feeds.txt",
		"feeds list file",
	)
	cmd.PersistentFlags().BoolVarP(
		&flagVerbose,
		"verbose",
		"v",
		false,
		"verbose",
	)
	logrus.SetFormatter(&formatter.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		//NoColors:        true,
		HideKeys:    true,
		CallerFirst: true,
		FieldsOrder: []string{"feed", "article", "source"},
	})
}
func run(c *cobra.Command, args []string) {
	if flagVerbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// create temp dir
	err := os.MkdirAll(tempDir, 0777)
	if err != nil {
		logrus.WithError(err).Fatalf("can not create temp directory")
		return
	}

	// parse records
	file, err := os.Open(recordFile)
	if err == nil {
		sc := bufio.NewScanner(file)
		for sc.Scan() {
			row := strings.TrimSpace(sc.Text())
			pair := strings.Split(row, recordSep)
			if len(pair) == 2 {
				guid, contentLen := pair[0], pair[1]
				recordMap[guid], _ = strconv.Atoi(contentLen)
				recordList = append(recordList, row)
			}
		}
		err = sc.Err()
		_ = file.Close()
		if err != nil {
			logrus.WithError(err).Fatalf("parse %s failed", recordFile)
		}
	}

	// parse feeds
	file, err = os.Open(*feedsFile)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			feed := strings.TrimSpace(scanner.Text())
			if feed == "" || strings.HasPrefix(feed, "//") {
				continue
			}
			feedList = append(feedList, feed)
		}
		err = scanner.Err()
		_ = file.Close()
	}
	if err != nil {
		logrus.WithError(err).Fatalf("parse %s failed", feedsFile)
	}

	for _, feed := range feedList {
		log := logrus.WithField("feed", feed)

		log.Debugf("feed fetch")
		fd, err := fetchFeed(feed)
		if err != nil {
			log.WithError(err).Errorf("fetch failed")
			continue
		}

		log.Debugf("feed process")
		_, err = process(fd)
		if err != nil {
			logrus.WithError(err).Errorf("process feed %s error", feed)
		}
	}

	// write records
	file, err = os.OpenFile(recordFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err == nil {
		if len(recordList) > recordMax {
			recordList = recordList[len(recordList)-recordMax:]
		}
		for _, record := range recordList {
			_, err = fmt.Fprintln(file, record)
			if err != nil {
				break
			}
		}
		_ = file.Close()
	}
	if err != nil {
		logrus.WithError(err).Fatalf("write %s failed", recordFile)
	}

	// remove outdated temp files
	keepPoint := time.Now().Add(-time.Hour * 24 * 7)
	filepath.Walk(tempDir, func(path string, info fs.FileInfo, err error) error {
		outdated := info.ModTime().Before(keepPoint)
		if outdated {
			logrus.Debugf("removed outdated temp file %s", path)
			err := os.Remove(path)
			if err != nil {
				logrus.WithError(err).Errorf("cannot remove outdated temp file %s", path)
			}
		}
		return nil
	})
}
func process(feed *gofeed.Feed) (emails []string, err error) {
	log := logrus.WithField("feed", feed.Title)

	for _, article := range feed.Items {
		log := log.WithFields(logrus.Fields{
			"feed":    feed.Title,
			"article": article.Title,
			"guid":    article.GUID,
		})

		article = articleFixes(article)

		contentLen, exist := recordMap[article.GUID]
		if exist {
			if len(article.Content) == contentLen {
				log.Debugf("skipped")
				continue
			} else {
				log.Debugf("has update")
				article.Title += ".update"
			}
		}

		log.Debugf("fetch")
		saves, err := fetchArticle(article)
		if err != nil {
			log.WithError(err).Errorf("fetch resource failed")
			return nil, err
		}
		log.Debugf("fetched")

		log.Debugf("save")
		email, err := saveEmail(article, saves)
		if err != nil {
			log.WithError(err).Error("save email failed")
			continue
		}
		emails = append(emails, email)
		log.Infof("saved as %s", email)

		record := fmt.Sprintf("%s%s%d", article.GUID, recordSep, len(article.Content))
		recordList = append(recordList, record)
		recordMap[article.GUID] = len(article.Content)
	}

	return
}

func Execute() {
	err := cmd.Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}
