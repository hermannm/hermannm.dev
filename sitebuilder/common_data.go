package sitebuilder

import (
	"html/template"
	"strings"
)

type LinkItem struct {
	Title string `yaml:"title" validate:"required"`
	// Text to display for the link. Optional - defaults to Link but with https:// stripped, or
	// https://github.com/ stripped if it's a GitHub link.
	LinkText string        `yaml:"linkText"`
	Link     string        `yaml:"link" validate:"omitempty,url"`
	Icon     template.HTML `yaml:"icon" validate:"omitempty,filepath"`
	// We use this in our HTML templates to not use bold text for sublink titles.
	IsSublink bool `yaml:"-"`
}

func (linkItem *LinkItem) populateLinkText() {
	if linkItem.LinkText != "" {
		return
	}

	linkItem.LinkText = strings.TrimPrefix(linkItem.Link, "https://")
	linkItem.LinkText = strings.TrimPrefix(linkItem.LinkText, "github.com/")
}

type TemplateMetadata struct {
	Common CommonPageData
	Page   Page
}

type CommonPageData struct {
	SiteName         string `validate:"required"`
	SiteDescription  string `validate:"required"`
	BaseURL          string `validate:"required,url"`
	GitHubIssuesLink string `validate:"required,url"`
	githubIcon       template.HTML
}

func (commonData CommonPageData) GitHubIcon() template.HTML {
	return commonData.githubIcon
}

type Page struct {
	Title        string `yaml:"title"        validate:"required"`
	Path         string `yaml:"path"         validate:"required,startswith=/"`
	TemplateName string `yaml:"templateName" validate:"required,filepath"`
	RedirectPath string `yaml:"redirectPath"` // Optional.

	// Must be set with [Page.SetCanonicalURL] after parsing.
	CanonicalURL string

	// Nil if page does not host a Go package.
	GoPackage *GoPackage `yaml:"goPackage" validate:"omitempty"`
}

func (page *Page) SetCanonicalURL(baseURL string) {
	if page.Path == "/" {
		page.CanonicalURL = baseURL
	} else if page.RedirectPath != "" {
		page.CanonicalURL = baseURL + page.RedirectPath
	} else {
		page.CanonicalURL = baseURL + page.Path
	}
}

type GoPackage struct {
	RootName  string `yaml:"rootName"  validate:"required"`
	GitHubURL string `yaml:"githubURL" validate:"required,url"`
}

var TemplateFunctions = template.FuncMap{
	"plus1": func(x int) int {
		return x + 1
	},
	"personalInfoTextWrapping": func(infoText string) template.HTML {
		words := strings.Split(infoText, " ")
		wordCount := len(words)
		if wordCount < 3 {
			return template.HTML(infoText)
		}

		var builder strings.Builder

		builder.WriteString(`<span class="whitespace-nowrap">`)

		cutoff := (wordCount - 1) / 2
		for i, word := range words {
			builder.WriteString(word)
			if i == cutoff {
				builder.WriteString("</span>")
			}
			if i != wordCount-1 {
				builder.WriteString(" ")
			}
		}

		return template.HTML(builder.String())
	},
}
