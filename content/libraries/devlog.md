---
name: devlog
path: /devlog
tagLine: Structured logging.
goPackage:
  rootName: hermannm.dev/devlog
  githubURL: https://github.com/hermannm/devlog
techStackTitle: Implemented in
techStack:
  - tech: Go
  - tech: Rust
  - tech: Kotlin
links:
  - title: Go version
    sublinks:
      - title: Code
        link: https://github.com/hermannm/devlog
      - title: Docs
        link: https://pkg.go.dev/hermannm.dev/devlog
  - title: Rust version
    sublinks:
      - title: Code
        link: https://github.com/hermannm/devlog-tracing
      - title: Docs
        link: https://docs.rs/devlog-tracing
      - title: Published on
        link: https://crates.io/crates/devlog-tracing
  - title: Kotlin version
    sublinks:
      - title: Code
        link: https://github.com/hermannm/devlog-kotlin
      - title: Docs
        link: https://devlog-kotlin.hermannm.dev
      - title: Published on
        link: https://klibs.io/project/hermannm/devlog-kotlin
---

After working with Go in multiple projects ([`casus-belli`](/casus-belli), [`analysis`](/analysis),
[`coffeetalk`](/coffeetalk), [Ignite](/ignite)), one of the things I missed was a nicer
human-readable log output format. So when Go added structured logging to its standard library, I
took the opportunity to write my own log handler! With help from an
[amazing guide](https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md)
written by one of the Go maintainers, I created _`devlog`_, a structured log handler with an output
format designed for readability in local development. I now use this in all my Go projects where I
need logging.

Later, I started writing more and more Rust (see [`gadd`](/gadd)), and there too I found myself
missing nicer log output formats. So I decided to write my own log subscriber for _`tracing`_, one
of the most popular logging libraries for Rust, to use the same output format as my Go library. And
so, <span class="whitespace-nowrap">_`devlog-tracing`_</span> was born.

Finally, after starting my job at [Liflig](/liflig), I started writing Kotlin for the backend. I
found myself unhappy with the logging library we were using at the time, and so I decided to write
yet another version of _`devlog`_. This implementation is a thin wrapper over the standard Java
logging libraries _SLF4J_ and _Logback_, but with a more ergonomic API for Kotlin.
