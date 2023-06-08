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

type ProjectCategoryTemplate struct {
	Title    string
	Projects []ProjectTemplate
}

type ProjectCategoryMarkdown struct {
	Title        string   `yaml:"title"`
	ProjectSlugs []string `yaml:"projectSlugs,flow"`
	ContentDir   string   `yaml:"contentDir"`
}

type Image struct {
	Path   string `yaml:"path"`
	Alt    string `yaml:"alt"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
}

func GetIndexPageTemplate(
	commonMetadata CommonMetadata, projects map[ProjectID]ProjectTemplate, birthday time.Time,
) (IndexPageTemplate, error) {
	indexMarkdownPath := fmt.Sprintf("%s/index.md", BaseContentDir)
	aboutMeBuffer := new(bytes.Buffer)
	var meta IndexPageMarkdown
	if err := ReadMarkdownWithFrontmatter(indexMarkdownPath, aboutMeBuffer, &meta); err != nil {
		return IndexPageTemplate{}, fmt.Errorf("failed to read markdown for index page: %w", err)
	}

	projectCategories, err := projectCategoriesFromMarkdown(
		meta.ProjectCategories, projects,
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

func projectCategoriesFromMarkdown(
	markdownCategories []ProjectCategoryMarkdown, projects map[ProjectID]ProjectTemplate,
) ([]ProjectCategoryTemplate, error) {
	categories := make([]ProjectCategoryTemplate, len(markdownCategories))

	for i, markdownCategory := range markdownCategories {
		includedProjects := make([]ProjectTemplate, len(markdownCategory.ProjectSlugs))

		for i, projectSlug := range markdownCategory.ProjectSlugs {
			id := ProjectID{slug: projectSlug, contentDir: markdownCategory.ContentDir}
			project, ok := projects[id]
			if !ok {
				return nil, fmt.Errorf("failed to find project with slug '%s'", projectSlug)
			}

			if project.TechStackTitle == "" {
				project.TechStackTitle = DefaultTechStackTitle
			}

			includedProjects[i] = project
		}

		categories[i] = ProjectCategoryTemplate{
			Title:    markdownCategory.Title,
			Projects: includedProjects,
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
