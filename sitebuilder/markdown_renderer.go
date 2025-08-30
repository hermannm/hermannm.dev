package sitebuilder

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"strconv"

	"hermannm.dev/wrap"

	"github.com/yuin/goldmark/ast"
	render "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// MarkdownRenderer is a markdown renderer which:
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

//goland:noinspection GoUnusedParameter
func (renderer MarkdownRenderer) RenderLink(
	writer util.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	link, ok := node.(*ast.Link)
	if !ok {
		return ast.WalkStop, fmt.Errorf("node was not ast.Link: %v", node)
	}

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

//goland:noinspection GoUnusedParameter
func (renderer MarkdownRenderer) RenderParagraph(
	writer util.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if entering {
		if node.ChildCount() == 1 && node.FirstChild().Kind() == ast.KindImage {
			_, _ = writer.WriteString(`<figure class="flex flex-col gap-2 items-center">`)
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

func (renderer MarkdownRenderer) RenderImage(
	writer util.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	img, ok := node.(*ast.Image)
	if !ok {
		return ast.WalkStop, fmt.Errorf("node was not ast.Image: %v", node)
	}

	img.SetAttribute(
		[]byte("class"),
		[]byte("rounded-lg border-2 border-solid border-gruvbox-bg2"),
	)

	if !renderer.Unsafe && html.IsDangerousURL(img.Destination) {
		return ast.WalkContinue, nil
	}

	destination := util.EscapeHTML(util.URLEscape(img.Destination, true))

	width, height, err := getImageDimensions(BaseOutputDir + string(destination))
	if err != nil {
		return ast.WalkStop, wrap.Errorf(
			err,
			"failed to get dimensions for img '%s'",
			string(destination),
		)
	}

	_, _ = writer.WriteString(`<a href="`)
	_, _ = writer.Write(destination)
	_, _ = writer.WriteString(`">`)

	_, _ = writer.WriteString(`<img src="`)
	_, _ = writer.Write(destination)

	_, _ = writer.WriteString(`" width="`)
	_, _ = writer.WriteString(strconv.Itoa(width))
	_, _ = writer.WriteString(`" height="`)
	_, _ = writer.WriteString(strconv.Itoa(height))

	// Set empty alt attribute, since we set figcaption below
	// Rationale: https://stackoverflow.com/a/58468470
	_, _ = writer.WriteString(`" alt=""`)

	if img.Title != nil {
		_, _ = writer.WriteString(` title="`)
		renderer.Writer.Write(writer, img.Title)
		_ = writer.WriteByte('"')
	}

	if img.Attributes() != nil {
		html.RenderAttributes(writer, img, html.ImageAttributeFilter)
	}

	if renderer.XHTML {
		_, _ = writer.WriteString(" />")
	} else {
		_, _ = writer.WriteString(">")
	}
	_, _ = writer.WriteString("</a>")

	altText := nodeToHTMLText(img, source)
	if len(altText) == 0 {
		return ast.WalkStop, errors.New("missing alt text for img")
	}

	_, _ = writer.WriteString(`<figcaption class="italic text-center mb-1">`)
	_, _ = writer.Write(altText)
	_, _ = writer.WriteString("</p>")

	return ast.WalkSkipChildren, nil
}

func nodeToHTMLText(n ast.Node, source []byte) []byte {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s, ok := c.(*ast.String); ok && s.IsCode() {
			buf.Write(s.Value)
		} else if t, ok := c.(*ast.Text); ok {
			buf.Write(util.EscapeHTML(t.Value(source)))
		} else {
			buf.Write(nodeToHTMLText(c, source))
		}
	}
	return buf.Bytes()
}

func getImageDimensions(path string) (width int, height int, err error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, wrap.Error(err, "failed to open image file")
	}

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, wrap.Error(err, "failed to decode config from image")
	}

	return config.Width, config.Height, nil
}
