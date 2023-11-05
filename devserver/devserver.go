package devserver

import (
	"fmt"
	"log/slog"
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
		err := sitebuilder.BuildSite(contentPaths, commonData, icons, cssFileName)
		if err == nil {
			log.Info(
				"website built successfully!",
				slog.String("outputDirectory", "./"+sitebuilder.BaseOutputDir),
			)
		} else {
			log.Error(err, "")
		}
	}

	log.Info("building website...")
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

				log.Info("rebuilding website...", slog.String("trigger", event.Name))
				buildSite()

				lastEvent = event
				timeOfLastBuild = time.Now()
			case err, ok := <-watcher.Errors:
				if !ok {
					return // Watcher closed
				}

				log.Error(err, "file system watcher error")
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
		"live-server",
		"npx",
		"live-server",
		sitebuilder.BaseOutputDir,
		"--port="+port,
	)
}
