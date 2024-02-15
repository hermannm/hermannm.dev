package sitebuilder

import (
	"bytes"
	"errors"
	"log/slog"
	"strconv"

	"github.com/yuin/goldmark/ast"
	render "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	"hermannm.dev/devlog/log"
)

// Custom markdown renderer which:
//   - adds class="break-words" to all links, and target="_blank" to all external links
//   - adds stand-alone images as <figure>, with alt text in a <figcaption>
//
// Rendering implementations are based on the originals from Goldmark:
// https://github.com/yuin/goldmark/blob/b2df67847ed38c31cf4f9e32483377a8e907a6ae/renderer/html/html.go
type MarkdownRenderer struct {
	html.Config
}

func NewMarkdownRenderer(opts ...html.Option) render.NodeRenderer {
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
		writer.WriteString(`<a href="`)
		if renderer.Unsafe || !html.IsDangerousURL(link.Destination) {
			writer.Write(util.EscapeHTML(util.URLEscape(link.Destination, true)))
		}
		writer.WriteByte('"')
		if link.Title != nil {
			writer.WriteString(` title="`)
			renderer.Writer.Write(writer, link.Title)
			writer.WriteByte('"')
		}
		if link.Attributes() != nil {
			html.RenderAttributes(writer, link, html.LinkAttributeFilter)
		}
		writer.WriteByte('>')
	} else {
		writer.WriteString("</a>")
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
			writer.WriteString("<p")
			html.RenderAttributes(writer, node, html.ParagraphAttributeFilter)
			writer.WriteByte('>')
		} else {
			writer.WriteString("<p>")
		}
	} else {
		if node.ChildCount() == 1 && node.FirstChild().Kind() == ast.KindImage {
			writer.WriteString("</figure>\n")
		} else {
			writer.WriteString("</p>\n")
		}
	}

	return ast.WalkContinue, nil
}

func (renderer MarkdownRenderer) RenderImage(
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

	if !renderer.Unsafe && html.IsDangerousURL(image.Destination) {
		return ast.WalkContinue, nil
	}

	destination := util.EscapeHTML(util.URLEscape(image.Destination, true))

	altText := string(nodeToHTMLText(image, source))
	if len(altText) == 0 {
		return ast.WalkStop, errors.New("missing alt text for image")
	}

	altText, width, height, dimensionsSpecified := imageDimensionsFromAltText(altText)
	if !dimensionsSpecified {
		log.Warn(
			"failed to extract image dimensions from alt text",
			slog.String("image", string(destination)),
		)
	}

	writer.WriteString(`<a href="`)
	writer.Write(destination)
	writer.WriteString(`">`)

	writer.WriteString(`<img src="`)
	writer.Write(destination)

	// Set empty alt attribute, since we set figcaption below
	// Rationale: https://stackoverflow.com/a/58468470
	writer.WriteString(`" alt=""`)

	if dimensionsSpecified {
		writer.WriteString(` width="`)
		writer.WriteString(width)
		writer.WriteString(`" height="`)
		writer.WriteString(height)
		writer.WriteByte('"')
	}

	if image.Title != nil {
		writer.WriteString(` title="`)
		renderer.Writer.Write(writer, image.Title)
		writer.WriteByte('"')
	}

	if image.Attributes() != nil {
		html.RenderAttributes(writer, image, html.ImageAttributeFilter)
	}

	if renderer.XHTML {
		writer.WriteString(" />")
	} else {
		writer.WriteString(">")
	}
	writer.WriteString("</a>")

	writer.WriteString(`<figcaption class="italic text-center mb-1">`)
	writer.WriteString(altText)
	writer.WriteString("</p>")

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

// We store image dimensions in Markdown in parentheses at the end of the alt text, like (1482x926).
// This extracts these dimensions from the alt text, to render them as proper attributes.
//
// This lets us avoid having <img> tags in our Markdown, which would complicate our custom renderer.
func imageDimensionsFromAltText(
	altText string,
) (altTextWithoutDimensions string, width string, height string, dimensionsSpecified bool) {
	if len(altText) < 2 {
		return "", "", "", false
	}

	closeParenthesis := len(altText) - 1
	if altText[closeParenthesis] != ')' {
		return "", "", "", false
	}

	openParenthesis, x := -1, -1
Loop:
	for i := closeParenthesis - 1; i >= 0; i-- {
		switch altText[i] {
		case 'x':
			x = i
		case '(':
			openParenthesis = i
			break Loop
		}
	}
	if openParenthesis == -1 || x == -1 {
		return "", "", "", false
	}

	width = altText[openParenthesis+1 : x]
	if _, err := strconv.Atoi(width); err != nil {
		return "", "", "", false
	}

	height = altText[x+1 : closeParenthesis]
	if _, err := strconv.Atoi(height); err != nil {
		return "", "", "", false
	}

	return altText[0:openParenthesis], width, height, true
}
