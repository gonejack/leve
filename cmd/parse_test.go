package cmd

import (
	"testing"
)

func TestProcessHTML(t *testing.T) {
	var html = `<img src="src"></img><iframe src="iframe"></iframe><script src="script"/>`

	t.Log(fixHTML(html, map[string]string{"src": "def"}, "<p>footer</p>"))
}
