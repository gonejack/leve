package cmd

import (
	"bufio"
	"encoding/json"
	"github.com/antonfisher/nested-logrus-formatter"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var (
	feedsFile  = "feeds.txt"
	stateFile  = "state.json"
	serverFile = "send.json"

	currState = make(map[string]int64)
	prevState = make(map[string]int64)

	flagVerbose = false
	flagSend    = false

	argFrom *string
	argTo   *string

	send Send

	command = &cobra.Command{
		Use:   "leve",
		Short: "Convert RSS to email",
		Run:   run,
	}
)

func init() {
	argFrom = command.PersistentFlags().StringP(
		"from",
		"f",
		"",
		"email from",
	)
	argTo = command.PersistentFlags().StringP(
		"to",
		"t",
		"",
		"email t",
	)
	command.PersistentFlags().BoolVarP(
		&flagSend,
		"send",
		"s",
		false,
		"Send emails",
	)
	command.PersistentFlags().BoolVarP(
		&flagVerbose,
		"verbose",
		"v",
		false,
		"Verbose",
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

	// parse send
	bytes, err := ioutil.ReadFile(serverFile)
	if err == nil && len(bytes) > 0 {
		err = json.Unmarshal(bytes, &send)
		if err != nil {
			logrus.WithError(err).Fatalf("parse %s failed", serverFile)
			return
		}
	}

	// parse state
	bytes, err = ioutil.ReadFile(stateFile)
	if err == nil && len(bytes) > 0 {
		err = json.Unmarshal(bytes, &prevState)
		if err != nil {
			logrus.WithError(err).Fatalf("parse %s failed", stateFile)
			return
		}
	}

	// parse feeds
	file, err := os.Open(feedsFile)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			feed := strings.TrimSpace(scanner.Text())
			if feed == "" {
				continue
			}

			checkTime, exist := prevState[feed]
			if !exist {
				logrus.Debugf("add new feed %s", feed)
			}
			currState[feed] = checkTime
		}
		err = scanner.Err()
	}
	if err != nil {
		logrus.WithError(err).Fatalf("parse %s failed", feedsFile)
	}

	for feed := range prevState {
		_, exist := currState[feed]
		if !exist {
			logrus.Info("removed %s", feed)
		}
	}

	for feed := range currState {
		log := logrus.WithField("feed", feed)

		log.Debugf("feed fetch")
		fd, err := fetchFeed(feed)
		if err != nil {
			log.WithError(err).Errorf("fetch failed")
			continue
		}

		log.Debugf("feed process")
		emails, err := process(fd)
		if err != nil {
			logrus.WithError(err).Errorf("process feed %s error", feed)
			continue
		}

		currState[feed] = time.Now().Unix()

		if flagSend {
			log.Debugf("send")
			sendAndRemove(emails)
		}
	}

	bytes, _ = json.Marshal(currState)
	ioutil.WriteFile(stateFile, bytes, 0666)
}
func process(feed *gofeed.Feed) (emails []string, err error) {
	log := logrus.WithField("feed", feed.Title)

	for _, article := range feed.Items {
		log := log.WithField("article", article.Title)

		log.Debugf("article fetch")
		saves, err := fetchResources(feed, article)
		if err != nil {
			log.WithError(err).Errorf("fetch resource failed")
			return nil, err
		}
		log.Debugf("article fetched")

		log.Debugf("email save")
		email, err := saveEmail(article, saves)
		if err != nil {
			log.WithError(err).Error("save email failed")
			continue
		}
		emails = append(emails, email)
		log.Infof("email saved %s", email)
	}

	return
}

func Execute() {
	err := command.Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}
