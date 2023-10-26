package sitebuilder

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/go-playground/validator/v10"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	"golang.org/x/sync/errgroup"
	"hermannm.dev/wrap"
)

const (
	BaseContentDir = "content"
	BaseOutputDir  = "static"

	BaseTemplatesDir      = "templates"
	PageTemplatesDir      = "pages"
	ComponentTemplatesDir = "components"

	TechIconDir = "img/tech"
)

var validate *validator.Validate = validator.New()

type ContentPaths struct {
	IndexPage   string
	ProjectDirs []string
	BasicPages  []string
}

type TechIconMap map[string]struct {
	Link                  string `validate:"required,url"`
	Icon                  string `validate:"required,filepath"`
	IndexPageFallbackIcon string `validate:"omitempty,filepath"`
}

func RenderPages(
	contentPaths ContentPaths,
	metadata CommonMetadata,
	techIcons TechIconMap,
	birthday time.Time,
) error {
	if err := validate.Struct(metadata); err != nil {
		return wrap.Errorf(err, "invalid common page metadata")
	}
	for tech, icon := range techIcons {
		if err := validate.Struct(icon); err != nil {
			return wrap.Errorf(err, "invalid tech icon config for '%s'", tech)
		}
	}

	projectFiles, err := readProjectContentDirs(contentPaths.ProjectDirs)
	if err != nil {
		return err
	}

	renderer, err := NewPageRenderer(metadata, len(projectFiles), len(contentPaths.BasicPages), 1)
	if err != nil {
		return err
	}

	var goroutines errgroup.Group

	for _, projectFile := range projectFiles {
		projectFile := projectFile // Copy mutating loop variable to use in goroutine
		goroutines.Go(func() error {
			return renderer.RenderProjectPage(projectFile, techIcons)
		})
	}

	goroutines.Go(func() error {
		return renderer.RenderIndexPage(contentPaths.IndexPage, birthday)
	})

	for _, basicPage := range contentPaths.BasicPages {
		basicPage := basicPage // Copy mutating loop variable to use in goroutine
		goroutines.Go(func() error {
			return renderer.RenderBasicPage(basicPage)
		})
	}

	goroutines.Go(func() error {
		return renderer.BuildSitemap()
	})

	return goroutines.Wait()
}

type PageRenderer struct {
	metadata  CommonMetadata
	templates *template.Template

	parsedProjects chan ParsedProject
	projectCount   int

	pagePaths chan string
	pageCount int

	ctx       context.Context
	cancelCtx func()
}

func NewPageRenderer(
	metadata CommonMetadata,
	projectCount int,
	basicPageCount int,
	otherPagesCount int,
) (PageRenderer, error) {
	templates, err := parseTemplates()
	if err != nil {
		return PageRenderer{}, err
	}

	parsedProjects := make(chan ParsedProject, projectCount)

	pageCount := basicPageCount + projectCount + otherPagesCount
	pagePaths := make(chan string, pageCount)

	ctx, cancelCtx := context.WithCancel(context.Background())

	return PageRenderer{
		metadata:       metadata,
		templates:      templates,
		parsedProjects: parsedProjects,
		projectCount:   projectCount,
		pagePaths:      pagePaths,
		pageCount:      pageCount,
		ctx:            ctx,
		cancelCtx:      cancelCtx,
	}, nil
}

func FormatRenderedPages() error {
	patternToFormat := fmt.Sprintf("%s/**/*.html", BaseOutputDir)
	return execCommand("prettier", "npx", "prettier", "--write", patternToFormat)
}

func GenerateTailwindCSS(cssFileName string) error {
	outputPath := fmt.Sprintf("%s/%s", BaseOutputDir, cssFileName)
	return execCommand(
		"tailwind",
		"npx",
		"tailwindcss",
		"-i",
		cssFileName,
		"-o",
		outputPath,
		"--minify",
	)
}

func execCommand(displayName string, commandName string, args ...string) error {
	command := exec.Command(commandName, args...)

	stderr, err := command.StderrPipe()
	if err != nil {
		return wrap.Errorf(err, "failed to get pipe to %s command's error output", displayName)
	}

	if err := command.Start(); err != nil {
		return wrap.Errorf(err, "failed to start %s command", displayName)
	}

	errScanner := bufio.NewScanner(stderr)
	var commandErrs strings.Builder
	for errScanner.Scan() {
		if commandErrs.Len() != 0 {
			commandErrs.WriteRune('\n')
		}
		commandErrs.WriteString(errScanner.Text())
	}

	if err := command.Wait(); err != nil {
		err = fmt.Errorf("%s command failed: %w", displayName, err)
		if commandErrs.Len() == 0 {
			return err
		} else {
			return wrap.Error(errors.New(commandErrs.String()), err.Error())
		}
	}

	return nil
}

func parseTemplates() (*template.Template, error) {
	templates := template.New(ProjectPageTemplateName).Funcs(TemplateFunctions)

	pageTemplates := fmt.Sprintf("%s/%s/*.tmpl", BaseTemplatesDir, PageTemplatesDir)
	templates, err := templates.ParseGlob(pageTemplates)
	if err != nil {
		return nil, wrap.Error(err, "failed to parse page templates")
	}

	componentTemplates := fmt.Sprintf("%s/%s/*.tmpl", BaseTemplatesDir, ComponentTemplatesDir)
	templates, err = templates.ParseGlob(componentTemplates)
	if err != nil {
		return nil, wrap.Error(err, "failed to parse component templates")
	}

	return templates, nil
}

const sitemapFileName = "sitemap.txt"

func (renderer *PageRenderer) BuildSitemap() error {
	pageURLs := make([]string, 0, renderer.pageCount)
	for i := 0; i < renderer.pageCount; i++ {
		select {
		case pagePath := <-renderer.pagePaths:
			if pagePath != "/404.html" {
				pageURLs = append(
					pageURLs,
					fmt.Sprintf("%s%s", renderer.metadata.BaseURL, pagePath),
				)
			}
		case <-renderer.ctx.Done():
			return nil
		}
	}

	slices.Sort(pageURLs)

	sitemap := strings.Join(pageURLs, "\n")

	sitemapFile, err := os.Create(fmt.Sprintf("%s/%s", BaseOutputDir, sitemapFileName))
	if err != nil {
		return wrap.Error(err, "failed to create sitemap file")
	}
	defer sitemapFile.Close()

	if _, err := fmt.Fprintln(sitemapFile, sitemap); err != nil {
		return wrap.Error(err, "failed to write to sitemap file")
	}

	return nil
}

func (renderer *PageRenderer) renderPage(meta TemplateMetadata, data any) error {
	outputPath, err := getRenderOutputPath(meta.Page.Path)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return wrap.Errorf(err, "failed to create template output file '%s'", outputPath)
	}
	defer outputFile.Close()

	if err := renderer.templates.ExecuteTemplate(
		outputFile, meta.Page.TemplateName, data,
	); err != nil {
		return wrap.Errorf(err, "failed to execute template '%s'", meta.Page.TemplateName)
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
		return "", wrap.Errorf(err, "failed to create template output directory '%s'", dir)
	}

	return fmt.Sprintf("%s/%s", dir, file), nil
}

func readMarkdownWithFrontmatter(
	markdownFilePath string,
	bodyDest io.Writer,
	frontmatterDest any,
) error {
	markdownFile, err := os.Open(markdownFilePath)
	if err != nil {
		return wrap.Errorf(err, "failed to open file '%s'", markdownFilePath)
	}
	defer markdownFile.Close()

	restOfFile, err := frontmatter.MustParse(markdownFile, frontmatterDest)
	if err != nil {
		return wrap.Errorf(err, "failed to parse markdown frontmatter of '%s'", markdownFilePath)
	}

	if err := newMarkdownParser().Convert(restOfFile, bodyDest); err != nil {
		return wrap.Errorf(err, "failed to parse body of markdown file '%s'", markdownFilePath)
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
