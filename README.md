# golang-fpm

A simple Makefile to create `.deb` and `.rpm` packages of the Golang compiler.

## Installing prerequiste tools

In order to install all pre-requisites and tools, run

```bash
$> make setup-tools
```

It automatically detects whether it's running on Ubuntu, Mint, Debian, Fedora or Red Hat Enterprise Linux and sets up tools accordingly.

## Building a [deb|rpm] package

In order to build a specific version of the Golang compiler package for Ubuntu or Debian based Linux distributions, run the Makefile as follows:

```bash
$> make deb
```

To build an RPM package, run as follows:

```bash
$> make rpm
```

The makefile will automatically download the `tar.gz` package from https://go.dev and repackage it.

To clean all packages, run `make clean`, and to remove downloaded files run `make reset`.



