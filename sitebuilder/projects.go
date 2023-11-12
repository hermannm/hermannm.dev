package sitebuilder

import (
	"bytes"
	"errors"
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
	TagLine string `yaml:"tagLine"  validate:"required"`
	// Optional if not included in index page.
	LogoPath              string `yaml:"logoPath" validate:"omitempty,filepath"`
	LogoAlt               string `yaml:"logoAlt"`
	IndexPageFallbackIcon template.HTML
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
	// May omit Icon field.
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
	Page       Page
	ContentDir string
}

type ProjectContentFile struct {
	name      string
	directory string
}

func (renderer *PageRenderer) RenderProjectPage(projectFile ProjectContentFile) (err error) {
	defer func() {
		if err != nil {
			renderer.cancelCtx()
		}
	}()

	var project ParsedProject
	if project, err = renderer.parseProject(projectFile); err != nil {
		return wrap.Errorf(err, "failed to parse project '%s'", projectFile.name)
	}

	renderer.parsedPages <- project.Page
	renderer.parsedProjects <- project

	projectPage := ProjectPageTemplate{
		Meta: TemplateMetadata{
			Common: renderer.commonData,
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

func (renderer *PageRenderer) parseProject(projectFile ProjectContentFile) (ParsedProject, error) {
	markdownFilePath := fmt.Sprintf(
		"%s/%s/%s", BaseContentDir, projectFile.directory, projectFile.name,
	)

	descriptionBuffer := new(bytes.Buffer)
	var project ProjectMarkdown
	if err := readMarkdownWithFrontmatter(markdownFilePath, descriptionBuffer, &project); err != nil {
		return ParsedProject{}, wrap.Error(err, "failed to read markdown for project")
	}

	project.Page.Title = fmt.Sprintf("%s/%s", renderer.commonData.SiteName, project.Slug)
	project.Page.Path = fmt.Sprintf("/%s", project.Slug)
	project.Page.TemplateName = ProjectPageTemplateName
	if project.TechStackTitle == "" {
		project.TechStackTitle = DefaultTechStackTitle
	}

	if err := validate.Struct(project); err != nil {
		return ParsedProject{}, wrap.Error(err, "invalid project metadata")
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

	// Waits for icons to finish rendering before using them
	select {
	case <-renderer.ctx.Done():
		return ParsedProject{}, renderer.ctx.Err()
	case <-renderer.iconsRendered:
	}

	if err := setLinkIcons(project.LinkGroups, renderer.icons); err != nil {
		return ParsedProject{}, wrap.Error(err, "failed to set link icons")
	}

	techStack, indexPageFallbackIcon, err := parseTechStack(project.TechStack, renderer.icons)
	if err != nil {
		return ParsedProject{}, wrap.Errorf(
			err,
			"failed to parse tech stack for project '%s'",
			project.Name,
		)
	}

	project.IndexPageFallbackIcon = indexPageFallbackIcon

	return ParsedProject{
		ProjectTemplate: ProjectTemplate{
			ProjectBase: project.ProjectBase,
			Description: template.HTML(descriptionBuffer.String()),
			TechStack:   techStack,
		},
		Page:       project.Page,
		ContentDir: projectFile.directory,
	}, nil
}

func parseTechStack(
	techStack []TechStackItemMarkdown,
	icons IconMap,
) (parsed []TechStackItemTemplate, indexPageFallbackIcon template.HTML, err error) {
	parsed = make([]TechStackItemTemplate, len(techStack))
	var firstIndexPageFallbackIcon template.HTML

	for i, tech := range techStack {
		linkItem, indexPageFallbackIcon, err := getTechIcon(tech.Tech, icons)
		if err != nil {
			return nil, "", err
		}
		if firstIndexPageFallbackIcon == "" && indexPageFallbackIcon != "" {
			firstIndexPageFallbackIcon = indexPageFallbackIcon
		}

		usedWith := make([]LinkItem, len(tech.UsedWith))
		for i, tech2 := range tech.UsedWith {
			linkItem2, _, err := getTechIcon(tech2, icons)
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
	icons IconMap,
) (linkItem LinkItem, indexPageFallbackIcon template.HTML, err error) {
	techIcon, ok := icons[techName]
	if !ok {
		return LinkItem{}, "", fmt.Errorf(
			"failed to find icon for technology '%s' in icon map",
			techName,
		)
	}

	return LinkItem{
		Text: techName,
		Link: techIcon.Link,
		Icon: template.HTML(techIcon.Icon),
	}, template.HTML(techIcon.IndexPageFallbackIcon), nil
}

func setLinkIcons(linkGroups []LinkGroup, icons IconMap) error {
	githubIcon, ok := icons["GitHub"]
	if !ok {
		return errors.New("failed to find GitHub icon in icon map")
	}

	gopherIcon, ok := icons["Gopher"]
	if !ok {
		return errors.New("failed to find Gopher icon in icon map")
	}

	for _, group := range linkGroups {
		for i, link := range group.Links {
			if link.Icon != "" {
				renderedIcon, ok := icons[string(link.Icon)]
				if !ok {
					return fmt.Errorf(
						"icon '%s' not found in icon map for link '%s'",
						link.Icon,
						link.Text,
					)
				}

				link.Icon = template.HTML(renderedIcon.Icon)
				group.Links[i] = link
			} else if strings.HasPrefix(link.Link, "https://github.com") {
				link.Icon = template.HTML(githubIcon.Icon)
				group.Links[i] = link
			} else if strings.HasPrefix(link.Link, "https://pkg.go.dev") {
				link.Icon = template.HTML(gopherIcon.Icon)
				group.Links[i] = link
			}
		}
	}

	return nil
}
