# Deployment & Distribution Guide

This document details instructions for packaging, configuring, and launching *Challenge-To-YOU* in standalone playtest environments.

---

## 1. Sandbox Server Executables

The server can be compiled using the cross-platform packaging script:
```bash
./tools/build.sh
```
This generates binaries in the `dist/` directory:
- `server_darwin_amd64` (macOS Intel)
- `server_darwin_arm64` (macOS Apple Silicon)
- `server_linux_amd64` (Linux Desktop/Server)
- `server_windows_amd64.exe` (Windows Executable)

### Distribution Requirements
To launch the server, the executable requires the `challenges/` directory to exist in the same relative path (e.g. `./challenges/...`) to load initial configuration matrices.

---

## 2. Server Environment Variables

Customize server execution using environment flags on launch:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Listening port for websocket traffic. | `8080` |
| `DB_PATH` | File path for SQLite meta-progression database. | `challenge.db` |
| `CHALLENGE_PATH` | Path to load the default static starting challenge. | `challenges/magitech_01.json` |
| `OLLAMA_URL` | Endpoint of the running local Ollama API server. | `http://localhost:11434` |
| `OLLAMA_MODEL` | Quantized local LLM identifier. | `qwen2.5:1.5b-instruct` |

---

## 3. Godot Client Export

To export the client into a standalone binary:
1. Open the project inside Godot 4.x.
2. Navigate to **Project -> Export**.
3. Add a target preset (**Windows Desktop**, **macOS**, or **Linux Desktop**).
4. Download the Godot export templates if prompted.
5. Click **Export Project** to compile the game shell.

Alternatively, use Godot's CLI export tools:
```bash
# Example Linux headless export
godot --headless --export-release "Linux/X11" dist/ChallengeToYou.x86_64
```

---

## 4. Ollama Local Model Setup

The Archon's narrative voice is powered by a quantized local model (default: `qwen2.5:1.5b-instruct` for fast inference).

1. Download and install [Ollama](https://ollama.com/).
2. Pull the instruction-tuned model:
   ```bash
   ollama pull qwen2.5:1.5b-instruct
   ```
3. Start the Ollama daemon (runs on `http://localhost:11434` by default). The sandbox server will connect automatically.
