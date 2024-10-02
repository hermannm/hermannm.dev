---
page:
  goPackage:
    rootName: hermannm.dev/devlog
    githubURL: https://github.com/hermannm/devlog
name: devlog
slug: devlog
tagLine: A structured log handler.
techStack:
  - tech: Go
  - tech: Rust
linkGroups:
  - title: Go version
    links:
      - text: hermannm/devlog
        link: https://github.com/hermannm/devlog
  - title: Rust version
    links:
      - text: hermannm/devlog-tracing
        link: https://github.com/hermannm/devlog-tracing
---

After working with Go in multiple projects ([casus-belli](/casus-belli), [analysis](/analysis),
[Ignite](/ignite), [coffeetalk](/coffeetalk)), one of the things I missed was a nicer human-readable
log output format. So when Go added structured logging to its standard library, I took the
opportunity to write my own log handler! With help from an
[amazing guide](https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md)
written by one of the Go maintainers, I created _devlog_, a structured log handler with an output
format designed for readability in local development. I now use this in all my Go projects where I
need logging.

Later, I started writing more and more Rust (see [gadd](/gadd)), and there too I found myself
missing nicer log output formats. So I decided to write my own log subscriber for _tracing_, one of
the most popular logging libraries for Rust, to use the same output format as my Go library. And so,
_devlog-tracing_ was born.
