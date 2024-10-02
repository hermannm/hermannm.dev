package main

import (
	"flag"
	"log/slog"
	"os"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"
	"hermannm.dev/personal-website/devserver"
	"hermannm.dev/personal-website/sitebuilder"
)

func main() {
	devlog.InitDefaultLogHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug})

	useDevServer, port := parseCLIFlags()
	if useDevServer {
		if err := devserver.ServeAndRebuildOnChange(
			contentPaths,
			commonData,
			icons,
			inputCSSFileName,
			port,
		); err != nil {
			log.ErrorCause(err, "devserver stopped")
			os.Exit(1)
		}
	} else {
		log.Info("building website...")
		if err := sitebuilder.RenderPages(contentPaths, commonData, icons); err != nil {
			log.Error(err)
			os.Exit(1)
		}
		if err := sitebuilder.FormatRenderedPages(); err != nil {
			log.ErrorCause(err, "failed to format rendered pages")
			os.Exit(1)
		}
		if err := sitebuilder.GenerateTailwindCSS(inputCSSFileName); err != nil {
			log.ErrorCause(err, "failed to generate css for rendered pages")
			os.Exit(1)
		}
		log.Info("website built successfully!", "outputDirectory", "./"+sitebuilder.BaseOutputDir)
	}
}

func parseCLIFlags() (useDevServer bool, port string) {
	flag.BoolVar(
		&useDevServer,
		"dev",
		false,
		"Serve and rebuild the site every time content/templates/sitebuilder files change",
	)
	flag.StringVar(&port, "port", "8080", "The port to serve the website from when using -dev")
	flag.Parse()
	return useDevServer, port
}

var (
	commonData = sitebuilder.CommonPageData{
		SiteName:         "hermannm.dev",
		SiteDescription:  "Hermann Mørkrid's personal website.",
		BaseURL:          "https://hermannm.dev",
		GitHubIssuesLink: "https://github.com/hermannm/hermannm.dev/issues",
	}

	contentPaths = sitebuilder.ContentPaths{
		IndexPage:   "index_page.md",
		ProjectDirs: []string{"projects", "companies", "libraries"},
		BasicPages:  []string{"404_page.md"},
	}

	inputCSSFileName = "styles.css"

	icons = sitebuilder.IconMap{
		"person": {
			Icon: "content/icons/person.svg",
		},
		"map-marker": {
			Icon: "content/icons/map-marker.svg",
		},
		"GitHub": {
			Icon: "content/icons/github.svg",
		},
		"LinkedIn": {
			Icon: "content/icons/linkedin.svg",
		},
		"Gopher": {
			Icon: "content/icons/gopher.svg",
		},
		"Go": {
			Icon:                  "content/icons/go.svg",
			Link:                  "https://go.dev/",
			IndexPageFallbackIcon: "content/icons/go-alt.svg",
		},
		"TypeScript": {
			Icon: "content/icons/typescript.svg",
			Link: "https://www.typescriptlang.org/",
		},
		"Rust": {
			Icon: "content/icons/rust.svg",
			Link: "https://www.rust-lang.org/",
		},
		"JavaScript": {
			Icon: "content/icons/javascript.svg",
			Link: "https://developer.mozilla.org/en-US/docs/Web/JavaScript",
		},
		"C#": {
			Icon: "content/icons/csharp.svg",
			Link: "https://dotnet.microsoft.com/en-us/languages/csharp",
		},
		"Java": {
			Icon: "content/icons/java.svg",
			Link: "https://www.java.com/en/download/help/whatis_java.html",
		},
		"Kotlin": {
			Icon: "content/icons/kotlin.svg",
			Link: "https://kotlinlang.org/",
		},
		"Python": {
			Icon: "content/icons/python.svg",
			Link: "https://www.python.org/",
		},
		"React": {
			Icon: "content/icons/react.svg",
			Link: "https://reactjs.org/",
		},
		"Next.js": {
			Icon: "content/icons/next-js.svg",
			Link: "https://nextjs.org/",
		},
		"Django": {
			Icon: "content/icons/django.svg",
			Link: "https://www.djangoproject.com/",
		},
		"Godot": {
			Icon: "content/icons/godot.svg",
			Link: "https://godotengine.org/",
		},
		"Unity": {
			Icon: "content/icons/unity.svg",
			Link: "https://unity.com/",
		},
		"libGDX": {
			Icon: "content/icons/libgdx.svg",
			Link: "https://libgdx.com/",
		},
		"gRPC": {
			Icon: "content/icons/grpc.svg",
			Link: "https://grpc.io/",
		},
		"GraphQL": {
			Icon: "content/icons/graphql.svg",
			Link: "https://graphql.org/",
		},
		"WebRTC": {
			Icon: "content/icons/webrtc.svg",
			Link: "https://webrtc.org/",
		},
		"MQTT": {
			Icon: "content/icons/mqtt.svg",
			Link: "https://mqtt.org/",
		},
		"ClickHouse": {
			Icon: "content/icons/clickhouse.svg",
			Link: "https://clickhouse.com/docs/en/intro",
		},
		"Elasticsearch": {
			Icon: "content/icons/elasticsearch.svg",
			Link: "https://www.elastic.co/guide/en/elasticsearch/reference/current/elasticsearch-intro.html",
		},
		"AWS CDK": {
			Icon: "content/icons/aws.svg",
			Link: "https://aws.amazon.com/cdk/",
		},
	}
)
