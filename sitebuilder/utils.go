package sitebuilder

import (
	"html/template"
	"strings"
)

func removeParagraphTagsAroundHTML(html string) template.HTML {
	html = strings.TrimSpace(html)
	html, _ = strings.CutPrefix(html, "<p>")
	html, _ = strings.CutSuffix(html, "</p>")
	return template.HTML(html)
}
