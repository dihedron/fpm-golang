# Version of the Go release to package; defaults to the latest stable one.
VERSION?=$(shell curl -fsS https://go.dev/VERSION?m=text | head -n 1 | sed -e "s/^go//")
# Target architecture: amd64 or arm64.
ARCH?=amd64
# Packaging revision: 1 is the plain upstream version, >= 2 adds a -N suffix.
REVISION?=1
# Expected tarball digest; if empty it is taken from the upstream feed (latest release only).
SHA256?=
# Pinned packager version.
NFPM_VERSION=v2.49.0
DOWNLOAD_URL=https://go.dev/

ifeq ($(ARCH),amd64)
ARCH_NAME=x86-64
else ifeq ($(ARCH),arm64)
ARCH_NAME=arm64
else
$(error unsupported ARCH '$(ARCH)': use amd64 or arm64)
endif

TARBALL=go$(VERSION).linux-$(ARCH).tar.gz
ifeq ($(shell test $(REVISION) -gt 1 && echo y),y)
RELEASE=$(REVISION)
SUFFIX=-$(REVISION)
endif
PACKAGE=google-golang-$(ARCH_NAME)-$(VERSION)$(SUFFIX)
CLI=bin/fpm-golang
NFPM=bin/nfpm
LDFLAGS=-X github.com/dihedron/fpm-golang/pkg/metadata.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev) \
	-X github.com/dihedron/fpm-golang/pkg/metadata.Commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown) \
	-X github.com/dihedron/fpm-golang/pkg/metadata.Date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) \
	-X github.com/dihedron/fpm-golang/pkg/metadata.GoVersion=$(shell go env GOVERSION) \
	-X github.com/dihedron/fpm-golang/pkg/metadata.BuildHost=$(shell go env GOOS)/$(shell go env GOARCH)

.PHONY: all
all: build

.PHONY: help
help: ## list the available targets and variables
	@echo "Targets: build test latest deb rpm apk packages setup-tools install-tools clean docker-*"
	@echo "Variables: VERSION (default: latest) ARCH (amd64|arm64) REVISION (default 1) SHA256"
	@echo "Signing: NFPM_SIGN_KEY (GPG key file) + NFPM_PASSPHRASE for deb and rpm, NFPM_APK_SIGN_KEY (RSA PEM) for apk; leave unset for unsigned"

.PHONY: build
build: $(CLI)

$(CLI): $(shell find cmd internal pkg -name '*.go') go.mod
	@go build -ldflags "$(LDFLAGS)" -o $(CLI) ./cmd/fpm-golang

.PHONY: test
test:
	@go test ./...

.PHONY: latest
latest:
	@echo "Latest golang version is $(VERSION)"

$(TARBALL): | $(CLI)
	@curl -fsSL -o $(TARBALL) $(DOWNLOAD_URL)dl/$(TARBALL)
	@$(CLI) release verify --tarball $(TARBALL) $(if $(SHA256),--sha256 $(SHA256)) || { rm -f $(TARBALL); exit 1; }

# nFPM is installed next to our binary at the pinned version (older ones emit broken apks).
$(NFPM): Makefile
	@GOBIN=$(CURDIR)/bin go install github.com/goreleaser/nfpm/v2/cmd/nfpm@$(NFPM_VERSION)

.PHONY: setup-tools
setup-tools: $(NFPM)

# Package the tarball with nFPM; the three formats share one recipe.
# Export NFPM_SIGN_KEY (path) and NFPM_PASSPHRASE to sign; leave unset for an unsigned build.
.PHONY: deb rpm apk
deb rpm apk: $(TARBALL) $(NFPM)
	@rm -rf go/
	@umask 0022 && tar xzf $(TARBALL)
	@echo "Create golang $(VERSION) $@ for $(ARCH)"
	@VERSION=$(VERSION) ARCH=$(ARCH) RELEASE=$(RELEASE) $(NFPM) package --packager $@ --target $(PACKAGE).$@
	@rm -rf go/

.PHONY: packages
packages: deb rpm apk

.PHONY: install-tools
install-tools:
	@go install github.com/josharian/impl@latest
	@go install github.com/haya14busa/goplay/cmd/goplay@latest
	@go install github.com/go-delve/delve/cmd/dlv@latest
	@go install golang.org/x/tools/gopls@latest

.PHONY: docker-go
docker-go: $(TARBALL)
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg UBUNTU_VERSION=latest -t golang -f Dockerfile-go .

.PHONY: docker-ubuntu-gcc
docker-ubuntu-gcc: $(TARBALL)
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg UBUNTU_VERSION=latest -t golang-ubuntu-gcc -f Dockerfile-ubuntu-gcc .

.PHONY: docker-ubuntu-clang
docker-ubuntu-clang: $(TARBALL)
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg UBUNTU_VERSION=latest -t golang-ubuntu-clang -f Dockerfile-ubuntu-clang .

.PHONY: docker-alma-gcc
docker-alma-gcc: $(TARBALL) 
	@docker build --progress=plain --no-cache --network=bridge --build-arg GOLANG_VERSION=$(VERSION) --build-arg ALMA_VERSION=8.9 -t golang-alma-gcc -f Dockerfile-alma-gcc .

.PHONY: docker-alma-clang
docker-alma-clang: $(TARBALL) 
	@docker build --progress=plain --no-cache --network=bridge --build-arg GOLANG_VERSION=$(VERSION) --build-arg ALMA_VERSION=8.9 -t golang-alma-clang -f Dockerfile-alma-clang .

.PHONY: clean
clean:
	@rm -rf *.deb *.rpm *.apk *.tar.gz
