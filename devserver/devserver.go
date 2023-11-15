package devserver

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"hermannm.dev/devlog/log"
	"hermannm.dev/personal-website/sitebuilder"
	"hermannm.dev/wrap"
)

func ServeAndRebuildOnChange(
	contentPaths sitebuilder.ContentPaths,
	commonData sitebuilder.CommonPageData,
	icons sitebuilder.IconMap,
	cssFileName string,
	port string,
) error {
	buildSite := func() {
		err := sitebuilder.ExecCommand(true, "go", "run", "hermannm.dev/personal-website")
		// We only want to log command setup errors, as actual build errors will be logged by the
		// program
		if err != nil && !strings.HasPrefix(err.Error(), "go failed") {
			log.Error(err)
		}
	}

	buildSite()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return wrap.Error(err, "failed to create file system watcher")
	}
	defer watcher.Close()

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

				log.ErrorCause(err, "file system watcher error")
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

	log.Info("serving website...", slog.String("port", port))
	return sitebuilder.ExecCommand(
		false,
		"npx",
		"live-server",
		sitebuilder.BaseOutputDir,
		"--port="+port,
	)
}
