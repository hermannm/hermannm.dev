package sitebuilder

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"strings"

	"hermannm.dev/wrap"
)

type ProjectProfile struct {
	Name    string `yaml:"name"     validate:"required"`
	Slug    string `yaml:"slug"     validate:"required"`
	TagLine string `yaml:"tagLine"`
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
	ProjectBase `                        yaml:",inline"`
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
	Tech     string   `yaml:"tech"          validate:"required"`
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

type ParsedProject struct {
	ProjectTemplate
	Page                      Page
	ContentDir                string
	IndexPageFallbackIconPath string
}

type ProjectContentFile struct {
	name      string
	directory string
}

func (renderer *PageRenderer) RenderProjectPage(
	projectFile ProjectContentFile,
	techIcons TechIconMap,
) (err error) {
	defer func() {
		if err != nil {
			renderer.cancelCtx()
		}
	}()

	var project ParsedProject
	if project, err = parseProject(projectFile, techIcons, renderer.metadata); err != nil {
		return wrap.Errorf(err, "failed to parse project '%s'", projectFile.name)
	}

	renderer.pagePaths <- project.Page.Path
	renderer.parsedProjects <- project

	projectPage := ProjectPageTemplate{
		Meta: TemplateMetadata{
			Common: renderer.metadata,
			Page:   project.Page,
		},
		Project: project.ProjectTemplate,
	}

	if err = renderer.renderPage(projectPage.Meta, projectPage); err != nil {
		return wrap.Errorf(err, "failed to render page for project '%s'", project.Slug)
	}

	return nil
}

func readProjectContentDirs(contentDirNames []string) ([]ProjectContentFile, error) {
	var files []ProjectContentFile
	baseContentDir := os.DirFS(BaseContentDir)

	for _, dirName := range contentDirNames {
		entries, err := fs.ReadDir(baseContentDir, dirName)
		if err != nil {
			return nil, wrap.Errorf(err, "failed to read project content directory '%s'", dirName)
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
	projectFile ProjectContentFile,
	techIcons TechIconMap,
	metadata CommonMetadata,
) (ParsedProject, error) {
	markdownFilePath := fmt.Sprintf(
		"%s/%s/%s", BaseContentDir, projectFile.directory, projectFile.name,
	)

	descriptionBuffer := new(bytes.Buffer)
	var project ProjectMarkdown
	if err := readMarkdownWithFrontmatter(markdownFilePath, descriptionBuffer, &project); err != nil {
		return ParsedProject{}, wrap.Error(err, "failed to read markdown for project")
	}

	project.Page.Title = fmt.Sprintf("%s/%s", metadata.SiteName, project.Slug)
	project.Page.Path = fmt.Sprintf("/%s", project.Slug)
	project.Page.TemplateName = ProjectPageTemplateName
	if project.TechStackTitle == "" {
		project.TechStackTitle = DefaultTechStackTitle
	}
	setGitHubLinkIcons(project.LinkGroups, metadata.GitHubIconPath)

	if err := validate.Struct(project); err != nil {
		return ParsedProject{}, wrap.Error(err, "invalid project metadata")
	}

	techStack, indexPageFallbackIcon, err := parseTechStack(project.TechStack, techIcons)
	if err != nil {
		return ParsedProject{}, wrap.Errorf(
			err,
			"failed to parse tech stack for project '%s'",
			project.Name,
		)
	}

	if project.Footnote != "" {
		var builder strings.Builder
		if err := newMarkdownParser().Convert([]byte(project.Footnote), &builder); err != nil {
			return ParsedProject{}, wrap.Errorf(
				err,
				"failed to parse footnote for project '%s' as markdown",
				project.Slug,
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
		Page:                      project.Page,
		ContentDir:                projectFile.directory,
		IndexPageFallbackIconPath: getTechIconPath(indexPageFallbackIcon),
	}, nil
}

func parseTechStack(
	techStack []TechStackItemMarkdown,
	techIcons TechIconMap,
) (parsed []TechStackItemTemplate, indexPageFallbackIcon string, err error) {
	parsed = make([]TechStackItemTemplate, len(techStack))
	var firstIndexPageFallbackIcon string

	for i, tech := range techStack {
		linkItem, indexPageFallbackIcon, err := getTechIcon(tech.Tech, techIcons)
		if err != nil {
			return nil, "", err
		}
		if firstIndexPageFallbackIcon == "" && indexPageFallbackIcon != "" {
			firstIndexPageFallbackIcon = indexPageFallbackIcon
		}

		usedWith := make([]LinkItem, len(tech.UsedWith))
		for i, tech2 := range tech.UsedWith {
			linkItem2, _, err := getTechIcon(tech2, techIcons)
			if err != nil {
				return nil, "", err
			}

			usedWith[i] = linkItem2
		}

		parsed[i] = TechStackItemTemplate{
			LinkItem: linkItem,
			UsedFor:  tech.UsedFor,
			UsedWith: usedWith,
		}
	}

	return parsed, firstIndexPageFallbackIcon, nil
}

func getTechIcon(
	techName string,
	techIcons TechIconMap,
) (linkItem LinkItem, indexPageFallbackIcon string, err error) {
	techIcon, ok := techIcons[techName]
	if !ok {
		return LinkItem{}, "", fmt.Errorf(
			"failed to find technology '%s' in tech icon map",
			techName,
		)
	}

	return LinkItem{
		Text:     techName,
		Link:     techIcon.Link,
		IconPath: getTechIconPath(techIcon.Icon),
	}, techIcon.IndexPageFallbackIcon, nil
}

func getTechIconPath(iconFileName string) string {
	return fmt.Sprintf("/%s/%s", TechIconDir, iconFileName)
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
