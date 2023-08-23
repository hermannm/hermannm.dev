package sitebuilder

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"strings"
)

type ProjectProfile struct {
	Name string `yaml:"name" validate:"required"`
	Slug string `yaml:"slug" validate:"required"`
	// Optional if not included in index page.
	IconPath string `yaml:"iconPath" validate:"omitempty,filepath"`
	IconAlt  string `yaml:"iconAlt"`
}

type ProjectBase struct {
	ProjectProfile `yaml:",inline"`
	// Optional, defaults to DefaultTechStackTitle when TechStack is not empty.
	TechStackTitle string        `yaml:"techStackTitle"`
	LinkGroups     []LinkGroup   `yaml:"linkGroups,flow"` // Optional.
	Footnote       template.HTML `yaml:"footnote"`        // Optional.
}

type LinkGroup struct {
	Title string `yaml:"title"`
	// May omit IconPath field.
	Links []LinkItem `yaml:"links,flow"`
}

type ProjectMarkdown struct {
	ProjectBase `yaml:",inline"`
	TechStack   []TechStackItemMarkdown `yaml:"techStack,flow"` // Optional.
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
	Tech     string   `yaml:"tech" validate:"required"`
	UsedFor  string   `yaml:"usedFor"`       // Optional.
	UsedWith []string `yaml:"usedWith,flow"` // Optional.
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

type ProjectContentFile struct {
	name      string
	directory string
}

func (renderer *PageRenderer) RenderProjectPage(
	projectFile ProjectContentFile, techResources TechResourceMap,
) (err error) {
	defer func() {
		if err != nil {
			renderer.cancelChannels()
		}
	}()

	markdownFilePath := fmt.Sprintf(
		"%s/%s/%s", BaseContentDir, projectFile.directory, projectFile.name,
	)
	var project ParsedProject
	if project, err = parseProject(markdownFilePath, techResources, renderer.metadata); err != nil {
		return fmt.Errorf("failed to parse project '%s': %w", projectFile.name, err)
	}

	renderer.pagePaths <- project.Page.Path

	renderer.parsedProjects <- ProjectWithContentDir{
		ProjectProfile: project.ProjectProfile,
		ContentDir:     projectFile.directory,
	}

	projectPage := ProjectPageTemplate{
		Meta: TemplateMetadata{
			Common: renderer.metadata,
			Page:   project.Page,
		},
		Project: project.ProjectTemplate,
	}

	if err = renderer.renderPage(projectPage.Meta, projectPage); err != nil {
		return fmt.Errorf("failed to render page for project '%s': %w", project.Slug, err)
	}

	return nil
}

func readProjectContentDirs(contentDirNames []string) ([]ProjectContentFile, error) {
	var files []ProjectContentFile
	baseContentDir := os.DirFS(BaseContentDir)

	for _, dirName := range contentDirNames {
		entries, err := fs.ReadDir(baseContentDir, dirName)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to read project content directory '%s': %w", dirName, err,
			)
		}

		for _, dirEntry := range entries {
			if !dirEntry.IsDir() {
				file := ProjectContentFile{name: dirEntry.Name(), directory: dirName}
				files = append(files, file)
			}
		}
	}

	return files, nil
}

const (
	ProjectPageTemplateName = "project_page.html.tmpl"
	DefaultTechStackTitle   = "Built with"
)

func parseProject(
	markdownFilePath string, techResources TechResourceMap, metadata CommonMetadata,
) (ParsedProject, error) {
	descriptionBuffer := new(bytes.Buffer)
	var project ProjectMarkdown
	if err := readMarkdownWithFrontmatter(markdownFilePath, descriptionBuffer, &project); err != nil {
		return ParsedProject{}, fmt.Errorf("failed to read markdown for project: %w", err)
	}

	project.Page.Title = fmt.Sprintf("%s/%s", metadata.SiteName, project.Slug)
	project.Page.Path = fmt.Sprintf("/%s", project.Slug)
	project.Page.TemplateName = ProjectPageTemplateName
	if project.TechStackTitle == "" {
		project.TechStackTitle = DefaultTechStackTitle
	}
	setGitHubLinkIcons(project.LinkGroups, metadata.GitHubIconPath)

	if err := validate.Struct(project); err != nil {
		return ParsedProject{}, fmt.Errorf("invalid project metadata: %w", err)
	}

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
		project.Footnote = removeParagraphTagsAroundHTML(builder.String())
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

func setGitHubLinkIcons(linkGroups []LinkGroup, githubIconPath string) {
	for _, group := range linkGroups {
		for i, link := range group.Links {
			if link.IconPath == "" && strings.HasPrefix(link.Link, githubBaseURL) {
				link.IconPath = githubIconPath
				group.Links[i] = link
			}
		}
	}
}
