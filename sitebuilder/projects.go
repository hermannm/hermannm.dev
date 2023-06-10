package sitebuilder

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"strings"

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

	// Optional, defaults to DefaultTechStackTitle when TechStack is not empty.
	TechStackTitle string `yaml:"techstackTitle"`
	// Optional.
	LinkCategories []LinkCategory `yaml:"linkCategories,flow"`
	//Optional.
	Footnote template.HTML `yaml:"footnote"`
}

type LinkCategory struct {
	Title string `yaml:"title"`
	// May omit IconPath field.
	Links []LinkItem `yaml:"links,flow"`
}

type ProjectMarkdown struct {
	ProjectBase `yaml:",inline"`

	// Optional.
	TechStack []TechStackItemMarkdown `yaml:"techStack,flow"`

	// Optional if project page only needs Title, Path and TemplateName (these are set
	// automatically). Other fields can be set here, e.g. if project page should host a Go package.
	Page Page `yaml:"page"`
}

type ProjectTemplate struct {
	ProjectBase
	Description template.HTML
	TechStack   []TechStackItemTemplate
}

type TechStackItemMarkdown struct {
	Tech string
	// Optional, but required if UsedWith is set.
	UsedFor string `yaml:"usedFor"`
	// Optional.
	UsedWith []string `yaml:"usedWith,flow"`
}

type TechStackItemTemplate struct {
	LinkItem
	UsedFor  string
	UsedWith []LinkItem
}

type ProjectPageTemplate struct {
	Meta    TemplateMetadata
	Project ProjectTemplate
}

type ProjectWithContentDir struct {
	ProjectProfile
	ContentDir string
}

type ParsedProject struct {
	ProjectTemplate
	Page Page
}

func RenderProjectPages(
	parsedProjects chan<- ProjectWithContentDir,
	projectReceiverCtx context.Context,
	contentDirNames []string,
	metadata CommonMetadata,
	techResources TechResourceMap,
	templates *template.Template,
) error {
	contentDirs, err := readProjectContentDirs(contentDirNames)
	if err != nil {
		close(parsedProjects)
		return err
	}

	var goroutines errgroup.Group

	for _, contentDir := range contentDirs {
		contentDir := contentDir // Copy mutating loop variable to use in goroutine

		for _, dirEntry := range contentDir.entries {
			if dirEntry.IsDir() {
				continue
			}

			dirEntry := dirEntry // Copy mutating loop variable to use in goroutine

			goroutines.Go(func() error {
				markdownFilePath := fmt.Sprintf(
					"%s/%s/%s", BaseContentDir, contentDir.name, dirEntry.Name(),
				)
				project, err := parseProject(
					markdownFilePath, techResources, metadata.SiteName, metadata.GitHubIconPath,
				)
				if err != nil {
					return fmt.Errorf("failed to parse project: %w", err)
				}

				select {
				case parsedProjects <- ProjectWithContentDir{
					ProjectProfile: project.ProjectProfile,
					ContentDir:     contentDir.name,
				}: // Sends if receiver is listening
				case <-projectReceiverCtx.Done(): // If context is done, receiver is done listening
				}

				return renderProjectPage(project, metadata, templates)
			})
		}
	}

	if err := goroutines.Wait(); err != nil {
		close(parsedProjects)
		return err
	}

	return nil
}

const (
	ProjectPageTemplateName = "project_page.html.tmpl"
	DefaultTechStackTitle   = "Built with"
)

func parseProject(
	markdownFilePath string,
	techResources TechResourceMap,
	siteName string,
	githubIconPath string,
) (ParsedProject, error) {
	descriptionBuffer := new(bytes.Buffer)
	var project ProjectMarkdown
	if err := readMarkdownWithFrontmatter(markdownFilePath, descriptionBuffer, &project); err != nil {
		return ParsedProject{}, fmt.Errorf("failed to read markdown for project: %w", err)
	}

	project.Page.Title = fmt.Sprintf("%s/%s", siteName, project.Slug)
	project.Page.Path = fmt.Sprintf("/%s", project.Slug)
	project.Page.TemplateName = ProjectPageTemplateName
	if project.TechStackTitle == "" {
		project.TechStackTitle = DefaultTechStackTitle
	}
	setGitHubLinkIcons(project.LinkCategories, githubIconPath)

	techStack, err := parseTechStack(project.TechStack, techResources)
	if err != nil {
		return ParsedProject{}, fmt.Errorf(
			"failed to parse tech stack for project %s: %w", project.Name, err,
		)
	}

	if project.Footnote != "" {
		var builder strings.Builder
		if err := newMarkdownParser().Convert([]byte(project.Footnote), &builder); err != nil {
			return ParsedProject{}, fmt.Errorf(
				"failed to parse footnote for project '%s' as markdown: %w", project.Slug, err,
			)
		}
		project.Footnote = template.HTML(builder.String())
	}

	return ParsedProject{
		ProjectTemplate: ProjectTemplate{
			ProjectBase: project.ProjectBase,
			Description: template.HTML(descriptionBuffer.String()),
			TechStack:   techStack,
		},
		Page: project.Page,
	}, nil
}

func parseTechStack(
	techStack []TechStackItemMarkdown, techResources TechResourceMap,
) ([]TechStackItemTemplate, error) {
	parsed := make([]TechStackItemTemplate, len(techStack))
	for i, tech := range techStack {
		linkItem, err := techLinkItemFromResource(tech.Tech, techResources)
		if err != nil {
			return nil, err
		}

		usedWith := make([]LinkItem, len(tech.UsedWith))
		for i, tech2 := range tech.UsedWith {
			linkItem2, err := techLinkItemFromResource(tech2, techResources)
			if err != nil {
				return nil, err
			}

			usedWith[i] = linkItem2
		}

		parsed[i] = TechStackItemTemplate{
			LinkItem: linkItem,
			UsedFor:  tech.UsedFor,
			UsedWith: usedWith,
		}
	}

	return parsed, nil
}

func techLinkItemFromResource(techName string, techResources TechResourceMap) (LinkItem, error) {
	techResource, ok := techResources[techName]
	if !ok {
		return LinkItem{}, fmt.Errorf(
			"failed to find technology '%s' in tech resource map", techName,
		)
	}

	iconPath := fmt.Sprintf("/%s/%s", TechIconDir, techResource.IconFile)
	return LinkItem{Text: techName, Link: techResource.Link, IconPath: iconPath}, nil
}

const githubBaseURL = "https://github.com"

func setGitHubLinkIcons(linkCategories []LinkCategory, githubIconPath string) {
	for _, category := range linkCategories {
		for i, link := range category.Links {
			if link.IconPath == "" && strings.HasPrefix(link.Link, githubBaseURL) {
				link.IconPath = githubIconPath
				category.Links[i] = link
			}
		}
	}
}

func renderProjectPage(
	project ParsedProject, metadata CommonMetadata, templates *template.Template,
) error {
	projectPage := ProjectPageTemplate{
		Meta: TemplateMetadata{
			Common: metadata,
			Page:   project.Page,
		},
		Project: project.ProjectTemplate,
	}

	if err := renderPage(projectPage.Meta, projectPage, templates); err != nil {
		return fmt.Errorf(
			"failed to render page for project '%s': %w", project.Slug, err,
		)
	}

	return nil
}

type ContentDir struct {
	name    string
	entries []fs.DirEntry
}

func readProjectContentDirs(contentDirNames []string) ([]ContentDir, error) {
	contentDirs := make([]ContentDir, len(contentDirNames))
	baseContentDir := os.DirFS(BaseContentDir)

	for i, dirName := range contentDirNames {
		entries, err := fs.ReadDir(baseContentDir, dirName)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to read project content directory '%s': %w", dirName, err,
			)
		}

		contentDirs[i] = ContentDir{name: dirName, entries: entries}
	}

	return contentDirs, nil
}
