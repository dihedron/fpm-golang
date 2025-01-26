# fpm-golang

A simple Makefile to create `.deb`, `.rpm` and `.apk` packages of the upstream Google Golang compiler.

## Building a [deb|rpm] package

In order to build a specific version of the Golang compiler package for Ubuntu or Debian based Linux distributions, run the Makefile as follows:

```bash
$> make deb
```

To build an RPM package, run as follows:

```bash
$> make rpm
```

To create an APK package (for Alpine) run:

```bash
$> make apk
```

The makefile will automatically download the `tar.gz` package from https://go.dev and repackage it.

To clean all packages and downloaded files run `make clean`.

## Prerequisites

In order to create DEB, RPM and APK packages, this project uses [nFPM](https://nfpm.goreleaser.com/); if not available locally, it uses `go install` to install it, so both `make` and `go` must already be available on the packaging machine if you don't want to install nFPM manually.
