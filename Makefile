## simple makefile to log workflow
.PHONY: all test clean proto build install

GOFLAGS ?= $(GOFLAGS:)

all: install test

fmt:
	@go fmt $(GOFLAGS) ./...

proto: head/head.proto
	protoc --go_out=. --python_out=. head/head.proto

build: proto
	@go build $(GOFLAGS) ./...

install:
	@go get $(GOFLAGS) ./...

test: install
	@go test $(GOFLAGS) ./...

bench: install
	@go test -run=NONE -bench=. $(GOFLAGS) ./...

clean:
	@go clean $(GOFLAGS) -i ./...

