---
name: devlog
path: /devlog
tagLine: Structured logging.
goPackage:
  rootName: hermannm.dev/devlog
  githubURL: https://github.com/hermannm/devlog
techStackTitle: Implemented in
techStack:
  - tech: Kotlin
  - tech: Go
  - tech: Rust
links:
  - title: Kotlin version
    sublinks:
      - title: Code
        link: https://github.com/hermannm/devlog-kotlin
      - title: Docs
        link: https://devlog-kotlin.hermannm.dev
      - title: Published on
        link: https://klibs.io/project/hermannm/devlog-kotlin
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
---

_`devlog`_ is the name of a set of logging libraries that I've built for different programming
languages. They all focus on developer-friendly structured logging, but they differ in their scope:

- The Go and Rust implementations provide an alternate log output format for the standard structured
  logging libraries, designed for use in local development.
- The Kotlin implementation is a more comprehensive library, providing a full-fledged logger API.

The Go implementation was the first version of _`devlog`_. After working with Go in multiple
projects ([`casus-belli`](/casus-belli), [`analysis`](/analysis), [`coffeetalk`](/coffeetalk),
[Ignite](/ignite)), one of the things I missed was a nicer human-readable log output format. So when
Go added structured logging to its standard library, I took the opportunity to write my own log
handler! With help from an
[amazing guide](https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md)
written by one of the Go maintainers, I built a structured log handler with an output format
designed for readability in local development. I now use this in all my Go projects where I need
logging.

Later, I started writing some Rust (see [`gadd`](/gadd)), and there too I found myself missing nicer
log output formats. So I decided to write my own log subscriber for _`tracing`_, one of the most
popular logging libraries for Rust, to use the same output format as my Go library. And so,
<span class="whitespace-nowrap">_`devlog-tracing`_</span> was born.

Finally, after starting my job at [Liflig](/liflig), I started writing Kotlin for the backend. I
found myself unhappy with the logging library we were using at the time, and so I decided to write
yet another version of _`devlog`_. My aim with the library is to provide an ergonomic logging API
that makes it as easy as possible to attach structured data to logs, while keeping the abstractions
near-zero-cost at runtime. The implementation wraps the standard Java logging libraries _SLF4J_ and
_Logback_, so that it can interoperate with logs from other libraries. We now use this library in
many of our backend services at Liflig, and I'm quite happy with it! It's quite satisfying to make a
library that directly addresses frustrations you've been having.
