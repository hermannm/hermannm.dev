package sitebuilder

import (
	"html/template"
	"strings"
)

type LinkItem struct {
	Text     string `yaml:"text" validate:"required"`
	Link     string `yaml:"link" validate:"omitempty,url"`
	IconPath string `yaml:"iconPath" validate:"omitempty,filepath"`
}

type TemplateMetadata struct {
	Common CommonMetadata
	Page   Page
}

type CommonMetadata struct {
	SiteName         string `validate:"required"`
	SiteDescription  string `validate:"required"`
	BaseURL          string `validate:"required,url"`
	GitHubIconPath   string `validate:"required,filepath"`
	GitHubIssuesLink string `validate:"required,url"`
}

type Page struct {
	Title        string `yaml:"title" validate:"required"`
	Path         string `yaml:"path"`
	TemplateName string `yaml:"templateName" validate:"required,filepath"`
	RedirectURL  string `yaml:"redirectURL"` // Optional.

	// Nil if page does not host a Go package.
	GoPackage *GoPackage `yaml:"goPackage" validate:"omitempty"`
}

type GoPackage struct {
	FullName  string `yaml:"fullName" validate:"required"`
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
