export GO15VENDOREXPERIMENT=1
export GO111MODULE=on
# Many Go tools take file globs or directories as arguments instead of packages.
COVERALLS_TOKEN=477d8d1f-b729-472f-b842-e0e4b03bc0c2
# The linting tools evolve with each Go version, so run them only on the latest
# stable release.
GO_VERSION := $(shell go version | cut -d " " -f 3)
GO_MINOR_VERSION := $(word 2,$(subst ., ,$(GO_VERSION)))
LINTABLE_MINOR_VERSIONS := 16
ifneq ($(filter $(LINTABLE_MINOR_VERSIONS),$(GO_MINOR_VERSION)),)
SHOULD_LINT := true
endif

.PHONY: all
all: lint examples test

.PHONY: dependencies
dependencies:
ifdef SHOULD_LINT
	@echo "Installing golint..."
	go get -v golang.org/x/lint/golint
else
	@echo "Not installing golint, since we don't expect to lint on" $(GO_VERSION)
endif
	@echo "Installing test dependencies..."
	go mod tidy

.PHONY: lint
lint:
ifdef SHOULD_LINT
	@rm -rf lint.log
	@echo "Installing test dependencies for vet..."
	@go test -i ./...
	@echo "Checking vet..."
	@go vet ./... 2>&1 | tee -a lint.log
	@echo "Checking lint..."
	@golint ./... 2>&1 | tee -a lint.log
	@[ ! -s lint.log ]
else
	@echo "Skipping linters on" $(GO_VERSION)
endif

.PHONY: test
test:
	@go test ./...

.PHONY: cover
cover:
	sh cover.sh

.PHONY: examples
examples:
	@go generate ./... && cd cmd/datax && go build && cd ../..

.PHONY: doc
doc:
	@godoc -http=:6080