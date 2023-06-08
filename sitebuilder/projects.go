package sitebuilder

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	"golang.org/x/sync/errgroup"
)

type ProjectProfile struct {
	Name     string `yaml:"name"`
	Slug     string `yaml:"slug"`
	IconPath string `yaml:"iconPath"`
	IconAlt  string `yaml:"iconAlt"`
}

type ProjectBase struct {
	ProjectProfile `yaml:",inline"`

	// Optional.
	TechStack []TechStackItem `yaml:"techStack,flow"`
	// Optional, defaults to TechStackDefaultTitle when TechStack is not empty.
	TechStackTitle string `yaml:"techstackTitle"`
	// Optional.
	LinkCategories []LinkCategory `yaml:"linkCategories,flow"`
	//Optional.
	Footnote template.HTML `yaml:"footnote"`
}

type TechStackItem struct {
	LinkItem `yaml:",inline"`
	// Optional.
	UsedWith []LinkItem `yaml:"usedWith,flow"`
	// Optional, but required if UsedWith is not empty.
	UsedFor string `yaml:"usedFor"`
}

type LinkCategory struct {
	Title string `yaml:"title"`
	// May omit IconPath field.
	Links []LinkItem `yaml:"links,flow"`
}

type ProjectMarkdown struct {
	ProjectBase `yaml:",inline"`

	// Nil if project should not have its own page.
	Page *Page `yaml:"page"`
}

type ProjectTemplate struct {
	ProjectBase
	Description template.HTML
}

type ProjectPageTemplate struct {
	Meta    TemplateMetadata
	Project ProjectTemplate
}

type ParsedProject struct {
	ProjectTemplate

	// Nil if project should not have its own page.
	// Path is set to the project's slug.
	Page *Page
}

type ProjectID struct {
	slug       string
	contentDir string
}

type ParsedProjects map[ProjectID]ParsedProject

func ParseProjects(contentDirNames []string) (ParsedProjects, error) {
	type ContentDir struct {
		name    string
		entries []fs.DirEntry
	}

	contentDirs := make([]ContentDir, len(contentDirNames))
	baseContentDir := os.DirFS(BaseContentDir)

	for i, dirName := range contentDirNames {
		entries, err := fs.ReadDir(baseContentDir, dirName)
		if err != nil {
			return nil, fmt.Errorf("failed to read project directory '%s': %w", dirName, err)
		}

		contentDirs[i] = ContentDir{name: dirName, entries: entries}
	}

	projects := make(map[ProjectID]ParsedProject)
	lock := &sync.Mutex{}
	var group errgroup.Group

	for _, contentDir := range contentDirs {
		contentDir := contentDir // Copy mutating loop variable to use in goroutine

		for _, dirEntry := range contentDir.entries {
			if dirEntry.IsDir() {
				continue
			}

			dirEntry := dirEntry // Copy mutating loop variable to use in goroutine

			group.Go(func() error {
				markdownFilePath := fmt.Sprintf(
					"%s/%s/%s", BaseContentDir, contentDir.name, dirEntry.Name(),
				)
				project, err := ParseProject(markdownFilePath)
				if err != nil {
					return err
				}

				id := ProjectID{slug: project.Slug, contentDir: contentDir.name}
				lock.Lock()
				projects[id] = project
				lock.Unlock()
				return nil
			})
		}
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return projects, nil
}

const (
	ProjectPageTemplateFile = "project_page.html.tmpl"
	DefaultTechStackTitle   = "Built with"
)

func ParseProject(markdownFilePath string) (ParsedProject, error) {
	descriptionBuffer := new(bytes.Buffer)
	var meta ProjectMarkdown
	if err := ReadMarkdownWithFrontmatter(markdownFilePath, descriptionBuffer, &meta); err != nil {
		return ParsedProject{}, fmt.Errorf("failed to read markdown for project: %w", err)
	}

	meta.Page.Path = fmt.Sprintf("/%s", meta.Slug)
	meta.Page.TemplateName = ProjectPageTemplateFile

	if meta.TechStackTitle == "" {
		meta.TechStackTitle = DefaultTechStackTitle
	}

	if meta.Footnote != "" {
		var builder strings.Builder
		if err := goldmark.Convert([]byte(meta.Footnote), &builder); err != nil {
			return ParsedProject{}, fmt.Errorf(
				"failed to parse project footnote as markdown: %w", err,
			)
		}
		meta.Footnote = template.HTML(builder.String())
	}

	return ParsedProject{
		ProjectTemplate: ProjectTemplate{
			ProjectBase: meta.ProjectBase,
			Description: template.HTML(descriptionBuffer.String()),
		},
		Page: meta.Page,
	}, nil
}

func RenderProjectPage(
	project ParsedProject, commonMetadata CommonMetadata, templates *template.Template,
) error {
	if project.Page == nil {
		return errors.New("attempted to render project with page field unset")
	}

	projectPage := ProjectPageTemplate{
		Meta: TemplateMetadata{
			Common: commonMetadata,
			Page:   *project.Page,
		},
		Project: project.ProjectTemplate,
	}

	if err := RenderPage(templates, projectPage.Meta, projectPage); err != nil {
		return fmt.Errorf(
			"failed to render page for project '%s': %w", project.Name, err,
		)
	}

	return nil
}
