// Package metadata holds build information injected at link time with -ldflags.
package metadata

import "fmt"

const (
	Name    = "fpm-golang"
	License = "BSD-3-Clause"
	Author  = "Andrea Funtò <dihedron.dev@gmail.com>"
)

// Set by the Makefile: -X github.com/dihedron/fpm-golang/pkg/metadata.Version=...
var (
	Version   = "dev"
	Commit    = "unknown"
	Date      = "unknown"
	GoVersion = "unknown" // toolchain that built the binary
	BuildHost = "unknown" // build environment (OS/arch)
)

// Short is the one-line form.
func Short() string { return fmt.Sprintf("%s %s", Name, Version) }

// Long is the verbose form.
func Long() string {
	return fmt.Sprintf("%s %s\nlicense:  %s\nauthor:   %s\ncommit:   %s\nbuilt:    %s\ngo:       %s\nplatform: %s\n",
		Name, Version, License, Author, Commit, Date, GoVersion, BuildHost)
}
