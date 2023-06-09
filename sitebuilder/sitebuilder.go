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
	"time"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
	"golang.org/x/sync/errgroup"
)

const (
	BaseContentDir = "content"
	BaseOutputDir  = "static"
	TemplatesDir   = "templates"
)

func RenderPages(projectContentDirs []string, metadata CommonMetadata, birthday time.Time) error {
	templates, err := parseTemplates()
	if err != nil {
		return err
	}

	var goroutines errgroup.Group
	parsedProjects := make(chan ProjectWithContentDir)
	ctx, cancelCtx := context.WithCancel(context.Background())

	goroutines.Go(func() error {
		return RenderProjectPages(parsedProjects, ctx, projectContentDirs, metadata, templates)
	})

	goroutines.Go(func() error {
		return RenderIndexPage(parsedProjects, cancelCtx, metadata, birthday, templates)
	})

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
	templates, err := template.New(ProjectPageTemplateFile).
		Funcs(TemplateFunctions).
		ParseGlob(fmt.Sprintf("%s/*.tmpl", TemplatesDir))
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return templates, nil
}

func renderPage(templates *template.Template, meta TemplateMetadata, data any) error {
	outputDir := fmt.Sprintf("%s%s", BaseOutputDir, meta.Page.Path)
	permissions := fs.FileMode(0755)
	if err := os.MkdirAll(outputDir, permissions); err != nil {
		return fmt.Errorf("failed to create template output directory '%s': %w", outputDir, err)
	}

	outputPath := fmt.Sprintf("%s/index.html", outputDir)
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

	if err := goldmark.Convert(restOfFile, bodyDest); err != nil {
		return fmt.Errorf("failed to parse body of markdown file '%s': %w", markdownFilePath, err)
	}

	return nil
}
