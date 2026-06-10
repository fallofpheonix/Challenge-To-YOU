# Chrysalis Engine Layer - Development Orchestration
# Use this Makefile to run the Go core and Godot client.

.PHONY: run-core run-client run-all test-core build-core

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

# Build the Go core binary
build-core:
	cd $(CORE_DIR) && go build -o ../bin/chrysalis-core main.go

# Run the Godot client
run-client:
	$(GODOT_BIN) --path $(CLIENT_DIR)

# Run both (Godot will launch the Go core)
run-all:
	$(GODOT_BIN) --path $(CLIENT_DIR)
