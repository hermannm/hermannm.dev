package main

import (
	"fmt"
	"log"
	"time"

	"hermannm.dev/personal-website/sitebuilder"
)

var (
	commonMetadata = sitebuilder.CommonMetadata{
		SiteName:         "hermannm.dev",
		BaseURL:          "https://hermannm.dev",
		GitHubIssuesLink: "https://github.com/hermannm/hermannm.dev/issues",
		GitHubIconPath:   "/img/icons/github.svg",
	}

	projectContentDirs = []string{"projects", "companies", "schools"}

	birthday = time.Date(1999, time.September, 12, 2, 0, 0, 0, time.UTC)
)

func main() {
	projects, err := sitebuilder.ParseProjects(projectContentDirs)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to parse projects: %w", err))
	}

	if err := sitebuilder.RenderPages(projects, commonMetadata, birthday); err != nil {
		log.Fatalln(err)
	}

	if err := sitebuilder.FormatRenderedPages(); err != nil {
		err = fmt.Errorf("failed to format rendered templates: %w", err)
		log.Fatalf("%v\n\n%s\n", err, "Do you have Prettier installed?")
	}

	fmt.Printf("Website built successfully! Output in ./%s\n", sitebuilder.BaseOutputDir)
}
