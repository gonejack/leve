package cmd

import "regexp"

var srcRegexp = regexp.MustCompile(`src="(http[^"]+)"`)

func parseSources(html string) (list []string) {
	unique := map[string]struct{}{}

	matches := srcRegexp.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		u := match[1]
		_, exist := unique[u]
		if !exist {
			list = append(list, u)
			unique[u] = struct{}{}
		}
	}

	return
}
