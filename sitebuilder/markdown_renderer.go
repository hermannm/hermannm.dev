package sitebuilder

import (
	"bytes"
	"errors"
	"hermannm.dev/wrap"
	"image"
	_ "image/png"
	"os"
	"strconv"

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

	width, height, err := getImageDimensions(BaseOutputDir + string(destination))
	if err != nil {
		return ast.WalkStop, wrap.Errorf(
			err,
			"failed to get dimensions for image '%s'",
			string(destination),
		)
	}

	writer.WriteString(`<a href="`)
	writer.Write(destination)
	writer.WriteString(`">`)

	writer.WriteString(`<img src="`)
	writer.Write(destination)

	writer.WriteString(`" width="`)
	writer.WriteString(strconv.Itoa(width))
	writer.WriteString(`" height="`)
	writer.WriteString(strconv.Itoa(height))

	// Set empty alt attribute, since we set figcaption below
	// Rationale: https://stackoverflow.com/a/58468470
	writer.WriteString(`" alt=""`)

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

	altText := nodeToHTMLText(image, source)
	if len(altText) == 0 {
		return ast.WalkStop, errors.New("missing alt text for image")
	}

	writer.WriteString(`<figcaption class="italic text-center mb-1">`)
	writer.Write(altText)
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
