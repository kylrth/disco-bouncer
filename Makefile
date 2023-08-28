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

.PHONY: docker
docker:
	docker build -t kylrth/disco-bouncer:latest .
