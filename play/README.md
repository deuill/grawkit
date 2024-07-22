# The Grawkit Playground

This folder contains an interactive version of Grawkit that can be run in a web-browser (Javascript
support optional), built entirely in Go. Once built, this depends only on the contents of the
`static` directory, as well as the `grawkit` script itself for operation.

Execution of Grawkit is done with [`goawk`](https://github.com/benhoyt/goawk), a POSIX-compatible
AWK implementation built entirely in Go, and used here as a library.

## Installation & Usage

Only a fairly recent version of Go is required to build this package, and the included `main.go`
file can be executed on-the-fly using `go run .`.

The program expects to find a `static` directory (which is provided alongside the source code
here), as well as the `grawkit` script; by default, these are expected to be found in the current
and parent directories, respectively. Alternatively, their locations may be set using command-line
arguments, run `play -help` for more.

By default, this will bind an HTTP server on port `8080`, though this can also be modified using
the command-line arguments. Run `play -help` for more.
