package sitebuilder

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
)

type IndexPageBase struct {
	// May omit Link field.
	PersonalInfo          []LinkItem `yaml:"personalInfo,flow"`
	ProfilePictureMobile  Image      `yaml:"profilePictureMobile"`
	ProfilePictureDesktop Image      `yaml:"profilePictureDesktop"`
}

type IndexPageMarkdown struct {
	IndexPageBase `yaml:",inline"`
	Page          Page                   `yaml:"page"`
	ProjectGroups []ProjectGroupMarkdown `yaml:"projectGroups,flow"`
}

type IndexPageTemplate struct {
	IndexPageBase
	Meta          TemplateMetadata
	AboutMe       template.HTML
	ProjectGroups []ProjectGroupTemplate
}

type ProjectGroupMarkdown struct {
	Title        string   `yaml:"title"`
	ProjectSlugs []string `yaml:"projectSlugs,flow"`
	ContentDir   string   `yaml:"contentDir"`
}

type ProjectGroupTemplate struct {
	Title    string
	Projects []ProjectProfile
}

type Image struct {
	Path   string `yaml:"path"`
	Alt    string `yaml:"alt"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
}

func (renderer *PageRenderer) RenderIndexPage(contentPath string, birthday time.Time) (err error) {
	defer func() {
		if err != nil {
			renderer.cancelChannels()
		}
	}()

	content, aboutMeText, err := parseIndexPageContent(contentPath, renderer.metadata, birthday)
	if err != nil {
		return fmt.Errorf("failed to parse index page data: %w", err)
	}

	renderer.pagePaths <- content.Page.Path

	groups := parseProjectGroups(content.ProjectGroups)

ProjectLoop:
	for i := 0; i < renderer.projectCount; i++ {
		select {
		case project := <-renderer.parsedProjects:
			if err = groups.AddIfIncluded(project); err != nil {
				return fmt.Errorf("failed to add project '%s' to groups: %w", project.Slug, err)
			}
			if groups.IsFull() {
				break ProjectLoop
			}
		case <-renderer.channelContext.Done():
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
		ProjectGroups: groups.ToSlice(),
	}
	if err = renderer.renderPage(pageTemplate.Meta, pageTemplate); err != nil {
		return fmt.Errorf("failed to render index page: %w", err)
	}

	return nil
}

func parseIndexPageContent(
	contentPath string, metadata CommonMetadata, birthday time.Time,
) (content IndexPageMarkdown, aboutMeText template.HTML, err error) {
	path := fmt.Sprintf("%s/%s", BaseContentDir, contentPath)
	aboutMeBuffer := new(bytes.Buffer)
	if err := readMarkdownWithFrontmatter(path, aboutMeBuffer, &content); err != nil {
		return IndexPageMarkdown{}, "", fmt.Errorf(
			"failed to read markdown for index page: %w", err,
		)
	}

	aboutMeText = removeParagraphTagsAroundHTML(aboutMeBuffer.String())
	setAge(content.PersonalInfo, birthday)

	return content, aboutMeText, nil
}

const ageReplacementPattern = "${age}"

func setAge(personalInfo []LinkItem, birthday time.Time) {
	ageText := strconv.Itoa(ageFromBirthday(birthday))

	for i, personalInfoField := range personalInfo {
		personalInfoField.Text = strings.Replace(
			personalInfoField.Text, ageReplacementPattern, ageText, 1,
		)
		personalInfo[i] = personalInfoField
	}
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
				Projects: make([]ProjectProfile, projectsLength),
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

func (groups *ParsedProjectGroups) AddIfIncluded(project ProjectWithContentDir) error {
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
