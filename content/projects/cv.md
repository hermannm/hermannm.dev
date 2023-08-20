---
page:
  goPackage:
    fullName: hermannm.dev/cv
    githubURL: https://github.com/hermannm/cv
name: cv
slug: cv
techStack:
  - tech: Go
linkGroups:
  - title: Code
    links:
      - text: hermannm/cv
        link: https://github.com/hermannm/cv
---

Dynamic CV and job application builder, rendering Markdown/YAML content into HTML templates.

Written in Go, using:

- [goldmark](https://github.com/yuin/goldmark) for Markdown parsing
- [go-yaml/yaml](https://github.com/go-yaml/yaml) for YAML parsing
- [html/template](https://pkg.go.dev/html/template) for HTML rendering
- [Tailwind CSS](https://tailwindcss.com/) for styling
