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

type IndexPageTemplate struct {
	IndexPageBase
	Meta              TemplateMetadata
	AboutMe           template.HTML
	ProjectCategories []ProjectCategoryTemplate
}

type IndexPageMarkdown struct {
	IndexPageBase     `yaml:",inline"`
	Page              Page                      `yaml:"page"`
	ProjectCategories []ProjectCategoryMarkdown `yaml:"projectCategories,flow"`
}

type ProjectCategoryBase struct {
	Title          string `yaml:"title"`
	ContentDirName string `yaml:"contentDirName"`
}

type ProjectCategoryTemplate struct {
	ProjectCategoryBase
	Projects []ProjectTemplate
}

type ProjectCategoryMarkdown struct {
	ProjectCategoryBase `yaml:",inline"`
	ProjectSlugs        []string `yaml:"projectSlugs,flow"`
}

type Image struct {
	Path   string `yaml:"path"`
	Alt    string `yaml:"alt"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
}

func GetIndexPageTemplate(
	commonMetadata CommonMetadata,
	projectTemplatesBySlug map[string]ProjectTemplate,
	birthday time.Time,
) (IndexPageTemplate, error) {
	indexMarkdownPath := fmt.Sprintf("%s/index.md", ContentDir)
	aboutMeBuffer := new(bytes.Buffer)
	var meta IndexPageMarkdown
	if err := ReadMarkdownWithFrontmatter(indexMarkdownPath, aboutMeBuffer, &meta); err != nil {
		return IndexPageTemplate{}, fmt.Errorf("failed to read markdown for index page: %w", err)
	}

	projectCategories, err := getProjectCategoriesFromMarkdown(
		meta.ProjectCategories, projectTemplatesBySlug,
	)
	if err != nil {
		return IndexPageTemplate{}, err
	}

	aboutMeText := removeParagraphTagsAroundHTML(aboutMeBuffer.String())
	setAge(meta.PersonalInfo, birthday)

	return IndexPageTemplate{
		IndexPageBase: meta.IndexPageBase,
		Meta: TemplateMetadata{
			Common: commonMetadata,
			Page:   meta.Page,
		},
		AboutMe:           template.HTML(aboutMeText),
		ProjectCategories: projectCategories,
	}, nil
}

func getProjectCategoriesFromMarkdown(
	markdownCategories []ProjectCategoryMarkdown, projectTemplatesBySlug map[string]ProjectTemplate,
) ([]ProjectCategoryTemplate, error) {
	categories := make([]ProjectCategoryTemplate, len(markdownCategories))

	for i, markdownCategory := range markdownCategories {
		projects := make([]ProjectTemplate, len(markdownCategory.ProjectSlugs))

		for i, projectSlug := range markdownCategory.ProjectSlugs {
			project, ok := projectTemplatesBySlug[projectSlug]
			if !ok {
				return nil, fmt.Errorf(
					"failed to find project template with slug '%s'", projectSlug,
				)
			}

			if project.TechStackTitle == "" {
				project.TechStackTitle = "Built with"
			}

			projects[i] = project
		}

		categories[i] = ProjectCategoryTemplate{
			ProjectCategoryBase: markdownCategory.ProjectCategoryBase,
			Projects:            projects,
		}
	}

	return categories, nil
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
