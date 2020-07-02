SHELL := /bin/bash
TARGETS = solrbulk

solrbulk: cmd/solrbulk/solrbulk.go
	go build cmd/solrbulk/solrbulk.go

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
	mkdir -p debian/solrbulk/usr/sbin
	cp $(TARGETS) debian/solrbulk/usr/sbin
	mkdir -p debian/solrbulk/usr/local/share/man/man1
	cp docs/solrbulk.1 debian/solrbulk/usr/local/share/man/man1
	cd debian && fakeroot dpkg-deb --build solrbulk .
	mv debian/solrbulk_*.deb .

.PHONY: rpm
rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/solrbulk.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	cp docs/solrbulk.1 $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh solrbulk
	cp $(HOME)/rpmbuild/RPMS/x86_64/solrbulk*.rpm .
