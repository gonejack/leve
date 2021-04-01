package cmd

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseResources(html string) (list []string) {
	list, err := parseReferenceByGoQuery(html)
	if err == nil {
		return
	}
	return
}

func parseReferenceByGoQuery(html string) (list []string, err error) {
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
