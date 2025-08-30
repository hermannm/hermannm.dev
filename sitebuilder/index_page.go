package sitebuilder

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"hermannm.dev/wrap"
	"hermannm.dev/wrap/ctxwrap"
)

type IndexPageBase struct {
	ProfilePictureMobile  Image `yaml:"profilePictureMobile"`
	ProfilePictureDesktop Image `yaml:"profilePictureDesktop"`
}

type IndexPageMarkdown struct {
	IndexPageBase `yaml:",inline"`
	Page          `yaml:",inline"`
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
	ProjectPaths []string `yaml:"projectPaths,flow" validate:"required,dive"`
	ContentDir   string   `yaml:"contentDir"        validate:"required"`
}

type ProjectGroupTemplate struct {
	Title    string
	Projects []ProjectProfile
}

type Image struct {
	Path   string `yaml:"path"   validate:"required,filepath"`
	Alt    string `yaml:"alt"    validate:"required"`
	Width  int    `yaml:"width"  validate:"required"`
	Height int    `yaml:"height" validate:"required"`
}

func (renderer *PageRenderer) RenderIndexPage(ctx context.Context, contentPath string) (err error) {
	content, aboutMeText, err := parseIndexPageContent(ctx, contentPath)
	if err != nil {
		return ctxwrap.Error(ctx, err, "failed to parse index page data")
	}
	content.Page.SetCanonicalURL(renderer.commonData.BaseURL)

	projectGroups := parseProjectGroups(content.ProjectGroups)

	renderer.parsedPages <- content.Page

	// Waits for icons to finish rendering before using them
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-renderer.iconsRendered:
	}

	personalInfo, err := content.PersonalInfo.toTemplateFields(renderer.icons)
	if err != nil {
		return ctxwrap.Error(ctx, err, "failed to parse personal info from index page content")
	}

ProjectLoop:
	for range renderer.projectCount {
		select {
		case project := <-renderer.parsedProjects:
			if err = projectGroups.AddIfIncluded(project); err != nil {
				return ctxwrap.Errorf(
					ctx,
					err,
					"failed to add project '%s' to groups",
					project.Name,
				)
			}
			if projectGroups.IsFull() {
				break ProjectLoop
			}
		case <-ctx.Done():
			return nil
		}
	}
	if !projectGroups.IsFull() {
		return fmt.Errorf(
			"index page did not receive all projects specified in 'projectGroups' in %s",
			contentPath,
		)
	}

	pageTemplate := IndexPageTemplate{
		IndexPageBase: content.IndexPageBase,
		Meta: TemplateMetadata{
			Common: renderer.commonData,
			Page:   content.Page,
		},
		AboutMe:       aboutMeText,
		PersonalInfo:  personalInfo,
		ProjectGroups: projectGroups.ToSlice(),
	}
	if err = renderer.renderPage(ctx, pageTemplate.Meta.Page, pageTemplate); err != nil {
		return ctxwrap.Error(ctx, err, "failed to render index page")
	}

	return nil
}

func parseIndexPageContent(
	ctx context.Context,
	contentPath string,
) (content IndexPageMarkdown, aboutMeText template.HTML, err error) {
	path := fmt.Sprintf("%s/%s", BaseContentDir, contentPath)
	aboutMeBuffer := new(bytes.Buffer)
	if err := readMarkdownWithFrontmatter(ctx, path, aboutMeBuffer, &content); err != nil {
		return IndexPageMarkdown{}, "", ctxwrap.Error(
			ctx,
			err,
			"failed to read markdown for index page",
		)
	}

	if err := validate.Struct(content); err != nil {
		return IndexPageMarkdown{}, "", ctxwrap.Error(ctx, err, "invalid index page metadata")
	}

	aboutMeText = removeParagraphTagsAroundHTML(aboutMeBuffer.String())

	return content, aboutMeText, nil
}

func (personalInfo PersonalInfoMarkdown) toTemplateFields(icons IconMap) ([]LinkItem, error) {
	birthday, err := time.Parse(time.DateOnly, personalInfo.Birthday)
	if err != nil {
		return nil, wrap.Errorf(err, "failed to parse birthday field '%s'", personalInfo.Birthday)
	}

	//nolint:exhaustruct
	fields := []LinkItem{
		{LinkText: fmt.Sprintf("%d years old", ageFromBirthday(birthday))},
		{LinkText: personalInfo.Location},
		{LinkText: "GitHub", Link: personalInfo.GitHubURL},
		{LinkText: "LinkedIn", Link: personalInfo.LinkedInURL},
	}

	for i, iconName := range [4]string{"person", "map-marker", "GitHub", "LinkedIn"} {
		icon, ok := icons[iconName]
		if !ok {
			return nil, fmt.Errorf("expected '%s' icon in icon map, but found none", iconName)
		}

		fields[i].Icon = template.HTML(icon.Icon)
	}

	return fields, nil
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
		projectsLength := len(group.ProjectPaths)

		projectIndicesBySlug := make(map[string]int, projectsLength)
		for j, slug := range group.ProjectPaths {
			projectIndicesBySlug[slug] = j
			targetNumberOfProjects++
		}

		parsedGroups[i] = ParsedProjectGroup{
			ProjectGroupTemplate: ProjectGroupTemplate{
				Title:    group.Title,
				Projects: make([]ProjectProfile, projectsLength),
			},
			projectIndiciesByPath: projectIndicesBySlug,
			contentDir:            group.ContentDir,
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
	projectIndiciesByPath map[string]int
	contentDir            string
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

	index, isIncluded := group.projectIndiciesByPath[project.Page.Path]
	if !isIncluded {
		return nil
	}

	projects := group.Projects
	if index >= len(projects) {
		return fmt.Errorf("project index in group '%s' is out-of-bounds", group.Title)
	}

	if project.Logo.Path == "" && project.IndexPageFallbackIcon == "" {
		return fmt.Errorf("no icon found for project '%s'", project.Name)
	}

	group.Projects[index] = project.ProjectProfile
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
