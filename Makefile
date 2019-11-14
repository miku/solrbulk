SHELL := /bin/bash
TARGETS = solrbulk

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go get -d && go test -v

imports:
	goimports -w .

all: imports test
	go build

clean:
	go clean
	rm -f coverage.out
	rm -f $(TARGETS)
	rm -f solrbulk-*.x86_64.rpm
	rm -f solrbulk*.deb
	rm -rf debian/solrbulk/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

solrbulk: cmd/solrbulk/solrbulk.go
	go build cmd/solrbulk/solrbulk.go

deb: $(TARGETS)
	mkdir -p debian/solrbulk/usr/sbin
	cp $(TARGETS) debian/solrbulk/usr/sbin
	mkdir -p debian/solrbulk/usr/local/share/man/man1
	cp docs/solrbulk.1 debian/solrbulk/usr/local/share/man/man1
	cd debian && fakeroot dpkg-deb --build solrbulk .
	mv debian/solrbulk_*.deb .

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/solrbulk.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	cp docs/solrbulk.1 $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh solrbulk
	cp $(HOME)/rpmbuild/RPMS/x86_64/solrbulk*.rpm .
