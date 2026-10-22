---
name: omnissiah
path: /omnissiah
tagLine: AI agent sandboxing.
goPackage:
  rootName: hermannm.dev/omnissiah
  githubURL: https://github.com/hermannm/omnissiah
logo:
  path: /img/logos/omnissiah.png
  altText: A hooded robot head
techStack:
  - tech: Go
links:
  - title: Code
    link: https://github.com/hermannm/omnissiah
footnote:
  Credit to [Robert "TheChirurgeon" Jones](https://www.tabletopbattles.com/author/thechirurgeon/)
  for the logo
---

The Omnissiah is a thin, opinionated wrapper around
[Docker Sandboxes](https://docs.docker.com/ai/sandboxes/) for sandboxing AI agents.

I built this tool in 2026, after I started working for the startup [Birdsview](/birdsview). I wanted
to make sure that no AI agent I used could do anything I didn't want it to, and the solution for
that is strict sandboxing, following the
[principle of least privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege).

I could have just used Docker Sandboxes directly, but I decided to create my own wrapper in order
to:

- Set stricter defaults
- Add utilities to streamline my workflow
- Have fun with sci-fi references in command names

And so, The Omnissiah was born!
