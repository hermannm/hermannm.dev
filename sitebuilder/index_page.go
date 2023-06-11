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
	IndexPageBase     `yaml:",inline"`
	Page              Page                      `yaml:"page"`
	ProjectCategories []ProjectCategoryMarkdown `yaml:"projectCategories,flow"`
}

type IndexPageTemplate struct {
	IndexPageBase
	Meta              TemplateMetadata
	AboutMe           template.HTML
	ProjectCategories []ProjectCategoryTemplate
}

type ProjectCategoryMarkdown struct {
	Title        string   `yaml:"title"`
	ProjectSlugs []string `yaml:"projectSlugs,flow"`
	ContentDir   string   `yaml:"contentDir"`
}

type ProjectCategoryTemplate struct {
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

	categories := parseProjectCategories(content.ProjectCategories)

ProjectLoop:
	for i := 0; i < renderer.projectCount; i++ {
		select {
		case project := <-renderer.parsedProjects:
			if err = categories.AddIfIncluded(project); err != nil {
				return fmt.Errorf("failed to add project '%s' to categories: %w", project.Slug, err)
			}
			if categories.IsFull() {
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
		AboutMe:           aboutMeText,
		ProjectCategories: categories.ToSlice(),
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

func parseProjectCategories(categories []ProjectCategoryMarkdown) ParsedProjectCategories {
	parsedCategories := make(map[string]ParsedProjectCategory, len(categories))
	targetNumberOfProjects := 0

	for _, category := range categories {
		projectsLength := len(category.ProjectSlugs)

		projectIndicesBySlug := make(map[string]int, projectsLength)
		for i, slug := range category.ProjectSlugs {
			projectIndicesBySlug[slug] = i
			targetNumberOfProjects++
		}

		parsedCategories[category.ContentDir] = ParsedProjectCategory{
			ProjectCategoryTemplate: ProjectCategoryTemplate{
				Title:    category.Title,
				Projects: make([]ProjectProfile, projectsLength),
			},
			projectIndicesBySlug: projectIndicesBySlug,
		}
	}

	return ParsedProjectCategories{
		categoriesByContentDir: parsedCategories,
		numberOfProjects:       0,
		targetNumberOfProjects: targetNumberOfProjects,
	}
}

type ParsedProjectCategories struct {
	categoriesByContentDir map[string]ParsedProjectCategory
	numberOfProjects       int
	targetNumberOfProjects int
}

type ParsedProjectCategory struct {
	ProjectCategoryTemplate
	projectIndicesBySlug map[string]int
}

func (categories *ParsedProjectCategories) AddIfIncluded(project ProjectWithContentDir) error {
	category, isIncluded := categories.categoriesByContentDir[project.ContentDir]
	if !isIncluded {
		return nil
	}

	index, isIncluded := category.projectIndicesBySlug[project.Slug]
	if !isIncluded {
		return nil
	}

	projects := category.Projects
	if index >= len(projects) {
		return fmt.Errorf("project index in category '%s' is out-of-bounds", category.Title)
	}

	category.Projects[index] = project.ProjectProfile
	categories.numberOfProjects++
	return nil
}

func (categories *ParsedProjectCategories) IsFull() bool {
	return categories.numberOfProjects == categories.targetNumberOfProjects
}

func (categories *ParsedProjectCategories) ToSlice() []ProjectCategoryTemplate {
	slice := make([]ProjectCategoryTemplate, 0, len(categories.categoriesByContentDir))

	for _, category := range categories.categoriesByContentDir {
		slice = append(slice, category.ProjectCategoryTemplate)
	}

	return slice
}
