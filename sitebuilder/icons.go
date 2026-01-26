package sitebuilder

import (
	"fmt"
	"html/template"
	"os"

	"golang.org/x/sync/errgroup"
	"hermannm.dev/wrap"
)

type IconMap map[string]*IconConfig

type IconConfig struct {
	Path string `validate:"required,filepath"`
	// Populated after [PageRenderer.RenderIcons] finishes.
	RenderedIcon          template.HTML
	Link                  string `validate:"omitempty,url"`
	IndexPageFallbackPath string `validate:"omitempty,filepath"`
	// Blank if there was no index page fallback icon for this entry.
	RenderedIndexPageFallbackIcon template.HTML
	// Base URL of links that this icon should be used for.
	IconForLinks []string `validate:"omitempty,dive,url"`
}

func (icons IconMap) getRenderedIcon(name string) (template.HTML, error) {
	icon, ok := icons[name]
	if !ok {
		return "", fmt.Errorf("failed to find expected icon '%s' in icon map", name)
	}
	if icon.RenderedIcon == "" {
		return "", fmt.Errorf("icon '%s' was not rendered", name)
	}
	return icon.RenderedIcon, nil
}

func (renderer *PageRenderer) RenderIcons() error {
	var group errgroup.Group

	for _, icon := range renderer.icons {
		// Combined icons, such as "Go+Rust", only define IndexPageFallbackPath
		if icon.Path != "" {
			group.Go(
				func() error {
					return readIconFile(icon.Path, &icon.RenderedIcon)
				},
			)
		}

		if icon.IndexPageFallbackPath != "" {
			group.Go(
				func() error {
					return readIconFile(
						icon.IndexPageFallbackPath,
						&icon.RenderedIndexPageFallbackIcon,
					)
				},
			)
		}
	}

	if err := group.Wait(); err != nil {
		return err
	}

	githubIcon, err := renderer.icons.getRenderedIcon("GitHub")
	if err != nil {
		return err
	}
	renderer.commonData.githubIcon = githubIcon

	// Signals to other goroutines that icons have finished rendering
	close(renderer.iconsRendered)
	return nil
}

func readIconFile(path string, out *template.HTML) error {
	svgBytes, err := os.ReadFile(path)
	if err != nil {
		return wrap.Errorf(err, "failed to read svg file for icon at path '%s'", path)
	}

	*out = template.HTML(svgBytes)
	return nil
}
