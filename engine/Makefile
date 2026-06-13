# Chrysalis Engine Layer - Development Orchestration
# Use this Makefile to run the Go core and Godot client.

.PHONY: run-core run-client run-all test-core test-race check build-core

# Paths
GODOT_BIN = /Applications/Godot.app/Contents/MacOS/Godot
CORE_DIR = ./core
CLIENT_DIR = ./client

# Run the Go core directly
run-core:
	cd $(CORE_DIR) && go run main.go

# Run Go unit tests
test-core:
	cd $(CORE_DIR) && go test ./...

test-race:
	cd $(CORE_DIR) && go test -race ./...

check: test-core
	cd $(CORE_DIR) && go vet ./...
	$(GODOT_BIN) --headless --path $(CLIENT_DIR) --editor --quit

# Build the Go core binary
build-core:
	cd $(CORE_DIR) && go build -o ../bin/chrysalis-core main.go

# Run the Godot client
run-client: build-core
	$(GODOT_BIN) --path $(CLIENT_DIR)

# Run both (Godot launches the built Go core)
run-all: build-core
	$(GODOT_BIN) --path $(CLIENT_DIR)
