package devserver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"hermannm.dev/devlog/errlog"
	"hermannm.dev/errclose"
	"hermannm.dev/wrap"

	"hermannm.dev/personal-website/sitebuilder"
)

func ServeAndRebuildOnChange(
	ctx context.Context,
	contentPaths sitebuilder.ContentPaths,
	cssFileName string,
	port string,
) (returnedErr error) {
	buildSite := func() {
		err := sitebuilder.ExecCommand(
			ctx,
			true,
			"go",
			"run",
			"hermannm.dev/personal-website",
			"-invoked-by-dev-server",
		)
		// We only log exec errors here, as actual build errors will be printed by the command
		if err != nil && !strings.HasPrefix(err.Error(), "go failed") {
			slog.ErrorContext(ctx, "", errlog.Cause(err))
		}
	}

	buildSite()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return wrap.Error(err, "failed to create file system watcher")
	}
	defer errclose.Close(watcher, &returnedErr, "file system watcher")

	go func() {
		var lastEvent fsnotify.Event
		var timeOfLastBuild time.Time

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return // Watcher closed
				}

				if event == lastEvent && time.Since(timeOfLastBuild) < 100*time.Millisecond {
					continue
				}

				buildSite()

				lastEvent = event
				timeOfLastBuild = time.Now()
			case err, ok := <-watcher.Errors:
				if !ok {
					return // Watcher closed
				}

				slog.ErrorContext(ctx, "File system watcher error", errlog.Cause(err))
			}
		}
	}()

	dirsToWatch := []string{
		"main.go",
		"sitebuilder",
		cssFileName,
		sitebuilder.PageTemplatesDir,
		sitebuilder.ComponentTemplatesDir,
		sitebuilder.BaseContentDir,
		sitebuilder.BaseContentDir + "/icons",
	}
	for _, projectDir := range contentPaths.ProjectDirs {
		dirsToWatch = append(
			dirsToWatch,
			fmt.Sprintf("%s/%s", sitebuilder.BaseContentDir, projectDir),
		)
	}

	for _, dir := range dirsToWatch {
		if err := watcher.Add(dir); err != nil {
			return wrap.Errorf(err, "failed to add '%s' to file system watcher", dir)
		}
	}

	slog.InfoContext(ctx, "Serving website...", "port", port)
	return sitebuilder.ExecCommand(
		ctx,
		false,
		"npx",
		"live-server",
		sitebuilder.BaseOutputDir,
		"--port="+port,
	)
}
