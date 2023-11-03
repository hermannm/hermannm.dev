package main

import (
	"log/slog"
	"os"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
	"hermannm.dev/personal-website/sitebuilder"
)

func main() {
	logger := slog.New(devlog.NewHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	log.Info("building website...")

	if err := sitebuilder.RenderPages(contentPaths, commonData); err != nil {
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
	contentPaths = sitebuilder.ContentPaths{
		IndexPage:   "index_page.md",
		ProjectDirs: []string{"projects", "companies", "libraries"},
		BasicPages:  []string{"404_page.md"},
	}

	commonData = sitebuilder.CommonPageData{
		SiteName:         "hermannm.dev",
		SiteDescription:  "Hermann Mørkrid's personal website.",
		BaseURL:          "https://hermannm.dev",
		GitHubIconPath:   "/img/icons/github.svg",
		GitHubIssuesLink: "https://github.com/hermannm/hermannm.dev/issues",
		Icons: sitebuilder.IconMap{
			"person": {
				Icon: "/img/icons/person.svg",
			},
			"map-marker": {
				Icon: "/img/icons/map-marker.svg",
			},
			"GitHub": {
				Icon: "/img/icons/github.svg",
			},
			"LinkedIn": {
				Icon: "/img/icons/linkedin.svg",
			},
			"Go": {
				Icon:                  "/img/icons/go.svg",
				Link:                  "https://go.dev/",
				IndexPageFallbackIcon: "/img/icons/go-alt.svg",
			},
			"TypeScript": {
				Icon: "/img/icons/typescript.svg",
				Link: "https://www.typescriptlang.org/",
			},
			"Rust": {
				Icon: "/img/icons/rust.svg",
				Link: "https://www.rust-lang.org/",
			},
			"JavaScript": {
				Icon: "/img/icons/javascript.svg",
				Link: "https://developer.mozilla.org/en-US/docs/Web/JavaScript",
			},
			"C#": {
				Icon: "/img/icons/csharp.svg",
				Link: "https://dotnet.microsoft.com/en-us/languages/csharp",
			},
			"Java": {
				Icon: "/img/icons/java.svg",
				Link: "https://www.java.com/en/download/help/whatis_java.html",
			},
			"Kotlin": {
				Icon: "/img/icons/kotlin.svg",
				Link: "https://kotlinlang.org/",
			},
			"Python": {
				Icon: "/img/icons/python.svg",
				Link: "https://www.python.org/",
			},
			"React": {
				Icon: "/img/icons/react.svg",
				Link: "https://reactjs.org/",
			},
			"Next.js": {
				Icon: "/img/icons/next-js.svg",
				Link: "https://nextjs.org/",
			},
			"Django": {
				Icon: "/img/icons/django.svg",
				Link: "https://www.djangoproject.com/",
			},
			"Godot": {
				Icon: "/img/icons/godot.svg",
				Link: "https://godotengine.org/",
			},
			"Unity": {
				Icon: "/img/icons/unity.svg",
				Link: "https://unity.com/",
			},
			"libGDX": {
				Icon: "/img/icons/libgdx.svg",
				Link: "https://libgdx.com/",
			},
			"gRPC": {
				Icon: "/img/icons/grpc.svg",
				Link: "https://grpc.io/",
			},
			"GraphQL": {
				Icon: "/img/icons/graphql.svg",
				Link: "https://graphql.org/",
			},
			"WebRTC": {
				Icon: "/img/icons/webrtc.svg",
				Link: "https://webrtc.org/",
			},
			"MQTT": {
				Icon: "/img/icons/mqtt.svg",
				Link: "https://mqtt.org/",
			},
			"AWS CDK": {
				Icon: "/img/icons/aws.svg",
				Link: "https://aws.amazon.com/cdk/",
			},
		},
	}
)
