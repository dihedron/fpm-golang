VERSION=1.18.1
MAINTAINER=sddc@bancaditalia.it
LICENSE="BSD 3 Clause"
ITERATION=1
URL=https://go.dev/

go$(VERSION).linux-amd64.tar.gz:
	@wget https://go.dev/dl/go$(VERSION).linux-amd64.tar.gz

.phony: deb
deb: go$(VERSION).linux-amd64.tar.gz
	@fpm -s tar -t deb --prefix /usr/local --name golang --version $(VERSION) --iteration $(ITERATION) --description "The Go Programming Language" --maintainer $(MAINTAINER) --license $(LICENSE) --directories /usr/local/go --url $(URL) --deb-compression bzip2 go$(VERSION).linux-amd64.tar.gz

.phony: rpm
rpm: go$(VERSION).linux-amd64.tar.gz
	@fpm -s tar -t rpm --prefix /usr/local --name golang --version $(VERSION) --iteration $(ITERATION) --description "The Go Programming Language" --maintainer $(MAINTAINER) --license $(LICENSE) --directories /usr/local/go --url $(URL) go$(VERSION).linux-amd64.tar.gz

.phony: clean
clean:
	@rm -rf golang_$(VERSION)-$(ITERATION)_amd64.deb

.phony: reset
reset: clean
	@rm -rf go$(VERSION).linux-amd64.tar.gz 

# see http://linuxmafia.com/faq/Admin/release-files.html
.phony: setup-tools
setup-tools:
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
