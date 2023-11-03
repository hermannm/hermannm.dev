---
page:
  goPackage:
    fullName: hermannm.dev/condqueue
    githubURL: https://github.com/hermannm/condqueue
name: condqueue
slug: condqueue
tagLine: A concurrent queue.
techStack:
  - tech: Go
linkGroups:
  - title: Code
    links:
      - text: hermannm/condqueue
        link: https://github.com/hermannm/condqueue
  - title: Docs
    links:
      - text: pkg.go.dev/hermannm.dev/condqueue
        link: https://pkg.go.dev/hermannm.dev/condqueue
        iconPath: /img/icons/gopher.svg
---

A small Go package providing a concurrent queue, on which consumers can wait for an item satisfying
a given condition, and producers can add items to wake consumers.
