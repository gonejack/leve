package cmd

import "regexp"

var imgRegExp = regexp.MustCompile(`<img\s[^>]*?src="((http|/)[^"]+)"`)

func parseSources(html string) (list []string) {
	unique := make(map[string]struct{})
	matches := imgRegExp.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		link := match[1]
		_, exist := unique[link]
		if !exist {
			list = append(list, link)
			unique[link] = struct{}{}
		}
	}

	return
}
