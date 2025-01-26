VERSION?=$(shell curl -s https://go.dev/VERSION?m=text | head -n 1 | sed -e "s/^go//")
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

.phony: docker-ubuntu-gcc
docker-ubuntu-gcc: go$(VERSION).linux-amd64.tar.gz 
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg UBUNTU_VERSION=latest -t golang-ubuntu-gcc -f Dockerfile-ubuntu-gcc .

.phony: docker-ubuntu-clang
docker-ubuntu-clang: go$(VERSION).linux-amd64.tar.gz 
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg UBUNTU_VERSION=latest -t golang-ubuntu-clang -f Dockerfile-ubuntu-clang .

.phony: docker-alma-gcc
docker-alma-gcc: go$(VERSION).linux-amd64.tar.gz 
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg ALMA_VERSION=8.9 -t golang-alma-gcc -f Dockerfile-alma-gcc .

.phony: docker-alma-clang
docker-alma-clang: go$(VERSION).linux-amd64.tar.gz 
	@docker build --progress=plain --no-cache --build-arg GOLANG_VERSION=$(VERSION) --build-arg ALMA_VERSION=8.9 -t golang-alma-clang -f Dockerfile-alma-clang .

.phony: clean
clean:
	@rm -rf *.deb *.rpm *.apk *.tar.gz

