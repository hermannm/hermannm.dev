package sitebuilder

import (
	"bytes"
	"fmt"
	"html/template"

	"hermannm.dev/wrap"
)

const BasicPageTemplateName = "basic_page.html.tmpl"

type BasicPageMarkdown struct {
	Page `yaml:",inline"`
}

type BasicPageTemplate struct {
	Meta    TemplateMetadata
	Content template.HTML
}

func (renderer *PageRenderer) RenderBasicPage(contentPath string) (err error) {
	defer func() {
		if err != nil {
			renderer.cancel()
		}
	}()

	path := fmt.Sprintf("%s/%s", BaseContentDir, contentPath)
	body := new(bytes.Buffer)
	var metadata BasicPageMarkdown
	if err = readMarkdownWithFrontmatter(path, body, &metadata); err != nil {
		return wrap.Error(err, "failed to read markdown for page")
	}

	metadata.Page.TemplateName = BasicPageTemplateName
	metadata.Page.SetCanonicalURL(renderer.commonData.BaseURL)

	if err = validate.Struct(metadata); err != nil {
		return wrap.Errorf(err, "invalid metadata for page '%s'", contentPath)
	}

	renderer.parsedPages <- metadata.Page

	pageTemplate := BasicPageTemplate{
		Meta: TemplateMetadata{
			Common: renderer.commonData,
			Page:   metadata.Page,
		},
		Content: template.HTML(body.String()),
	}
	if err = renderer.renderPageWithAndWithoutTrailingSlash(
		pageTemplate.Meta.Page,
		pageTemplate,
	); err != nil {
		return wrap.Errorf(err, "failed to render page '%s'", pageTemplate.Meta.Page.Path)
	}

	return nil
}

// Implements WithPager to work with [PageRenderer.renderPageWithAndWithoutTrailingSlash].
func (template BasicPageTemplate) withPage(page Page) any {
	template.Meta.Page = page
	return template
}
