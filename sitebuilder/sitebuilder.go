package sitebuilder

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"os/exec"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
)

const (
	BaseContentDir = "content"
	BaseOutputDir  = "static"
	TemplatesDir   = "templates"
)

func CreateAndRenderTemplate(meta TemplateMetadata, data any) error {
	template, err := CreateTemplate(meta.Page.TemplateName)
	if err != nil {
		return err
	}

	return RenderTemplate(template, meta, data)
}

func CreateTemplate(templateName string) (*template.Template, error) {
	template, err := template.New(templateName).
		Funcs(TemplateFunctions).
		ParseGlob(fmt.Sprintf("%s/*.tmpl", TemplatesDir))
	if err != nil {
		return nil, fmt.Errorf("failed to create template '%s': %w", templateName, err)
	}

	return template, nil
}

func RenderTemplate(template *template.Template, meta TemplateMetadata, data any) error {
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

	if err := template.Execute(outputFile, data); err != nil {
		errMessage := fmt.Sprintf("failed to execute template '%s'", template.Name())
		return closeOnErr(outputFile, err, errMessage)
	}

	if err := outputFile.Close(); err != nil {
		return fmt.Errorf("failed to close file '%s': %w", outputPath, err)
	}

	return nil
}

func FormatRenderedTemplates() error {
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

func ReadMarkdownWithFrontmatter(
	markdownFilePath string, bodyDest io.Writer, frontmatterDest any,
) error {
	projectFile, err := os.Open(markdownFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %w", markdownFilePath, err)
	}

	restOfFile, err := frontmatter.MustParse(projectFile, frontmatterDest)
	if err != nil {
		errMessage := fmt.Sprintf("failed to parse markdown frontmatter of '%s'", markdownFilePath)
		return closeOnErr(projectFile, err, errMessage)
	}

	if err := projectFile.Close(); err != nil {
		return fmt.Errorf("failed to close file '%s': %w", markdownFilePath, err)
	}

	if err := goldmark.Convert(restOfFile, bodyDest); err != nil {
		return fmt.Errorf("failed to read body of markdown file '%s': %w", markdownFilePath, err)
	}

	return nil
}
