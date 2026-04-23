SHELL := /bin/bash
TARGETS = solrbulk
PKGNAME = solrbulk
VERSION = 0.4.9

SEMVER := $(shell echo $(VERSION) | sed 's/^v//')

solrbulk: cmd/solrbulk/solrbulk.go
	go build -o $@ $^

.PHONY: all
all: imports test
	go build

.PHONY: test
test:
	go test -v ./...

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
	rm -f $(PKGNAME)_*.deb
	rm -f $(PKGNAME)-*.rpm

.PHONY: cover
cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

# nfpm-based packaging.
.PHONY: deb
deb: $(TARGETS) docs/solrbulk.1
	SEMVER=$(SEMVER) GOARCH=amd64 nfpm package -p deb -f nfpm.yaml

.PHONY: rpm
rpm: $(TARGETS) docs/solrbulk.1
	SEMVER=$(SEMVER) GOARCH=amd64 nfpm package -p rpm -f nfpm.yaml
