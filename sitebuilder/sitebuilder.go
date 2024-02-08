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

	PageTemplatesDir      = "templates/pages"
	ComponentTemplatesDir = "templates/components"

	IconDir = "img/icons"
)

var validate *validator.Validate = validator.New()

type ContentPaths struct {
	IndexPage   string
	ProjectDirs []string
	BasicPages  []string
}

func RenderPages(contentPaths ContentPaths, commonData CommonPageData, icons IconMap) error {
	if err := validate.Struct(commonData); err != nil {
		return wrap.Errorf(err, "invalid common page data")
	}

	projectFiles, err := readProjectContentDirs(contentPaths.ProjectDirs)
	if err != nil {
		return err
	}

	renderer, err := NewPageRenderer(
		commonData,
		icons,
		len(projectFiles),
		len(contentPaths.BasicPages),
		1,
	)
	if err != nil {
		return err
	}

	var goroutines errgroup.Group
	goroutines.Go(renderer.RenderIcons)

	for _, projectFile := range projectFiles {
		goroutines.Go(func() error {
			return renderer.RenderProjectPage(projectFile)
		})
	}

	goroutines.Go(func() error {
		return renderer.RenderIndexPage(contentPaths.IndexPage)
	})

	for _, basicPage := range contentPaths.BasicPages {
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
	commonData CommonPageData
	templates  *template.Template

	parsedProjects chan ParsedProject
	projectCount   int

	parsedPages chan Page
	pageCount   int

	icons         IconMap
	iconsRendered chan struct{}

	ctx    context.Context
	cancel func()
}

func NewPageRenderer(
	commonData CommonPageData,
	icons IconMap,
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
	pagePaths := make(chan Page, pageCount)

	ctx, cancel := context.WithCancel(context.Background())

	return PageRenderer{
		commonData:     commonData,
		templates:      templates,
		parsedProjects: parsedProjects,
		projectCount:   projectCount,
		parsedPages:    pagePaths,
		pageCount:      pageCount,
		icons:          icons,
		iconsRendered:  make(chan struct{}),
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

func FormatRenderedPages() error {
	patternToFormat := fmt.Sprintf("%s/**/*.html", BaseOutputDir)
	return ExecCommand(false, "npx", "prettier", "--write", patternToFormat)
}

func GenerateTailwindCSS(cssFileName string) error {
	outputPath := fmt.Sprintf("%s/%s", BaseOutputDir, cssFileName)
	return ExecCommand(false, "npx", "tailwindcss", "-i", cssFileName, "-o", outputPath, "--minify")
}

func ExecCommand(printOutput bool, commandName string, args ...string) error {
	var displayName string
	if commandName == "npx" && len(args) != 0 {
		displayName = args[0]
	} else {
		displayName = commandName
	}

	command := exec.Command(commandName, args...)

	stderr, err := command.StderrPipe()
	if err != nil {
		return wrap.Errorf(err, "failed to get pipe to %s's error output", displayName)
	}

	if printOutput {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	if err := command.Start(); err != nil {
		return wrap.Errorf(err, "failed to run %s", displayName)
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
		err = fmt.Errorf("%s failed: %w", displayName, err)
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

	pageTemplates := fmt.Sprintf("%s/*.tmpl", PageTemplatesDir)
	templates, err := templates.ParseGlob(pageTemplates)
	if err != nil {
		return nil, wrap.Error(err, "failed to parse page templates")
	}

	componentTemplates := fmt.Sprintf("%s/*.tmpl", ComponentTemplatesDir)
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
		case page := <-renderer.parsedPages:
			if page.Path != "/404.html" && page.RedirectURL == "" {
				pageURLs = append(
					pageURLs,
					fmt.Sprintf("%s%s", renderer.commonData.BaseURL, page.Path),
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
		renderer.WithNodeRenderers(util.Prioritized(NewMarkdownRenderer(), 1)),
	)

	return goldmark.New(markdownOptions)
}
