package sitebuilder

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Renders all external links to open in new tabs.
type MarkdownLinkRenderer struct {
	html.Config
}

func NewMarkdownLinkRenderer(opts ...html.Option) renderer.NodeRenderer {
	linkRenderer := &MarkdownLinkRenderer{
		Config: html.NewConfig(),
	}

	for _, opt := range opts {
		opt.SetHTMLOption(&linkRenderer.Config)
	}

	return linkRenderer
}

func (linkRenderer MarkdownLinkRenderer) RegisterFuncs(
	registerer renderer.NodeRendererFuncRegisterer,
) {
	registerer.Register(ast.KindLink, linkRenderer.RenderLink)
}

// Copied from goldmark HTML renderer, but now adds target="_blank" to all external links.
//
// https://github.com/yuin/goldmark/blob/b2df67847ed38c31cf4f9e32483377a8e907a6ae/renderer/html/html.go#L552
func (linkRenderer MarkdownLinkRenderer) RenderLink(
	writer util.BufWriter,
	source []byte,
	node ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	linkNode := node.(*ast.Link)

	if bytes.HasPrefix(linkNode.Destination, []byte("http")) {
		linkNode.SetAttribute([]byte("target"), []byte("_blank"))
	}

	if entering {
		_, _ = writer.WriteString("<a href=\"")
		if linkRenderer.Unsafe || !html.IsDangerousURL(linkNode.Destination) {
			_, _ = writer.Write(util.EscapeHTML(util.URLEscape(linkNode.Destination, true)))
		}
		_ = writer.WriteByte('"')
		if linkNode.Title != nil {
			_, _ = writer.WriteString(` title="`)
			linkRenderer.Writer.Write(writer, linkNode.Title)
			_ = writer.WriteByte('"')
		}
		if linkNode.Attributes() != nil {
			html.RenderAttributes(writer, linkNode, html.LinkAttributeFilter)
		}
		_ = writer.WriteByte('>')
	} else {
		_, _ = writer.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}
