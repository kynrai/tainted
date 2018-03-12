BINARIES=$$(go list ./cmd/...)
TESTABLE=$$(go list ./... | grep -v /vendor/)

all : test build

deps:
	@dep ensure && dep ensure -update
.PHONY: deps

build:
	@go install -v  $(BINARIES)
.PHONY: build

test:
	@go test -v $(TESTABLE)
.PHONY: test
