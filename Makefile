.PHONY: all build clean agent cli runner test

# Binary names
BINARY_CLI=goskills-cli
BINARY_RUNNER=goskills
BINARY_AGENT=agent-cli

# Build directory
BUILD_DIR=.

all: build

build: cli runner agent

agent:
	go build -o $(BUILD_DIR)/$(BINARY_AGENT) ./cmd/agent-cli

cli:
	go build -o $(BUILD_DIR)/$(BINARY_CLI) ./cmd/skill-cli

runner:
	go build -o $(BUILD_DIR)/$(BINARY_RUNNER) ./cmd/skill-runner

clean:
	rm -f $(BUILD_DIR)/$(BINARY_CLI)
	rm -f $(BUILD_DIR)/$(BINARY_RUNNER)
	rm -f $(BUILD_DIR)/$(BINARY_AGENT)

test:
	go test ./...