---
name: gadd
slug: gadd
tagLine: Command-line utility for Git.
iconPath: /img/projects/ferris-the-git-crab.png
iconAlt: Ferris the Crab, mascot of the Rust programming language, with the Git logo on its forehead
techStack:
  - tech: Rust
linkGroups:
  - title: Code
    links:
      - text: hermannm/gadd
        link: https://github.com/hermannm/gadd
footnote:
  Git logo adapted from [Jason Long](https://git-scm.com/downloads/logos) (licensed under [CC BY
  3.0](https://creativecommons.org/licenses/by/3.0/))
---

Ever since I learned Git, I've used it in all my development projects. Even on solo projects, I like
to commit to Git, as it's such an ingrained part of my workflow.

I generally prefer using Git from the terminal rather than a GUI. I feel it gives me more control,
since Git was built for the terminal. However, one thing annoyed me about Git's terminal experience:
staging individual files. _git-add_ has an
[interactive mode](https://git-scm.com/docs/git-add#_interactive_mode) that aims to alleviate this,
but I found its interface clunky. I looked around for alternatives, but those I found either did too
much or had other problems. Thus, I decided to create _gadd_: a small command-line utility for
staging files to Git.

I wanted to program more in Rust after
[using it for Advent of Code](https://github.com/hermannm/advent-of-rust), and found it suitable for
a terminal application like this. To interact with Git, I used Rust bindings for
[libgit2](https://libgit2.org/). This taught me a lot about how Git works under the hood ⁠— as is
often the case, it is more complex than it looks on the surface!

Now I use _gadd_ almost daily, and quite enjoy it. There's something extra satisfying about creating
your own tool and tailoring it exactly to your needs.
