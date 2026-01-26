package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/log"

	"hermannm.dev/personal-website/devserver"
	"hermannm.dev/personal-website/sitebuilder"
)

func main() {
	log.SetDefault(devlog.NewHandler(os.Stdout, &devlog.Options{Level: slog.LevelDebug}))

	ctx := context.Background()

	args := parseCommandLineArgs()
	if args.useDevServer {
		if err := devserver.ServeAndRebuildOnChange(
			ctx,
			contentPaths,
			inputCSSFileName,
			args.devServerPort,
		); err != nil {
			log.Error(ctx, err, "Dev server stopped")
			os.Exit(1)
		}
	} else {
		log.Info(ctx, "Building website...")
		if err := sitebuilder.RenderPages(
			ctx,
			contentPaths,
			commonData,
			icons,
			args.invokedByDevServer,
		); err != nil {
			log.Error(ctx, err, "")
			os.Exit(1)
		}
		if err := sitebuilder.FormatRenderedPages(ctx); err != nil {
			log.Error(ctx, err, "Failed to format rendered pages")
			os.Exit(1)
		}
		if err := sitebuilder.GenerateTailwindCSS(ctx, inputCSSFileName); err != nil {
			log.Error(ctx, err, "Failed to generate CSS for rendered pages")
			os.Exit(1)
		}
		log.Info(
			ctx,
			"Website built successfully!",
			"outputDirectory",
			"./"+sitebuilder.BaseOutputDir,
		)
	}
}

type commandLineArgs struct {
	useDevServer       bool
	devServerPort      string
	invokedByDevServer bool
}

func parseCommandLineArgs() commandLineArgs {
	var args commandLineArgs

	flag.BoolVar(
		&args.useDevServer,
		"dev",
		false,
		"Serve and rebuild the site every time content/templates/sitebuilder files change",
	)
	flag.StringVar(
		&args.devServerPort,
		"port",
		"8080",
		"The port to serve the website from when using -dev",
	)
	flag.BoolVar(
		&args.invokedByDevServer,
		"invoked-by-dev-server",
		false,
		"Internal flag: identifies if we are being invoked by the dev server",
	)

	flag.Parse()
	return args
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
		ProjectDirs: []string{"projects", "companies", "libraries-and-tools"},
		BasicPages:  []string{"404_page.md"},
	}

	inputCSSFileName = "styles.css"

	icons = sitebuilder.IconMap{
		"person": {
			Path: "content/icons/person.svg",
		},
		"map-marker": {
			Path: "content/icons/map-marker.svg",
		},
		"arrow-left": {
			Path: "content/icons/arrow-left.svg",
		},
		"arrow-right": {
			Path: "content/icons/arrow-right.svg",
		},
		"GitHub": {
			Path:         "content/icons/github.svg",
			IconForLinks: []string{"https://github.com"},
		},
		"LinkedIn": {
			Path: "content/icons/linkedin.svg",
		},
		"Gopher": {
			Path:         "content/icons/gopher.svg",
			IconForLinks: []string{"https://pkg.go.dev"},
		},
		"Go": {
			Path:                  "content/icons/go.svg",
			Link:                  "https://go.dev/",
			IndexPageFallbackPath: "content/icons/go-alt.svg",
		},
		"Rust": {
			Path:                  "content/icons/rust.svg",
			Link:                  "https://www.rust-lang.org/",
			IndexPageFallbackPath: "content/icons/rust-alt.svg",
			IconForLinks:          []string{"https://docs.rs"},
		},
		"Cargo": {
			Path:         "content/icons/cargo.svg",
			IconForLinks: []string{"https://crates.io"},
		},
		"Kotlin": {
			Path:         "content/icons/kotlin.svg",
			Link:         "https://kotlinlang.org/",
			IconForLinks: []string{"https://devlog-kotlin.hermannm.dev"},
		},
		"JetBrains": {
			Path:         "content/icons/jetbrains.svg",
			IconForLinks: []string{"https://klibs.io", "https://plugins.jetbrains.com"},
		},
		"Kotlin+Go+Rust": {
			IndexPageFallbackPath: "content/icons/kotlin-go-rust-combined.svg",
		},
		"TypeScript": {
			Path: "content/icons/typescript.svg",
			Link: "https://www.typescriptlang.org/",
		},
		"JavaScript": {
			Path: "content/icons/javascript.svg",
			Link: "https://developer.mozilla.org/en-US/docs/Web/JavaScript",
		},
		"C#": {
			Path: "content/icons/csharp.svg",
			Link: "https://dotnet.microsoft.com/en-us/languages/csharp",
		},
		"Java": {
			Path: "content/icons/java.svg",
			Link: "https://www.java.com/en/download/help/whatis_java.html",
		},
		"Python": {
			Path: "content/icons/python.svg",
			Link: "https://www.python.org/",
		},
		"React": {
			Path: "content/icons/react.svg",
			Link: "https://reactjs.org/",
		},
		"Next.js": {
			Path: "content/icons/next-js.svg",
			Link: "https://nextjs.org/",
		},
		"Django": {
			Path: "content/icons/django.svg",
			Link: "https://www.djangoproject.com/",
		},
		"PostgreSQL": {
			Path: "content/icons/postgres.svg",
			Link: "https://www.postgresql.org/",
		},
		"Godot": {
			Path: "content/icons/godot.svg",
			Link: "https://godotengine.org/",
		},
		"Unity": {
			Path: "content/icons/unity.svg",
			Link: "https://unity.com/",
		},
		"libGDX": {
			Path: "content/icons/libgdx.svg",
			Link: "https://libgdx.com/",
		},
		"gRPC": {
			Path: "content/icons/grpc.svg",
			Link: "https://grpc.io/",
		},
		"GraphQL": {
			Path: "content/icons/graphql.svg",
			Link: "https://graphql.org/",
		},
		"WebRTC": {
			Path: "content/icons/webrtc.svg",
			Link: "https://webrtc.org/",
		},
		"MQTT": {
			Path: "content/icons/mqtt.svg",
			Link: "https://mqtt.org/",
		},
		"ClickHouse": {
			Path: "content/icons/clickhouse.svg",
			Link: "https://clickhouse.com/docs/en/intro",
		},
		"Elasticsearch": {
			Path: "content/icons/elasticsearch.svg",
			Link: "https://www.elastic.co/guide/en/elasticsearch/reference/current/elasticsearch-intro.html",
		},
		"AWS": {
			Path: "content/icons/aws.svg",
			Link: "https://aws.amazon.com/cdk/",
		},
		"Azure": {
			Path: "content/icons/azure.svg",
			Link: "https://azure.microsoft.com/",
		},
		"VSCode": {
			Path:         "content/icons/vscode.svg",
			IconForLinks: []string{"https://marketplace.visualstudio.com"},
		},
		"NTNU": {
			Path:         "content/icons/ntnu.svg",
			IconForLinks: []string{"https://ntnuopen.ntnu.no"},
		},
	}
)
