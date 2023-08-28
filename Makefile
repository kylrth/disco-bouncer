all: lint build

BINDIR := bin

## LINT

LINTER_VERSION := 1.54.2
LINTER := $(BINDIR)/golangci-lint_$(LINTER_VERSION)
DEV_OS := $(shell uname -s | tr A-Z a-z)

$(LINTER):
	mkdir -p $(BINDIR)
	wget "https://github.com/golangci/golangci-lint/releases/download/v$(LINTER_VERSION)/golangci-lint-$(LINTER_VERSION)-$(DEV_OS)-amd64.tar.gz" -O - \
		| tar -xz -C $(BINDIR) --strip-components=1 --exclude=README.md --exclude=LICENSE
	mv $(BINDIR)/golangci-lint $(LINTER)

.PHONY: lint
lint: $(LINTER)
	$(LINTER) run --deadline=2m

## TEST

.PHONY: test
test:
	go test -race -cover ./...

## BUILD

GO_FILES = $(shell find . -type f -name '*.go')

$(BINDIR)/bouncer: $(GO_FILES)
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/bouncer ./cmd

$(BINDIR)/client: $(GO_FILES)
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/client ./cmd/client

.PHONY: build
build: $(BINDIR)/bouncer $(BINDIR)/client

## CLIENT RELEASE BUILDS

RELEASEDIR := $(BINDIR)/release
ARCHES := amd64 arm64

$(RELEASEDIR)/bouncer-client-linux-%: $(GO_FILES)
	mkdir -p $(RELEASEDIR)
	GOOS=linux GOARCH=$* go build -o $(RELEASEDIR)/bouncer-client-linux-$* ./cmd

$(RELEASEDIR)/bouncer-client-darwin-%: $(GO_FILES)
	mkdir -p $(RELEASEDIR)
	GOOS=darwin GOARCH=$* go build -o $(RELEASEDIR)/bouncer-client-darwin-$* ./cmd

$(RELEASEDIR)/bouncer-client-windows-%.exe: $(GO_FILES)
	mkdir -p $(RELEASEDIR)
	GOOS=windows GOARCH=$* go build -o $(RELEASEDIR)/bouncer-client-windows-$*.exe ./cmd

.PHONY: release
release: $(foreach arch,$(ARCHES),$(RELEASEDIR)/bouncer-client-linux-$(arch) $(RELEASEDIR)/bouncer-client-darwin-$(arch) $(RELEASEDIR)/bouncer-client-windows-$(arch).exe)

## DOCKER

.PHONY: docker
docker:
	docker build -t ghcr.io/kylrth/disco-bouncer:latest .
