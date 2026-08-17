SHELL := /bin/bash
TARGETS = solrbulk
PKGNAME = solrbulk
VERSION = 0.4.12

SEMVER := $(shell echo $(VERSION) | sed 's/^v//')

.PHONY: all
all: test
	$(MAKE) $(TARGETS)

.PHONY: test
test:
	go test -v ./...

docs/solrbulk.1: docs/solrbulk.md
	md2man-roff docs/solrbulk.md > docs/solrbulk.1

.PHONY: clean
clean:
	go clean
	rm -f coverage.out
	rm -f $(TARGETS)
	rm -rf build
	rm -f $(PKGNAME)_*.deb
	rm -f $(PKGNAME)-*.rpm

.PHONY: cover
cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

solrbulk: cmd/solrbulk/solrbulk.go
	go build -ldflags "-s -w -X github.com/miku/solrbulk.Version=$(VERSION)" -o $@ $<

# Cross-compiled linux/amd64 binary used for packaging (deb/rpm), kept in a
# separate directory so it never overwrites the native dev binary above and
# the host platform (e.g. macOS arm64) does not leak into the package.
build/solrbulk: cmd/solrbulk/solrbulk.go
	@mkdir -p build
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X github.com/miku/solrbulk.Version=$(VERSION)" -o $@ $<

# nfpm-based packaging.
.PHONY: deb
deb: build/solrbulk docs/solrbulk.1
	SEMVER=$(SEMVER) GOARCH=amd64 nfpm package -p deb -f nfpm.yaml

.PHONY: rpm
rpm: build/solrbulk docs/solrbulk.1
	SEMVER=$(SEMVER) GOARCH=amd64 nfpm package -p rpm -f nfpm.yaml
