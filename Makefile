VERSION=1.21.1
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
	@fpm -s tar -t deb --prefix /usr/local --name $(NAME) --version $(VERSION) --iteration $(RELEASE) --description "The Go Programming Language" --vendor $(VENDOR) --maintainer $(MAINTAINER) --license $(LICENSE) --directories /usr/local/go --url $(PRODUCER_URL) --deb-compression bzip2 go$(VERSION).linux-amd64.tar.gz

.phony: rpm
rpm: go$(VERSION).linux-amd64.tar.gz
	@fpm -s tar -t rpm --prefix /usr/local --name $(NAME) --version $(VERSION) --iteration $(RELEASE) --description "The Go Programming Language" --vendor $(VENDOR) --maintainer $(MAINTAINER) --license $(LICENSE) --directories /usr/local/go --url $(PRODUCER_URL) go$(VERSION).linux-amd64.tar.gz

.phony: clean
clean:
	@rm -rf $(NAME)_$(VERSION)-$(RELEASE)_amd64.deb
	@rm -rf $(NAME)_$(VERSION)-$(RELEASE)_amd64.rpm

.phony: reset
reset: clean
	@rm -rf go$(VERSION).linux-amd64.tar.gz 

.phony: help
help:
	@echo "make setup       - install FPM and other tools"
	@echo "make deb         - create a DEB package"
	@echo "make rpm         - create a RPM package"
	@echo "make clean       - remove the DEB or RPM file"
	@echo "make reset       - remove the downloaded archive"
	@echo "make install     - install the package"
	@echo "make remove      - remove the package"

# see http://linuxmafia.com/faq/Admin/release-files.html
.phony: setup
setup:
ifneq (,$(wildcard /etc/lsb-release))
	@echo "Setting up prerequisite tools for Ubuntu or Mint Linux"
	sudo apt-get update && sudo apt-get install ruby-dev build-essential && sudo gem install fpm
else ifneq (,$(wildcard /etc/debian_version)) 
	@echo "Setting up prerequisite tools for Debian Linux (TODO)"
else ifneq (,$(wildcard /etc/redhat-release)) 
	@echo "Setting up prerequisite tools for Red Hat Enterprise Linux"
	yum install -y wget ruby rpm-build && gem install fpm
else ifneq (,$(wildcard /etc/fedora-release))
	@echo "Setting up prerequisite tools for Fedora Linux (TODO)"
endif
