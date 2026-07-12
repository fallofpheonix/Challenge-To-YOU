#!/usr/bin/env bash
# Challenge To YOU — Development Startup Script
# Starts the Go backend server for local development

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/backend"

# Defaults
CHALLENGE_PATH="${CHALLENGE_PATH:-challenges/magitech_tier1/magitech_01.json}"
DB_PATH="${DB_PATH:-challenge.db}"
PORT="${PORT:-8080}"
OLLAMA_URL="${OLLAMA_URL:-http://localhost:11434}"
OLLAMA_MODEL="${OLLAMA_MODEL:-llama3}"

echo "╔══════════════════════════════════════════════════════╗"
echo "║         CHALLENGE TO YOU — AXIOMATIC FABRIC          ║"
echo "╚══════════════════════════════════════════════════════╝"
echo ""
echo "  Challenge Path : $CHALLENGE_PATH"
echo "  Database       : $DB_PATH"
echo "  Port           : $PORT"
echo "  Ollama URL     : $OLLAMA_URL"
echo "  Ollama Model   : $OLLAMA_MODEL"
echo ""

# Build if sandbox binary doesn't exist
if [ ! -f "$BACKEND_DIR/sandbox" ]; then
    echo "Building backend..."
    cd "$BACKEND_DIR"
    go build -o sandbox ./cmd/sandbox/
    echo "Build complete."
fi

cd "$BACKEND_DIR"

echo "Starting Challenge Engine on ws://localhost:$PORT/rift"
echo "Open Godot client to begin."
echo ""

CHALLENGE_PATH="$CHALLENGE_PATH" \
  DB_PATH="$DB_PATH" \
  OLLAMA_URL="$OLLAMA_URL" \
  OLLAMA_MODEL="$OLLAMA_MODEL" \
  ./sandbox
