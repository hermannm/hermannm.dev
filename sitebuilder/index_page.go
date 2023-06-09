package sitebuilder

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"hermannm.dev/set"
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

func RenderIndexPage(
	ctx context.Context,
	cancelCtx context.CancelFunc,
	parsedProjects <-chan ProjectWithContentDir,
	metadata CommonMetadata,
	birthday time.Time,
	templates *template.Template,
) error {
	pageData, aboutMeText, err := parseIndexPageData(metadata, birthday)
	if err != nil {
		return fmt.Errorf("failed to parse index page data: %w", err)
	}

	categories, projectIDs := parseProjectCategories(pageData.ProjectCategories)
	for project := range parsedProjects {
		id := project.ID()
		if projectIDs.Contains(id) {
			if err := categories.Add(project); err != nil {
				return err
			}

			projectIDs.Remove(id)
		}

		// Since we remove from projectIDs when we receive a project, we are done when there are
		// none left
		if projectIDs.IsEmpty() {
			cancelCtx()
			break
		}
	}

	pageTemplate := IndexPageTemplate{
		IndexPageBase: pageData.IndexPageBase,
		Meta: TemplateMetadata{
			Common: metadata,
			Page:   pageData.Page,
		},
		AboutMe:           aboutMeText,
		ProjectCategories: categories.ToSlice(),
	}

	if err := renderPage(templates, pageTemplate.Meta, pageTemplate); err != nil {
		return fmt.Errorf("failed to render index page: %w", err)
	}

	return nil
}

func parseIndexPageData(
	metadata CommonMetadata, birthday time.Time,
) (pageData IndexPageMarkdown, aboutMeText template.HTML, err error) {
	path := fmt.Sprintf("%s/index.md", BaseContentDir)
	aboutMeBuffer := new(bytes.Buffer)
	if err := readMarkdownWithFrontmatter(path, aboutMeBuffer, &pageData); err != nil {
		return IndexPageMarkdown{}, "", fmt.Errorf(
			"failed to read markdown for index page: %w", err,
		)
	}

	aboutMeText = removeParagraphTagsAroundHTML(aboutMeBuffer.String())
	setAge(pageData.PersonalInfo, birthday)

	return pageData, aboutMeText, nil
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

func parseProjectCategories(
	categories []ProjectCategoryMarkdown,
) (emptyCategories ProjectCategoriesByContentDir, projectIDs set.Set[ProjectID]) {
	emptyCategories = make(map[string]ProjectCategoryTemplate, len(categories))
	projectIDs = set.New[ProjectID]()

	for _, category := range categories {
		emptyCategories[category.ContentDir] = ProjectCategoryTemplate{
			Title:    category.Title,
			Projects: make([]ProjectProfile, 0, len(category.ProjectSlugs)),
		}

		for _, slug := range category.ProjectSlugs {
			projectIDs.Add(ProjectID{slug: slug, contentDir: category.ContentDir})
		}
	}

	return emptyCategories, projectIDs
}

type ProjectCategoriesByContentDir map[string]ProjectCategoryTemplate

func (categories ProjectCategoriesByContentDir) Add(project ProjectWithContentDir) error {
	category, ok := categories[project.ContentDir]
	if !ok {
		return fmt.Errorf(
			"attempted to add project '%s' from content directory '%s' to category map, but no entry exists for it",
			project.Slug, project.ContentDir,
		)
	}

	category.Projects = append(category.Projects, project.ProjectProfile)
	categories[project.ContentDir] = category
	return nil
}

func (categories ProjectCategoriesByContentDir) ToSlice() []ProjectCategoryTemplate {
	slice := make([]ProjectCategoryTemplate, 0, len(categories))

	for _, category := range categories {
		slice = append(slice, category)
	}

	return slice
}
