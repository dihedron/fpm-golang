VERSION=1.22.2
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

.phony: clean
clean:
	@rm -rf *.deb *.rpm *.apk *.tar.gz

