VERSION=1.22.2
NAME=golang-sddc
MAINTAINER=maintainer@example.com
VENDOR=vendor@example.com
LICENSE="BSD 3 Clause"
RELEASE=1
PRODUCER_URL=https://go.dev/
DOWNLOAD_URL=$(PRODUCER_URL)

go$(VERSION).linux-amd64.tar.gz:
	@wget $(DOWNLOAD_URL)dl/go$(VERSION).linux-amd64.tar.gz

.phony: deb
deb: go$(VERSION).linux-amd64.tar.gz 
ifeq (, $(shell which nfpm))
	@echo "Need to install nFPM first..."
	@go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
endif
	@rm -rf go/ 
	@tar xzvf go$(VERSION).linux-amd64.tar.gz 2>&1 > /dev/null
	@echo -n "Create golang $(VERSION) "
	@VERSION=$(VERSION) nfpm package --packager deb --target .
	@rm -rf go/

.phony: rpm
rpm: go$(VERSION).linux-amd64.tar.gz 
ifeq (, $(shell which nfpm))
	@echo "Need to install nFPM first..."
	@go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
endif
	@rm -rf go/ 
	@tar xzvf go$(VERSION).linux-amd64.tar.gz 2>&1 > /dev/null
	@echo -n "Create golang $(VERSION) "
	@VERSION=$(VERSION) nfpm package --packager rpm --target .
	@rm -rf go/

.phony: apk
apk: go$(VERSION).linux-amd64.tar.gz 
ifeq (, $(shell which nfpm))
	@echo "Need to install nFPM first..."
	@go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
endif
	@rm -rf go/ 
	@tar xzvf go$(VERSION).linux-amd64.tar.gz 2>&1 > /dev/null
	@echo -n "Create golang $(VERSION) "
	@VERSION=$(VERSION) nfpm package --packager apk --target .
	@rm -rf go/

.phony: clean
clean:
	@rm -rf $(NAME)_$(VERSION)-$(RELEASE)_amd64.deb
	@rm -rf $(NAME)_$(VERSION)-$(RELEASE)_amd64.rpm

.phony: reset
reset: clean
	@rm -rf go$(VERSION).linux-amd64.tar.gz 

