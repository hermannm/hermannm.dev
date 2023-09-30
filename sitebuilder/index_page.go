package sitebuilder

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"hermannm.dev/wrap"
)

type IndexPageBase struct {
	ProfilePictureMobile  Image `yaml:"profilePictureMobile"`
	ProfilePictureDesktop Image `yaml:"profilePictureDesktop"`
}

type IndexPageMarkdown struct {
	IndexPageBase `                       yaml:",inline"`
	Page          Page                   `yaml:"page"`
	PersonalInfo  PersonalInfoMarkdown   `yaml:"personalInfo"`
	ProjectGroups []ProjectGroupMarkdown `yaml:"projectGroups,flow" validate:"required,dive"`
}

type PersonalInfoMarkdown struct {
	Birthday    string `yaml:"birthday"    validate:"required"`
	Location    string `yaml:"location"    validate:"required"`
	GitHubURL   string `yaml:"githubURL"   validate:"required,url"`
	LinkedInURL string `yaml:"linkedinURL" validate:"required,url"`
}

type IndexPageTemplate struct {
	IndexPageBase
	Meta          TemplateMetadata
	AboutMe       template.HTML
	PersonalInfo  []LinkItem // May omit Link field.
	ProjectGroups []ProjectGroupTemplate
}

type ProjectGroupMarkdown struct {
	Title        string   `yaml:"title"             validate:"required"`
	ProjectSlugs []string `yaml:"projectSlugs,flow" validate:"required,dive"`
	ContentDir   string   `yaml:"contentDir"        validate:"required"`
}

type ProjectGroupTemplate struct {
	Title    string
	Projects []ProjectTemplate
}

type Image struct {
	Path   string `yaml:"path"   validate:"required,filepath"`
	Alt    string `yaml:"alt"    validate:"required"`
	Width  int    `yaml:"width"  validate:"required"`
	Height int    `yaml:"height" validate:"required"`
}

func (renderer *PageRenderer) RenderIndexPage(
	contentPath string,
	birthday time.Time,
) (err error) {
	defer func() {
		if err != nil {
			renderer.cancelCtx()
		}
	}()

	content, aboutMeText, err := parseIndexPageContent(contentPath, renderer.metadata, birthday)
	if err != nil {
		return wrap.Error(err, "failed to parse index page data")
	}

	personalInfo, err := content.PersonalInfo.toTemplateFields()
	if err != nil {
		return wrap.Error(err, "failed to parse personal info from index page content")
	}

	renderer.pagePaths <- content.Page.Path

	groups := parseProjectGroups(content.ProjectGroups)

ProjectLoop:
	for i := 0; i < renderer.projectCount; i++ {
		select {
		case project := <-renderer.parsedProjects:
			if err = groups.AddIfIncluded(project); err != nil {
				return wrap.Errorf(err, "failed to add project '%s' to groups", project.Slug)
			}
			if groups.IsFull() {
				break ProjectLoop
			}
		case <-renderer.ctx.Done():
			return nil
		}
	}

	pageTemplate := IndexPageTemplate{
		IndexPageBase: content.IndexPageBase,
		Meta: TemplateMetadata{
			Common: renderer.metadata,
			Page:   content.Page,
		},
		AboutMe:       aboutMeText,
		PersonalInfo:  personalInfo,
		ProjectGroups: groups.ToSlice(),
	}
	if err = renderer.renderPage(pageTemplate.Meta, pageTemplate); err != nil {
		return wrap.Error(err, "failed to render index page")
	}

	return nil
}

func parseIndexPageContent(
	contentPath string,
	metadata CommonMetadata,
	birthday time.Time,
) (content IndexPageMarkdown, aboutMeText template.HTML, err error) {
	path := fmt.Sprintf("%s/%s", BaseContentDir, contentPath)
	aboutMeBuffer := new(bytes.Buffer)
	if err := readMarkdownWithFrontmatter(path, aboutMeBuffer, &content); err != nil {
		return IndexPageMarkdown{}, "", wrap.Error(err, "failed to read markdown for index page")
	}

	if err := validate.Struct(content); err != nil {
		return IndexPageMarkdown{}, "", wrap.Error(err, "invalid index page metadata")
	}

	aboutMeText = removeParagraphTagsAroundHTML(aboutMeBuffer.String())

	return content, aboutMeText, nil
}

func (personalInfo PersonalInfoMarkdown) toTemplateFields() ([]LinkItem, error) {
	birthday, err := time.Parse(time.DateOnly, personalInfo.Birthday)
	if err != nil {
		return nil, wrap.Errorf(err, "failed to parse birthday field '%s'", personalInfo.Birthday)
	}
	birthdayField := LinkItem{
		Text:     fmt.Sprintf("%d years old", ageFromBirthday(birthday)),
		IconPath: "/img/icons/person.svg",
	}
	locationField := LinkItem{
		Text:     personalInfo.Location,
		IconPath: "/img/icons/map-marker.svg",
	}
	githubField := LinkItem{
		Text:     "GitHub",
		Link:     personalInfo.GitHubURL,
		IconPath: "/img/icons/github.svg",
	}
	linkedinField := LinkItem{
		Text:     "LinkedIn",
		Link:     personalInfo.LinkedInURL,
		IconPath: "/img/icons/linkedin.svg",
	}
	return []LinkItem{birthdayField, locationField, githubField, linkedinField}, nil
}

func ageFromBirthday(birthday time.Time) int {
	now := time.Now()
	age := now.Year() - birthday.Year()

	birthdayCelebratedThisYear := now.YearDay() >= birthday.YearDay()
	if !birthdayCelebratedThisYear {
		age--
	}

	return age
}

func parseProjectGroups(groups []ProjectGroupMarkdown) ParsedProjectGroups {
	parsedGroups := make([]ParsedProjectGroup, len(groups))
	targetNumberOfProjects := 0

	for i, group := range groups {
		projectsLength := len(group.ProjectSlugs)

		projectIndicesBySlug := make(map[string]int, projectsLength)
		for j, slug := range group.ProjectSlugs {
			projectIndicesBySlug[slug] = j
			targetNumberOfProjects++
		}

		parsedGroups[i] = ParsedProjectGroup{
			ProjectGroupTemplate: ProjectGroupTemplate{
				Title:    group.Title,
				Projects: make([]ProjectTemplate, projectsLength),
			},
			projectIndicesBySlug: projectIndicesBySlug,
			contentDir:           group.ContentDir,
		}
	}

	return ParsedProjectGroups{
		list:                   parsedGroups,
		numberOfProjects:       0,
		targetNumberOfProjects: targetNumberOfProjects,
	}
}

type ParsedProjectGroups struct {
	list                   []ParsedProjectGroup
	numberOfProjects       int
	targetNumberOfProjects int
}

type ParsedProjectGroup struct {
	ProjectGroupTemplate
	projectIndicesBySlug map[string]int
	contentDir           string
}

func (groups *ParsedProjectGroups) AddIfIncluded(project ParsedProject) error {
	var group ParsedProjectGroup
	isIncluded := false
	for _, candidate := range groups.list {
		if candidate.contentDir == project.ContentDir {
			group = candidate
			isIncluded = true
			break
		}
	}
	if !isIncluded {
		return nil
	}

	index, isIncluded := group.projectIndicesBySlug[project.Slug]
	if !isIncluded {
		return nil
	}

	projects := group.Projects
	if index >= len(projects) {
		return fmt.Errorf("project index in group '%s' is out-of-bounds", group.Title)
	}

	if project.IconPath == "" {
		if project.IndexPageFallbackIconPath != "" {
			project.IconPath = project.IndexPageFallbackIconPath
		} else {
			return fmt.Errorf("no icon found for project '%s'", project.Slug)
		}
	}

	group.Projects[index] = project.ProjectTemplate
	groups.numberOfProjects++
	return nil
}

func (groups *ParsedProjectGroups) IsFull() bool {
	return groups.numberOfProjects == groups.targetNumberOfProjects
}

func (groups *ParsedProjectGroups) ToSlice() []ProjectGroupTemplate {
	slice := make([]ProjectGroupTemplate, 0, len(groups.list))

	for _, group := range groups.list {
		slice = append(slice, group.ProjectGroupTemplate)
	}

	return slice
}
