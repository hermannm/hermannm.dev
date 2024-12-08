---
name: devlog
path: /devlog
tagLine: Structured log handling.
goPackage:
  rootName: hermannm.dev/devlog
  githubURL: https://github.com/hermannm/devlog
techStack:
  - tech: Go
  - tech: Rust
  - tech: Kotlin
linkGroups:
  - title: Go version
    links:
      - text: hermannm/devlog
        link: https://github.com/hermannm/devlog
  - title: Rust version
    links:
      - text: hermannm/devlog-tracing
        link: https://github.com/hermannm/devlog-tracing
  - title: Kotlin version
    links:
      - text: hermannm/devlog-kotlin
        link: https://github.com/hermannm/devlog-kotlin
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
yet another version of _`devlog`_. The implementation is a thin wrapper over the standard Java
logging libraries _SLF4J_ and _Logback_, but with a more ergonomic API for Kotlin. It also provides
a log output encoder for local development, using the same format as my Go and Rust implementations.
