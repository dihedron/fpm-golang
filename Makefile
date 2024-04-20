VERSION=$(shell curl -s https://go.dev/VERSION?m=text | head -n 1 | sed -e "s/^go//")
DOWNLOAD_URL=https://go.dev/

go$(VERSION).linux-amd64.tar.gz:
	@wget $(DOWNLOAD_URL)dl/go$(VERSION).linux-amd64.tar.gz

.phony: deb
deb: go$(VERSION).linux-amd64.tar.gz 
ifeq ($(GITLAB_CI),)
ifeq ($(shell which nfpm),)
	@echo "Need to install nFPM first..."
	@go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
endif
endif
	@rm -rf go/ 
	@tar xzvf go$(VERSION).linux-amd64.tar.gz 2>&1 > /dev/null
	@echo -n "Create golang $(VERSION) "
	@VERSION=$(VERSION) nfpm package --packager deb --target .
	@rm -rf go/

.phony: rpm
rpm: go$(VERSION).linux-amd64.tar.gz 
ifeq ($(GITLAB_CI),)
ifeq ($(shell which nfpm),)
	@echo "Need to install nFPM first..."
	@go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
endif
endif
	@rm -rf go/ 
	@tar xzvf go$(VERSION).linux-amd64.tar.gz 2>&1 > /dev/null
	@echo -n "Create golang $(VERSION) "
	@VERSION=$(VERSION) nfpm package --packager rpm --target .
	@rm -rf go/

# TODO: run a cleanup task removing go/ only once:
# see https://gist.github.com/APTy/9a9eb218f68bc0b4beb133b89c9def14

.phony: apk
apk: go$(VERSION).linux-amd64.tar.gz 
ifeq ($(GITLAB_CI),)
ifeq ($(shell which nfpm),)
	@echo "Need to install nFPM first..."
	@go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
endif
endif
	@rm -rf go/ 
	@tar xzvf go$(VERSION).linux-amd64.tar.gz 2>&1 > /dev/null
	@echo -n "Create golang $(VERSION) "
	@VERSION=$(VERSION) nfpm package --packager apk --target .
	@rm -rf go/

.phony: docker-go
docker-go: go$(VERSION).linux-amd64.tar.gz 
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg UBUNTU_VERSION=latest -t golang -f Dockerfile-go .

.phony: docker-gcc
docker-gcc: go$(VERSION).linux-amd64.tar.gz 
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg UBUNTU_VERSION=latest -t golang-gcc -f Dockerfile-gcc .

.phony: docker-clang
docker-clang: go$(VERSION).linux-amd64.tar.gz 
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg UBUNTU_VERSION=latest -t golang-clang -f Dockerfile-clang .

.phony: clean
clean:
	@rm -rf *.deb *.rpm *.apk *.tar.gz

