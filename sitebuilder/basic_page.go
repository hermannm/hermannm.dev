package sitebuilder

import (
	"bytes"
	"fmt"
	"html/template"
)

const BasicPageTemplateName = "basic_page.html.tmpl"

type BasicPageMarkdown struct {
	Page Page `yaml:"page"`
}

type BasicPageTemplate struct {
	Meta    TemplateMetadata
	Content template.HTML
}

func RenderBasicPage(contentPath string, metadata CommonMetadata, templates *template.Template) error {
	path := fmt.Sprintf("%s/%s", BaseContentDir, contentPath)
	body := new(bytes.Buffer)
	var frontmatter BasicPageMarkdown
	if err := readMarkdownWithFrontmatter(path, body, &frontmatter); err != nil {
		return fmt.Errorf("failed to read markdown for page: %w", err)
	}

	frontmatter.Page.TemplateName = BasicPageTemplateName

	pageTemplate := BasicPageTemplate{
		Meta: TemplateMetadata{
			Common: metadata,
			Page:   frontmatter.Page,
		},
		Content: template.HTML(body.String()),
	}
	if err := renderPage(pageTemplate.Meta, pageTemplate, templates); err != nil {
		return fmt.Errorf("failed to render page: %w", err)
	}

	return nil
}
