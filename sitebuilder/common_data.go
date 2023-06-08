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
	BaseURL          string
	GitHubIssuesLink string
	GitHubIconPath   string
}

type Page struct {
	Title           string   `yaml:"title"`
	Description     string   `yaml:"description"`
	Path            string   `yaml:"path"`
	IncludedScripts []string `yaml:"includedScripts,flow"`
	TemplateName    string   `yaml:"templateName"`

	// Nil if page does not host a Go package.
	GoPackage *GoPackage `yaml:"goPackage"`
}

type GoPackage struct {
	FullName  string `yaml:"fullName"`
	GitHubURL string `yaml:"githubURL"`
}

const noTextWrapCSSClass = "no-text-wrap"

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

		builder.WriteString(`<span class="`)
		builder.WriteString(noTextWrapCSSClass)
		builder.WriteString(`">`)

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
