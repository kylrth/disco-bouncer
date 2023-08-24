all: lint build

BINDIR := bin

## LINT

LINTER_VERSION := 1.53.3
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
