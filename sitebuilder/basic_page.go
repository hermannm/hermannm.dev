package sitebuilder

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"hermannm.dev/devlog/ctxlog"
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

func (renderer *PageRenderer) RenderBasicPage(ctx context.Context, contentPath string) (err error) {
	ctx = ctxlog.WithAttrs(ctx, "page", contentPath)

	path := fmt.Sprintf("%s/%s", BaseContentDir, contentPath)
	body := new(bytes.Buffer)
	var metadata BasicPageMarkdown
	if err = readMarkdownWithFrontmatter(ctx, path, body, &metadata); err != nil {
		return wrap.ErrorContext(ctx, err, "failed to read markdown for page")
	}

	metadata.Page.TemplateName = BasicPageTemplateName
	metadata.Page.SetCanonicalURL(renderer.commonData.BaseURL)

	if err = validate.Struct(metadata); err != nil {
		return wrap.ErrorContext(ctx, err, "invalid metadata for page")
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
		ctx,
		pageTemplate.Meta.Page,
		pageTemplate,
	); err != nil {
		return wrap.ErrorContext(ctx, err, "failed to render page")
	}

	return nil
}

// Implements [withPager] to work with [PageRenderer.renderPageWithAndWithoutTrailingSlash].
func (template BasicPageTemplate) withPage(page Page) any {
	template.Meta.Page = page
	return template
}
