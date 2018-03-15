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

package:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o linux.amd64.chooser -ldflags="-s -w" ./cmd/chooser
