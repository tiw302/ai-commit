# Simple Makefile for ai-commit

BINARY_NAME=ai-commit
MAIN_PATH=./cmd/ai-commit

.PHONY: all build clean install test

all: build

build:
	@echo "Building binary..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	go clean

install: build
	@echo "Installing binary to /usr/local/bin..."
	sudo mv $(BINARY_NAME) /usr/local/bin/

test:
	@echo "Running tests..."
	go test ./...
