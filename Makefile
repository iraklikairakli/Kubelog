# Makefile for the kubelog project in the root directory

# Binary output name
BINARY_NAME=kubelog

# Path to the source code
SOURCE_DIR=cmd/kubelog

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Build the project
all: test build

# Build binary for the current platform
build:
	cd $(SOURCE_DIR) && $(GOBUILD) -o ../../$(BINARY_NAME) -v

# Build binary for Linux platform
build-linux:
	cd $(SOURCE_DIR) && GOOS=linux GOARCH=amd64 $(GOBUILD) -o ../../$(BINARY_NAME) -v

# Run tests
test:
	$(GOTEST) -v ./...

# Clean up binaries
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Fetch dependencies
deps:
	$(GOGET) -v ./...

# Run the program
run:
	cd $(SOURCE_DIR) && $(GOBUILD) -o ../../$(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

# Install the program
install:
	@echo "Installing $(BINARY_NAME) to /usr/local/bin"
	@sudo mv kubelog /usr/local/bin/
	@sudo chmod +x /usr/local/bin/kubelog
