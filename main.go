package main

import (
	"fmt"
	"os"
	"time"

	"hermannm.dev/personal-website/sitebuilder"
	"hermannm.dev/wrap"
)

func main() {
	if err := sitebuilder.RenderPages(contentPaths, metadata, techResources, birthday); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := sitebuilder.FormatRenderedPages(); err != nil {
		fmt.Println(wrap.Error(err, "failed to format rendered html"))
		os.Exit(1)
	}

	if err := sitebuilder.GenerateTailwindCSS("styles.css"); err != nil {
		fmt.Println(wrap.Error(err, "failed to generate tailwind css"))
		os.Exit(1)
	}

	fmt.Printf("Website built successfully! Output in ./%s\n", sitebuilder.BaseOutputDir)
}

var (
	metadata = sitebuilder.CommonMetadata{
		SiteName:         "hermannm.dev",
		SiteDescription:  "Hermann Mørkrid's personal website.",
		BaseURL:          "https://hermannm.dev",
		GitHubIconPath:   "/img/icons/github.svg",
		GitHubIssuesLink: "https://github.com/hermannm/hermannm.dev/issues",
	}

	contentPaths = sitebuilder.ContentPaths{
		IndexPage:   "index_page.md",
		ProjectDirs: []string{"projects", "companies", "schools"},
		BasicPages:  []string{"404_page.md"},
	}

	birthday = time.Date(1999, time.September, 12, 2, 0, 0, 0, time.UTC)

	techResources = sitebuilder.TechResourceMap{
		"Go": {
			Link:     "https://go.dev/",
			IconFile: "go.svg",
		},
		"TypeScript": {
			Link:     "https://www.typescriptlang.org/",
			IconFile: "typescript.svg",
		},
		"Rust": {
			Link:     "https://www.rust-lang.org/",
			IconFile: "rust.svg",
		},
		"JavaScript": {
			Link:     "https://developer.mozilla.org/en-US/docs/Web/JavaScript",
			IconFile: "javascript.svg",
		},
		"C#": {
			Link:     "https://docs.microsoft.com/en-us/dotnet/csharp/tour-of-csharp/",
			IconFile: "csharp.svg",
		},
		"Java": {
			Link:     "https://www.java.com/en/download/help/whatis_java.html",
			IconFile: "java.svg",
		},
		"Kotlin": {
			Link:     "https://kotlinlang.org/",
			IconFile: "kotlin.svg",
		},
		"Python": {
			Link:     "https://www.python.org/",
			IconFile: "python.svg",
		},
		"React": {
			Link:     "https://reactjs.org/",
			IconFile: "react.svg",
		},
		"Next.js": {
			Link:     "https://nextjs.org/",
			IconFile: "next-js.svg",
		},
		"Django": {
			Link:     "https://www.djangoproject.com/",
			IconFile: "django.svg",
		},
		"Unity": {
			Link:     "https://unity.com/",
			IconFile: "unity.svg",
		},
		"libGDX": {
			Link:     "https://libgdx.com/",
			IconFile: "libgdx.svg",
		},
		"gRPC": {
			Link:     "https://grpc.io/",
			IconFile: "grpc.svg",
		},
		"GraphQL": {
			Link:     "https://graphql.org/",
			IconFile: "graphql.svg",
		},
		"WebRTC": {
			Link:     "https://webrtc.org/",
			IconFile: "webrtc.svg",
		},
		"MQTT": {
			Link:     "https://mqtt.org/",
			IconFile: "mqtt.svg",
		},
		"AWS CDK": {
			Link:     "https://aws.amazon.com/cdk/",
			IconFile: "aws.svg",
		},
	}
)
