package cmd

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	feedsFile = "feeds.txt"
	stateFile = "state.json"

	parsedFeeds = make(map[string]int64)
	parsedState = make(map[string]int64)

	verbose = false

	command = &cobra.Command{
		Use:   "leve [file] [file2] [file3]...",
		Short: "Convert RSS to email",
		Run:   Run,
	}
)

func init() {
	command.PersistentFlags().BoolVarP(
		&verbose,
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

func Run(md *cobra.Command, args []string) {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// parse state
	bytes, err := ioutil.ReadFile(stateFile)
	if err == nil && len(bytes) > 0 {
		err = json.Unmarshal(bytes, &parsedState)
		if err != nil {
			logrus.WithError(err).Fatalf("parse %s failed", stateFile)
			return
		}
	}

	// parse feeds
	fp, err := os.Open(feedsFile)
	if err == nil {
		scanner := bufio.NewScanner(fp)
		for scanner.Scan() {
			feed := strings.TrimSpace(scanner.Text())
			if feed == "" {
				continue
			}

			ts, exist := parsedState[feed]
			if !exist {
				logrus.Debugf("add new feed %s", feed)
			}
			parsedFeeds[feed] = ts
		}
		err = scanner.Err()
	}
	if err != nil {
		logrus.WithError(err).Fatalf("parse %s failed", feedsFile)
	}

	for feed := range parsedState {
		_, exist := parsedFeeds[feed]
		if !exist {
			logrus.Debugf("remove feed %s", feed)
		}
	}

	for feed := range parsedFeeds {
		err := process(feed)
		if err == nil {
			parsedFeeds[feed] = time.Now().Unix()
		} else {
			logrus.WithError(err).Errorf("process feed %s error", feed)
		}
	}
}

func Execute() {
	err := command.Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}
