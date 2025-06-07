---
name: condqueue
path: /condqueue
tagLine: A concurrent queue.
goPackage:
  rootName: hermannm.dev/condqueue
  githubURL: https://github.com/hermannm/condqueue
techStack:
  - tech: Go
links:
  - title: Code
    link: https://github.com/hermannm/condqueue
  - title: Docs
    link: https://pkg.go.dev/hermannm.dev/condqueue
---

A small Go package providing a concurrent queue, on which consumers can wait for an item satisfying
a given condition, and producers can add items to wake consumers.
