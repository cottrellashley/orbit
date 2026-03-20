BINARY  := orbit
CMD     := ./cmd/orbit

.DEFAULT_GOAL := help

.PHONY: all build web-build test test-short install clean fmt vet check help

## all:        build frontend + build Go binary + install to GOPATH/bin
all: web-build build install

## build:      compile the orbit binary (assumes web/dist exists)
build:
	go build -o $(BINARY) $(CMD)

## web-build:  install npm deps (if needed) and build React frontend
web-build:
	cd web && npm install --prefer-offline --no-audit && npx vite build

## test:       run all tests with verbose output
test:
	go test ./... -v

## test-short: run all tests
test-short:
	go test ./...

## install:    install orbit to GOPATH/bin
install:
	go install $(CMD)

## clean:      remove build artifacts
clean:
	rm -f $(BINARY)
	rm -rf web/dist/assets
	go clean -cache -testcache

## fmt:        format all Go source files
fmt:
	go fmt ./...

## vet:        run go vet on all packages
vet:
	go vet ./...

## check:      fmt + vet + test (run before committing)
check: fmt vet test

## help:       show available targets
help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /' | column -t -s ':'
