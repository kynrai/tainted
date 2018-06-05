BINARIES=$$(go list ./...)
TESTABLE=$$(go list ./...)

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
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o linux.amd64.tainted -ldflags="-s -w" . && \
	tar czf linux.amd64.tainted.tar.gz linux.amd64.tainted && \
	rm linux.amd64.tainted
