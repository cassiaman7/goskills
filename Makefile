.PHONY: all build clean

# Binary names
BINARY_CLI=goskills-cli
BINARY_RUNNER=goskills

# Build directory
BUILD_DIR=.

all: build

build: cli runner

cli:
	go build -o $(BUILD_DIR)/$(BINARY_CLI) ./cmd/skill-cli

runner:
	go build -o $(BUILD_DIR)/$(BINARY_RUNNER) ./cmd/skill-runner

clean:
	rm -f $(BUILD_DIR)/$(BINARY_CLI)
	rm -f $(BUILD_DIR)/$(BINARY_RUNNER)

test:
	go test ./...
