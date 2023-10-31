package main

import (
	"log/slog"
	"os"
	"time"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
	"hermannm.dev/personal-website/sitebuilder"
)

func main() {
	logger := slog.New(devlog.NewHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	log.Info("building website...")

	if err := sitebuilder.RenderPages(contentPaths, metadata, techIcons, birthday); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := sitebuilder.FormatRenderedPages(); err != nil {
		log.Error(err, "failed to format rendered html")
		os.Exit(1)
	}

	if err := sitebuilder.GenerateTailwindCSS("styles.css"); err != nil {
		log.Error(err, "failed to generate tailwind css")
		os.Exit(1)
	}

	log.Info(
		"website built successfully!",
		slog.String("outputDirectory", "./"+sitebuilder.BaseOutputDir),
	)
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
		ProjectDirs: []string{"projects", "libraries"},
		BasicPages:  []string{"404_page.md"},
	}

	birthday = time.Date(1999, time.September, 12, 2, 0, 0, 0, time.UTC)

	techIcons = sitebuilder.TechIconMap{
		"Go": {
			Link:                  "https://go.dev/",
			Icon:                  "go.svg",
			IndexPageFallbackIcon: "go-alt.svg",
		},
		"TypeScript": {
			Link: "https://www.typescriptlang.org/",
			Icon: "typescript.svg",
		},
		"Rust": {
			Link: "https://www.rust-lang.org/",
			Icon: "rust.svg",
		},
		"JavaScript": {
			Link: "https://developer.mozilla.org/en-US/docs/Web/JavaScript",
			Icon: "javascript.svg",
		},
		"C#": {
			Link: "https://dotnet.microsoft.com/en-us/languages/csharp",
			Icon: "csharp.svg",
		},
		"Java": {
			Link: "https://www.java.com/en/download/help/whatis_java.html",
			Icon: "java.svg",
		},
		"Kotlin": {
			Link: "https://kotlinlang.org/",
			Icon: "kotlin.svg",
		},
		"Python": {
			Link: "https://www.python.org/",
			Icon: "python.svg",
		},
		"React": {
			Link: "https://reactjs.org/",
			Icon: "react.svg",
		},
		"Next.js": {
			Link: "https://nextjs.org/",
			Icon: "next-js.svg",
		},
		"Django": {
			Link: "https://www.djangoproject.com/",
			Icon: "django.svg",
		},
		"Godot": {
			Link: "https://godotengine.org/",
			Icon: "godot.svg",
		},
		"Unity": {
			Link: "https://unity.com/",
			Icon: "unity.svg",
		},
		"libGDX": {
			Link: "https://libgdx.com/",
			Icon: "libgdx.svg",
		},
		"gRPC": {
			Link: "https://grpc.io/",
			Icon: "grpc.svg",
		},
		"GraphQL": {
			Link: "https://graphql.org/",
			Icon: "graphql.svg",
		},
		"WebRTC": {
			Link: "https://webrtc.org/",
			Icon: "webrtc.svg",
		},
		"MQTT": {
			Link: "https://mqtt.org/",
			Icon: "mqtt.svg",
		},
		"AWS CDK": {
			Link: "https://aws.amazon.com/cdk/",
			Icon: "aws.svg",
		},
	}
)
