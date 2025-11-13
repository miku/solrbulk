SHELL := /bin/bash
TARGETS = solrbulk
PKGNAME = solrbulk
VERSION = 0.4.2

solrbulk: cmd/solrbulk/solrbulk.go
	go build -o $@ $^

.PHONY: all
all: imports test
	go build

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
.PHONY: test
test:
	go get -d && go test -v

.PHONY: imports
imports:
	goimports -w .

docs/solrbulk.1: docs/solrbulk.md
	md2man-roff docs/solrbulk.md > docs/solrbulk.1

.PHONY: clean
clean:
	go clean
	rm -f coverage.out
	rm -f $(TARGETS)
	rm -f solrbulk-*.x86_64.rpm
	rm -f solrbulk*.deb
	rm -rf debian/solrbulk/usr

.PHONY: cover
cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out


.PHONY: deb
deb: $(TARGETS)
	mkdir -p debian/$(PKGNAME)/usr/local/bin
	cp $(TARGETS) debian/$(PKGNAME)/usr/local/bin
	mkdir -p debian/$(PKGNAME)/usr/local/share/man/man1
	cp docs/$(PKGNAME).1 debian/$(PKGNAME)/usr/local/share/man/man1
	cd debian && fakeroot dpkg-deb -Zzstd --build $(PKGNAME) .
	mv debian/$(PKGNAME)_*.deb .

.PHONY: rpm
rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	mkdir -p $(HOME)/rpmbuild/SOURCES/$(PKGNAME)
	cp ./packaging/$(PKGNAME).spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/SOURCES/$(PKGNAME)
	cp docs/$(PKGNAME).1 $(HOME)/rpmbuild/SOURCES/$(PKGNAME)
	./packaging/buildrpm.sh $(PKGNAME)
	cp $(HOME)/rpmbuild/RPMS/x86_64/$(PKGNAME)-$(VERSION)*.rpm .

