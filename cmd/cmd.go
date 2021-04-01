package cmd

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/antonfisher/nested-logrus-formatter"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	confDir   *string // default ~/.leve
	cacheDir  = "cache"
	seenDir   = "seen"
	feedTxt   = "feeds.txt"
	hexRegexp = regexp.MustCompile("^[0-9a-z]+$")

	from *string
	to   *string

	feeds   []string
	newSeen []string
	seenMap = make(map[string]bool)

	verbose = false

	cmd = &cobra.Command{
		Use:   "leve [-c conf-dir] [feed urls...]",
		Short: "Command line tool to save RSS articles as .eml files.",
		Run:   run,
	}
)

func defaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, ".leve")
}
func init() {
	cmd.Flags().SortFlags = false
	cmd.PersistentFlags().SortFlags = false
	confDir = cmd.PersistentFlags().StringP(
		"config-dir",
		"c",
		defaultConfigDir(),
		"config directory",
	)
	from = cmd.PersistentFlags().StringP(
		"from",
		"",
		"",
		"from address",
	)
	to = cmd.PersistentFlags().StringP(
		"to",
		"",
		"",
		"to address",
	)
	cmd.PersistentFlags().BoolVarP(
		&verbose,
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
		FieldsOrder: []string{"feed", "article", "link", "file"},
	})
}
func run(c *cobra.Command, args []string) {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infof("config dir is %s", *confDir)
	{
		err := os.MkdirAll(*confDir, 0766)
		if err != nil {
			logrus.WithError(err).Fatalf("can not create config dir")
			return
		}
		cacheDir = filepath.Join(*confDir, cacheDir)
		seenDir = filepath.Join(*confDir, seenDir)
		feedTxt = filepath.Join(*confDir, feedTxt)
	}

	// create cache dir
	err := os.MkdirAll(cacheDir, 0766)
	if err != nil {
		logrus.WithError(err).Fatalf("can not create cache dir")
		return
	}

	// parse records.txt
	err = os.MkdirAll(seenDir, 0766)
	if err != nil {
		logrus.WithError(err).Fatalf("can not create seen dir")
		return
	}

	err = filepath.Walk(seenDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if info.Size() == 0 && hexRegexp.MatchString(info.Name()) {
			duration := time.Since(info.ModTime())

			if duration > time.Hour*24*4 {
				e := os.Remove(path)
				if e != nil {
					return e
				}
			} else {
				seenMap[info.Name()] = true
			}
		}

		return nil
	})
	if err != nil {
		logrus.WithError(err).Fatalf("can not remove outdated seen items")
		return
	}

	// parse feeds
	if len(args) > 0 {
		feeds = args
	} else {
		file, err := os.OpenFile(feedTxt, os.O_RDONLY, 0766)
		if errors.Is(err, os.ErrNotExist) {
			file, err = os.Create(feedTxt)
		}
		if err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				feed := strings.TrimSpace(scanner.Text())

				switch {
				case feed == "":
					continue
				case strings.HasPrefix(feed, "//"):
					continue
				case strings.HasPrefix(feed, "#"):
					continue
				}

				feeds = append(feeds, feed)
			}
			err = scanner.Err()
			_ = file.Close()
		}
		if err != nil {
			logrus.WithError(err).Fatalf("parse %s failed", feedTxt)
			return
		}
	}

	if len(feeds) == 0 {
		logrus.Errorf("no feeds given")
		logrus.Infof("pass urls or put feed urls in %s", feedTxt)
		return
	}

	// process
	for _, feedURL := range feeds {
		log := logrus.WithField("feed", feedURL)

		log.Debugf("feed fetch")
		feed, err := fetchFeed(feedURL)
		if err != nil {
			log.WithError(err).Errorf("fetch failed")
			continue
		}

		log.Debugf("feed process")
		err = process(feed)
		if err != nil {
			logrus.WithError(err).Errorf("process feed %s error", feed.Title)
		}
	}

	// write
	for _, seen := range newSeen {
		f, err := os.Create(filepath.Join(seenDir, seen))
		if err != nil {
			logrus.WithError(err).Fatalf("write %s failed", seen)
			return
		}
		_ = f.Close()
	}

	// remove outdated temp files
	filepath.Walk(cacheDir, func(path string, info fs.FileInfo, err error) error {
		if time.Since(info.ModTime()) > time.Hour*24*7 {
			logrus.Debugf("remove outdated cache file %s", path)
			err := os.Remove(path)
			if err != nil {
				logrus.WithError(err).Errorf("cannot remove outdated cache file %s", path)
			}
		}
		return nil
	})
}
func process(feed *gofeed.Feed) (err error) {
	log := logrus.WithField("feed", feed.Title)

	for _, fi := range feed.Items {
		item := newLeveItem(fi)
		itemKey := item.key()

		log := log.WithFields(logrus.Fields{
			"feed":    feed.Title,
			"article": item.Title,
			"key":     itemKey,
		})

		if seenMap[itemKey] {
			log.Debugf("skipped")
			continue
		}

		log.Debugf("fetch")
		downloads, err := fetchResource(item)
		{
			if err != nil {
				log.WithError(err).Errorf("fetch resource failed")
				return err
			}
			log.Debugf("fetched")
		}

		log.Debugf("save")
		{
			email, err := item.saveEmail(downloads)
			if err == nil {
				log.Infof("saved as %s", email)
			} else {
				log.WithError(err).Error("save email failed")
				continue
			}
		}

		newSeen = append(newSeen, itemKey)
		seenMap[itemKey] = true
	}

	return
}
func Execute() {
	err := cmd.Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}
