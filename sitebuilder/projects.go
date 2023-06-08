package sitebuilder

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"

	"golang.org/x/sync/errgroup"
)

const DefaultTechStackTitle = "Built with"

type ProjectBase struct {
	Name string `yaml:"name"`
	Slug string `yaml:"slug"`

	IconPath string `yaml:"iconPath"`
	IconAlt  string `yaml:"iconAlt"`

	// Optional.
	TechStack []TechStackItem `yaml:"techStack,flow"`
	// Optional, defaults to TechStackDefaultTitle when TechStack is not empty.
	TechStackTitle string `yaml:"techstackTitle"`
	// Optional.
	LinkCategories []LinkCategory `yaml:"linkCategories,flow"`
	// Optional.
	Footnote template.HTML `yaml:"footnote"`
}

type ProjectTemplate struct {
	ProjectBase
	Description template.HTML
}

type ProjectMarkdown struct {
	ProjectBase `yaml:",inline"`

	// Nil if project should not have its own page.
	Page *Page `yaml:"page"`
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

type ProjectID struct {
	slug       string
	contentDir string
}

func GetProjectTemplates(contentDirNames []string) (map[ProjectID]ProjectTemplate, error) {
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

	projectTemplates := make(map[ProjectID]ProjectTemplate)
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
				projectTemplate, err := GetProjectTemplate(markdownFilePath)
				if err != nil {
					return err
				}

				id := ProjectID{slug: projectTemplate.Slug, contentDir: contentDir.name}
				projectTemplates[id] = projectTemplate
				return nil
			})
		}
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return projectTemplates, nil
}

func GetProjectTemplate(markdownFilePath string) (ProjectTemplate, error) {
	descriptionBuffer := new(bytes.Buffer)
	var meta ProjectMarkdown
	if err := ReadMarkdownWithFrontmatter(markdownFilePath, descriptionBuffer, &meta); err != nil {
		return ProjectTemplate{}, fmt.Errorf("failed to read markdown for project: %w", err)
	}

	return ProjectTemplate{
		ProjectBase: meta.ProjectBase,
		Description: template.HTML(descriptionBuffer.String()),
	}, nil
}
