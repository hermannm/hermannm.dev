package sitebuilder

import (
	"bytes"
	"errors"

	"github.com/yuin/goldmark/ast"
	render "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Custom markdown renderer which:
//   - adds class="break-words" to all links, and target="_blank" to all external links
//   - adds stand-alone images as <figure>, with alt text in a <figcaption>
type MarkdownRenderer struct {
	html.Config
}

func NewMarkdownLinkRenderer(opts ...html.Option) render.NodeRenderer {
	linkRenderer := &MarkdownRenderer{
		Config: html.NewConfig(),
	}

	for _, opt := range opts {
		opt.SetHTMLOption(&linkRenderer.Config)
	}

	return linkRenderer
}

func (renderer MarkdownRenderer) RegisterFuncs(registerer render.NodeRendererFuncRegisterer) {
	registerer.Register(ast.KindLink, renderer.RenderLink)
	registerer.Register(ast.KindParagraph, renderer.RenderParagraph)
	registerer.Register(ast.KindImage, renderer.RenderImage)
}

// Copied from goldmark HTML renderer, but now adds class="break-words" to all links, and
// target="_blank" to all external links.
//
// https://github.com/yuin/goldmark/blob/b2df67847ed38c31cf4f9e32483377a8e907a6ae/renderer/html/html.go#L552
func (renderer MarkdownRenderer) RenderLink(
	writer util.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	link := node.(*ast.Link)

	link.SetAttribute([]byte("class"), []byte("break-words"))
	if bytes.HasPrefix(link.Destination, []byte("http")) {
		link.SetAttribute([]byte("target"), []byte("_blank"))
	}

	if entering {
		_, _ = writer.WriteString(`<a href="`)
		if renderer.Unsafe || !html.IsDangerousURL(link.Destination) {
			_, _ = writer.Write(util.EscapeHTML(util.URLEscape(link.Destination, true)))
		}
		_ = writer.WriteByte('"')
		if link.Title != nil {
			_, _ = writer.WriteString(` title="`)
			renderer.Writer.Write(writer, link.Title)
			_ = writer.WriteByte('"')
		}
		if link.Attributes() != nil {
			html.RenderAttributes(writer, link, html.LinkAttributeFilter)
		}
		_ = writer.WriteByte('>')
	} else {
		_, _ = writer.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

func (renderer MarkdownRenderer) RenderParagraph(
	writer util.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if entering {
		if node.ChildCount() == 1 && node.FirstChild().Kind() == ast.KindImage {
			writer.WriteString(`<figure class="flex flex-col gap-2 items-center">`)
		} else if node.Attributes() != nil {
			_, _ = writer.WriteString("<p")
			html.RenderAttributes(writer, node, html.ParagraphAttributeFilter)
			_ = writer.WriteByte('>')
		} else {
			_, _ = writer.WriteString("<p>")
		}
	} else {
		if node.ChildCount() == 1 && node.FirstChild().Kind() == ast.KindImage {
			_, _ = writer.WriteString("</figure>\n")
		} else {
			_, _ = writer.WriteString("</p>\n")
		}
	}
	return ast.WalkContinue, nil
}

func (linkRenderer MarkdownRenderer) RenderImage(
	writer util.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	image := node.(*ast.Image)

	image.SetAttribute(
		[]byte("class"),
		[]byte("rounded-lg border-2 border-solid border-gruvbox-bg2"),
	)

	if !linkRenderer.Unsafe && html.IsDangerousURL(image.Destination) {
		return ast.WalkContinue, nil
	}

	destination := util.EscapeHTML(util.URLEscape(image.Destination, true))

	_, _ = writer.WriteString(`<a href="`)
	_, _ = writer.Write(destination)
	_, _ = writer.WriteString(`">`)

	_, _ = writer.WriteString(`<img src="`)
	_, _ = writer.Write(destination)

	// Set empty alt attribute, since we set figcaption below
	// Rationale: https://stackoverflow.com/a/58468470
	_, _ = writer.WriteString(`" alt=""`)

	if image.Title != nil {
		_, _ = writer.WriteString(` title="`)
		linkRenderer.Writer.Write(writer, image.Title)
		_ = writer.WriteByte('"')
	}

	if image.Attributes() != nil {
		html.RenderAttributes(writer, image, html.ImageAttributeFilter)
	}

	if linkRenderer.XHTML {
		_, _ = writer.WriteString(" />")
	} else {
		_, _ = writer.WriteString(">")
	}
	_, _ = writer.WriteString("</a>")

	altText := nodeToHTMLText(image, source)
	if len(altText) == 0 {
		return ast.WalkStop, errors.New("missing alt text for image")
	}
	_, _ = writer.WriteString(`<figcaption class="italic mb-1">`)
	_, _ = writer.Write(altText)
	_, _ = writer.WriteString("</p>")

	return ast.WalkSkipChildren, nil
}

func nodeToHTMLText(n ast.Node, source []byte) []byte {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s, ok := c.(*ast.String); ok && s.IsCode() {
			buf.Write(s.Text(source))
		} else if !c.HasChildren() {
			buf.Write(util.EscapeHTML(c.Text(source)))
		} else {
			buf.Write(nodeToHTMLText(c, source))
		}
	}
	return buf.Bytes()
}
