# Makefile

# Project variables
PROJECT_NAME := synapse
MAIN_PACKAGE := ./cmd/synapse
BUILD_DIR := .
RELEASE_DIR := $(BUILD_DIR)/synapse
ZIP_NAME := synapse.zip
CONFIG_DIR := $(RELEASE_DIR)/conf

# Add any Go build flags or linker flags (LDFLAGS) here
LDFLAGS := "-s -w"
DEBUG_FLAGS := "-gcflags=all=-N -l"

.PHONY: all deps build package clean test

## Default target: build + package
all: deps build package

deps:
	@echo "Ensuring Go module dependencies..."
	go mod tidy

## Build for the host OS/architecture
build:
	@echo "Building $(PROJECT_NAME) for the host OS..."
	go build -ldflags=$(LDFLAGS) -o bin/$(PROJECT_NAME) $(MAIN_PACKAGE)

## Build with debug information
build-debug:
	@echo "Building $(PROJECT_NAME) with debug information..."
	go build $(DEBUG_FLAGS) -o bin/$(PROJECT_NAME) $(MAIN_PACKAGE)

test:
	@echo "Running tests..."
	go test -v ./...	

## Package the binary and folder structure into Synapse.zip
package: build test

	@echo "Cleaning up..."
	rm -rf $(RELEASE_DIR)

	@echo "Packaging into $(ZIP_NAME)..."
	# 1. Create necessary directories
	mkdir -p $(RELEASE_DIR)/bin
	mkdir -p $(CONFIG_DIR) # Create the config directory  
	mkdir -p $(RELEASE_DIR)/artifacts/APIs
	mkdir -p $(RELEASE_DIR)/artifacts/Endpoints
	mkdir -p $(RELEASE_DIR)/artifacts/Sequences
	mkdir -p $(RELEASE_DIR)/artifacts/Inbounds

	# 2. Copy the binary
	cp bin/$(PROJECT_NAME) $(RELEASE_DIR)/bin/

	# 3. Create the ZIP file
	cd $(BUILD_DIR) && zip -r $(ZIP_NAME) synapse

	@echo "Package created at $(BUILD_DIR)/$(ZIP_NAME)"

	# 4. Clean up
	rm -rf $(RELEASE_DIR)
	rm -rf bin

## Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(RELEASE_DIR)