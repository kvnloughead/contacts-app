// Package ui uses the built-in embed package to embed static files in our Go binary when its compiled.
//
// Usage notes for built-in embed package.
//
// The package level comment directive `go:embed "static"“ instructs Go to store the files (recursively) in ./static in an embedded filesystem referenced by global variable Files. This filesystem is rooted on the directed where the comment directive is placed.
//
// Additional directories and files can be specified in the same comment directive and * works as a wildcard.
//
// Files beginning with a . or _ are ignored unless the all: prefix is specified. For example: "all:static".
package ui

import "embed"

//go:embed "html" "static"
var Files embed.FS
