BINARY  := orbit
CMD     := ./cmd/orbit
INSTALL := $(GOPATH)/bin/$(BINARY)

.PHONY: build test install clean fmt vet lint check

## build: compile the orbit binary
build:
	go build -o $(BINARY) $(CMD)

## test: run all tests
test:
	go test ./... -v

## test-short: run tests without verbose output
test-short:
	go test ./...

## install: install orbit to GOPATH/bin
install:
	go install $(CMD)

## clean: remove build artifacts
clean:
	rm -f $(BINARY)
	go clean -cache -testcache

## fmt: format all Go source files
fmt:
	go fmt ./...

## vet: run go vet on all packages
vet:
	go vet ./...

## check: fmt + vet + test (use before committing)
check: fmt vet test

## help: show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
