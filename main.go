package main

import (
	"fmt"
	"log"
	"time"

	"hermannm.dev/personal-website/sitebuilder"
)

var commonMetadata = sitebuilder.CommonMetadata{
	BaseURL:          "https://hermannm.dev",
	GitHubIssuesLink: "https://github.com/hermannm/hermannm.dev/issues",
	GitHubIconPath:   "/img/icons/github.svg",
}

var projectDirNames = []string{"projects", "companies", "schools"}

var birthday = time.Date(1999, time.September, 12, 2, 0, 0, 0, time.UTC)

func main() {
	projectTemplates, err := sitebuilder.GetProjectTemplates(projectDirNames)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to get project templates: %w", err))
	}

	indexTemplate, err := sitebuilder.GetIndexPageTemplate(
		commonMetadata, projectTemplates, birthday,
	)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to get index page data: %w", err))
	}

	if err := sitebuilder.CreateAndRenderTemplate(indexTemplate.Meta, indexTemplate); err != nil {
		log.Fatalln(err)
	}

	if err := sitebuilder.FormatRenderedTemplates(); err != nil {
		err = fmt.Errorf("failed to format rendered templates: %w", err)
		log.Fatalf("%v\n\n%s\n", err, "Do you have Prettier installed?")
	}

	fmt.Printf("Website built successfully! Output in ./%s\n", sitebuilder.BaseOutputDir)
}
