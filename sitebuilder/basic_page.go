package sitebuilder

import (
	"bytes"
	"fmt"
	"html/template"

	"hermannm.dev/wrap"
)

const BasicPageTemplateName = "basic_page.html.tmpl"

type BasicPageMarkdown struct {
	Page Page `yaml:"page"`
}

type BasicPageTemplate struct {
	Meta    TemplateMetadata
	Content template.HTML
}

func (renderer *PageRenderer) RenderBasicPage(contentPath string) (err error) {
	defer func() {
		if err != nil {
			renderer.cancelCtx()
		}
	}()

	path := fmt.Sprintf("%s/%s", BaseContentDir, contentPath)
	body := new(bytes.Buffer)
	var frontmatter BasicPageMarkdown
	if err = readMarkdownWithFrontmatter(path, body, &frontmatter); err != nil {
		return wrap.Error(err, "failed to read markdown for page")
	}

	frontmatter.Page.Path = fmt.Sprintf("/%s", frontmatter.Page.Path)
	frontmatter.Page.TemplateName = BasicPageTemplateName

	if err = validate.Struct(frontmatter); err != nil {
		return wrap.Errorf(err, "invalid metadata for page '%s'", contentPath)
	}

	renderer.pagePaths <- frontmatter.Page.Path

	pageTemplate := BasicPageTemplate{
		Meta: TemplateMetadata{
			Common: renderer.metadata,
			Page:   frontmatter.Page,
		},
		Content: template.HTML(body.String()),
	}
	if err = renderer.renderPage(pageTemplate.Meta, pageTemplate); err != nil {
		return wrap.Error(err, "failed to render page")
	}

	return nil
}
