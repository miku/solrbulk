SHELL := /bin/bash
TARGETS = solrbulk

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go get -d && go test -v

imports:
	goimports -w .

fmt:
	go fmt ./...

all: fmt test
	go build

install:
	go install

clean:
	go clean
	rm -f coverage.out
	rm -f $(TARGETS)
	rm -f solrbulk-*.x86_64.rpm
	rm -f debian/solrbulk*.deb
	rm -rf debian/solrbulk/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

solrbulk:
	go build cmd/solrbulk/solrbulk.go

# ==== packaging

deb: $(TARGETS)
	mkdir -p debian/solrbulk/usr/sbin
	cp $(TARGETS) debian/solrbulk/usr/sbin
	cd debian && fakeroot dpkg-deb --build solrbulk .

REPOPATH = /usr/share/nginx/html/repo/CentOS/6/x86_64

publish: rpm
	cp solrbulk-*.rpm $(REPOPATH)
	createrepo $(REPOPATH)

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/solrbulk.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh solrbulk
	cp $(HOME)/rpmbuild/RPMS/x86_64/solrbulk*.rpm .
