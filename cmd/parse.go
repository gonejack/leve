package cmd

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseSources(html string) (list []string) {
	list, err := parseSourceByGoQuery(html)
	if err == nil {
		return
	}
	return
}

func parseSourceByGoQuery(html string) (list []string, err error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return
	}

	var srcs []string
	doc.Find("img").Each(func(i int, selection *goquery.Selection) {
		src, _ := selection.Attr("src")
		if src != "" {
			srcs = append(srcs, src)
		}
	})

	uniq := make(map[string]struct{})
	for _, src := range srcs {
		uniq[src] = struct{}{}
	}
	if len(srcs) == len(uniq) {
		list = srcs
	} else {
		for src := range uniq {
			list = append(list, src)
		}
	}

	return
}
