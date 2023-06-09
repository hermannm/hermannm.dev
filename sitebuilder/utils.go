package sitebuilder

import (
	"fmt"
	"html/template"
	"os"
	"strings"
)

func removeParagraphTagsAroundHTML(html string) template.HTML {
	html = strings.TrimSpace(html)
	html, _ = strings.CutPrefix(html, "<p>")
	html, _ = strings.CutSuffix(html, "</p>")
	return template.HTML(html)
}

func closeOnErr(file *os.File, err error, wrappingErrMessage string) error {
	if err == nil {
		return nil
	}

	if closeErr := file.Close(); closeErr != nil {
		closeErrMessage := "failed to close file"
		return fmt.Errorf(
			"%s AND %s:\n\t%w\n\t%w", wrappingErrMessage, closeErrMessage, err, closeErr,
		)
	}

	return fmt.Errorf("%s: %w", wrappingErrMessage, err)
}
