package main

import (
	"fmt"
	"log"
	"time"

	"hermannm.dev/personal-website/sitebuilder"
)

var (
	metadata = sitebuilder.CommonMetadata{
		SiteName:         "hermannm.dev",
		BaseURL:          "https://hermannm.dev",
		GitHubIssuesLink: "https://github.com/hermannm/hermannm.dev/issues",
		GitHubIconPath:   "/img/icons/github.svg",
	}

	projectContentDirs = []string{"projects", "companies", "schools"}

	birthday = time.Date(1999, time.September, 12, 2, 0, 0, 0, time.UTC)
)

func main() {
	if err := sitebuilder.RenderPages(projectContentDirs, metadata, birthday); err != nil {
		log.Fatalln(err)
	}

	if err := sitebuilder.FormatRenderedPages(); err != nil {
		err = fmt.Errorf("failed to format rendered templates: %w", err)
		log.Fatalf("%v\n\n%s\n", err, "Do you have Prettier installed?")
	}

	fmt.Printf("Website built successfully! Output in ./%s\n", sitebuilder.BaseOutputDir)
}
