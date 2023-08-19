package sitebuilder

import (
	"html/template"
	"strings"
)

type LinkItem struct {
	Text     string `yaml:"text"`
	Link     string `yaml:"link"`
	IconPath string `yaml:"iconPath"`
}

type TemplateMetadata struct {
	Common CommonMetadata
	Page   Page
}

type CommonMetadata struct {
	SiteName         string
	SiteDescription  string
	BaseURL          string
	GitHubIconPath   string
	GitHubIssuesLink string
}

type Page struct {
	Title        string `yaml:"title"`
	Path         string `yaml:"path"`
	TemplateName string `yaml:"templateName"`

	// Optional.
	RedirectURL string `yaml:"redirectURL"`

	// Nil if page does not host a Go package.
	GoPackage *GoPackage `yaml:"goPackage"`
}

type GoPackage struct {
	FullName  string `yaml:"fullName"`
	GitHubURL string `yaml:"githubURL"`
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
