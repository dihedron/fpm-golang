NAME=golang-sddc
VERSION=1.18.2
MAINTAINER=maintainer@example.com
LICENSE="BSD 3 Clause"
RELEASE=1
URL=https://go.dev/

go$(VERSION).linux-amd64.tar.gz:
	@wget $(URL)dl/go$(VERSION).linux-amd64.tar.gz

.phony: deb
deb: go$(VERSION).linux-amd64.tar.gz
	@fpm -s tar -t deb --prefix /usr/local --name $(NAME) --version $(VERSION) --iteration $(RELEASE) --description "The Go Programming Language" --maintainer $(MAINTAINER) --license $(LICENSE) --directories /usr/local/go --url $(URL) --deb-compression bzip2 go$(VERSION).linux-amd64.tar.gz

.phony: rpm
rpm: go$(VERSION).linux-amd64.tar.gz
	@fpm -s tar -t rpm --prefix /usr/local --name $(NAME) --version $(VERSION) --iteration $(RELEASE) --description "The Go Programming Language" --maintainer $(MAINTAINER) --license $(LICENSE) --directories /usr/local/go --url $(URL) go$(VERSION).linux-amd64.tar.gz

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
	@echo "Setting up prerequisite tools for Debian Linux"
else ifneq (,$(wildcard /etc/redhat-release)) 
	@echo "Setting up prerequisite tools for Red Hat Enterprise Linux"
else ifneq (,$(wildcard /etc/fedora-release))
	@echo "Setting up prerequisite tools for Fedora Linux"
endif
