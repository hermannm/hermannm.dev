package sitebuilder

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	"golang.org/x/sync/errgroup"
)

const (
	BaseContentDir        = "content"
	BaseOutputDir         = "static"
	BaseTemplatesDir      = "templates"
	PageTemplatesDir      = "pages"
	ComponentTemplatesDir = "components"
)

type ContentPaths struct {
	IndexPage   string
	ProjectDirs []string
	BasicPages  []string
}

func RenderPages(contentPaths ContentPaths, metadata CommonMetadata, birthday time.Time) error {
	templates, err := parseTemplates()
	if err != nil {
		return err
	}

	var goroutines errgroup.Group
	parsedProjects := make(chan ProjectWithContentDir)
	ctx, cancelCtx := context.WithCancel(context.Background())

	goroutines.Go(func() error {
		return RenderProjectPages(
			parsedProjects, ctx, contentPaths.ProjectDirs, metadata, templates,
		)
	})

	goroutines.Go(func() error {
		return RenderIndexPage(
			parsedProjects, cancelCtx, contentPaths.IndexPage, metadata, birthday, templates,
		)
	})

	for _, basicPage := range contentPaths.BasicPages {
		basicPage := basicPage // Copy mutating loop variable to use in goroutine
		goroutines.Go(func() error {
			return RenderBasicPage(basicPage, metadata, templates)
		})
	}

	return goroutines.Wait()
}

func FormatRenderedPages() error {
	patternToFormat := fmt.Sprintf("%s/**/*.html", BaseOutputDir)
	command := exec.Command("npx", "prettier", "--write", patternToFormat)

	stderr, err := command.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get pipe to prettier's error output: %w", err)
	}

	if err := command.Start(); err != nil {
		return fmt.Errorf("failed to start prettier command: %w", err)
	}

	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		fmt.Printf("error from prettier: %s\n", errScanner.Text())
	}

	if err := command.Wait(); err != nil {
		return fmt.Errorf("failed to complete prettier command: %w", err)
	}

	return nil
}

func parseTemplates() (*template.Template, error) {
	templates := template.New(ProjectPageTemplateName).Funcs(TemplateFunctions)

	pageTemplates := fmt.Sprintf("%s/%s/*.tmpl", BaseTemplatesDir, PageTemplatesDir)
	templates, err := templates.ParseGlob(pageTemplates)
	if err != nil {
		return nil, fmt.Errorf("failed to parse page templates: %w", err)
	}

	componentTemplates := fmt.Sprintf("%s/%s/*.tmpl", BaseTemplatesDir, ComponentTemplatesDir)
	templates, err = templates.ParseGlob(componentTemplates)
	if err != nil {
		return nil, fmt.Errorf("failed to parse component templates: %w", err)
	}

	return templates, nil
}

func renderPage(meta TemplateMetadata, data any, templates *template.Template) error {
	outputPath, err := getRenderOutputPath(meta.Page.Path)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create template output file '%s': %w", outputPath, err)
	}

	if err := templates.ExecuteTemplate(outputFile, meta.Page.TemplateName, data); err != nil {
		errMessage := fmt.Sprintf("failed to execute template '%s'", meta.Page.TemplateName)
		return closeOnErr(outputFile, err, errMessage)
	}

	if err := outputFile.Close(); err != nil {
		return fmt.Errorf("failed to close file '%s': %w", outputPath, err)
	}

	return nil
}

func getRenderOutputPath(basePath string) (string, error) {
	var dir string
	var file string
	if strings.HasSuffix(basePath, ".html") {
		pathElements := strings.Split(basePath, "/")

		dirs := make([]string, 0, len(pathElements))
		for i, pathElement := range pathElements {
			if i == len(pathElements)-1 {
				file = pathElement
			} else {
				dirs = append(dirs, pathElement)
			}
		}

		dir = strings.Join(dirs, "/")
	} else {
		dir = basePath
		file = "index.html"
	}

	dir = fmt.Sprintf("%s%s", BaseOutputDir, dir)

	permissions := fs.FileMode(0755)
	if err := os.MkdirAll(dir, permissions); err != nil {
		return "", fmt.Errorf("failed to create template output directory '%s': %w", dir, err)
	}

	return fmt.Sprintf("%s/%s", dir, file), nil
}

func readMarkdownWithFrontmatter(
	markdownFilePath string, bodyDest io.Writer, frontmatterDest any,
) error {
	markdownFile, err := os.Open(markdownFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %w", markdownFilePath, err)
	}

	restOfFile, err := frontmatter.MustParse(markdownFile, frontmatterDest)
	if err != nil {
		errMessage := fmt.Sprintf("failed to parse markdown frontmatter of '%s'", markdownFilePath)
		return closeOnErr(markdownFile, err, errMessage)
	}

	if err := markdownFile.Close(); err != nil {
		return fmt.Errorf("failed to close file '%s': %w", markdownFilePath, err)
	}

	if err := newMarkdownParser().Convert(restOfFile, bodyDest); err != nil {
		return fmt.Errorf("failed to parse body of markdown file '%s': %w", markdownFilePath, err)
	}

	return nil
}

func newMarkdownParser() goldmark.Markdown {
	markdownOptions := goldmark.WithRendererOptions(
		html.WithUnsafe(),
		renderer.WithNodeRenderers(util.Prioritized(NewMarkdownLinkRenderer(), 1)),
	)

	return goldmark.New(markdownOptions)
}
