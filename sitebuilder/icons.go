package sitebuilder

import (
	"errors"
	"html/template"
	"os"

	"golang.org/x/sync/errgroup"
	"hermannm.dev/wrap"
)

type IconMap map[string]*IconConfig

type IconConfig struct {
	Icon                  string `validate:"required,filepath"`
	Link                  string `validate:"omitempty,url"`
	IndexPageFallbackIcon string `validate:"omitempty,filepath"`
	// Base URL of links that this icon should be used for.
	IconForLinks []string `validate:"omitempty,dive,url"`
}

func (renderer *PageRenderer) RenderIcons() error {
	var group errgroup.Group

	for _, icon := range renderer.icons {
		// Combined icons, such as "Go+Rust", only define IndexPageFallbackIcon
		if icon.Icon != "" {
			group.Go(
				func() error {
					return replaceIconWithSVG(&icon.Icon)
				},
			)
		}

		if icon.IndexPageFallbackIcon != "" {
			group.Go(
				func() error {
					return replaceIconWithSVG(&icon.IndexPageFallbackIcon)
				},
			)
		}
	}

	if err := group.Wait(); err != nil {
		return err
	}

	githubIcon, ok := renderer.icons["GitHub"]
	if !ok {
		return errors.New("expected icon map to have entry for 'GitHub'")
	}
	renderer.commonData.githubIcon = template.HTML(githubIcon.Icon)

	// Signals to other goroutines that icons have finished rendering
	close(renderer.iconsRendered)
	return nil
}

func replaceIconWithSVG(icon *string) error {
	svgBytes, err := os.ReadFile(*icon)
	if err != nil {
		return wrap.Errorf(err, "failed to read svg file for icon '%s'", *icon)
	}

	*icon = string(svgBytes)
	return nil
}
