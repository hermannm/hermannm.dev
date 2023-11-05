# hermannm.dev

hermannm's personal website and Go package host. Built with Go's
[html/template](https://pkg.go.dev/html/template), the [goldmark](https://github.com/yuin/goldmark)
Markdown parser and [Tailwind CSS](https://tailwindcss.com/).

## Development setup

1. Install [Go](https://go.dev/) and [Node.js](https://nodejs.org/en)
2. Run `npm ci` to install NPM dependencies
3. Run `go run .` to build the site once
4. Run `go run . -dev` to serve and rebuild the site every time content/templates/sitebuilder files
   change
