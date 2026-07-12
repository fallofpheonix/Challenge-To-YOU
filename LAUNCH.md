# Challenge To YOU — Launch Guide

## Quick Start

### Prerequisites
- Go 1.25+
- Godot 4.x
- Ollama (optional — for AI Archon taunts)

### 1. Start the Backend

```bash
# Clone and enter the project
cd challenge-to-you

# Option A: Use the startup script
./start.sh

# Option B: Manual start (static challenge)
cd backend
CHALLENGE_PATH=challenges/magitech_01.json ./sandbox

# Option C: Run with procedural generation
cd backend
./sandbox
# Then connect from Godot with seed/luck/paradigm params
```

### 2. Open the Godot Client

```bash
cd client
# Open Godot 4 and load project.godot
# Or from command line:
godot4 project.godot
```

### 3. Configure and Play

In the terminal UI:
1. Enter a **Seed** (number or word — anything works)
2. Set **Luck** (0.0 = hard, 1.0 = generous flaws)
3. Select **Era** (Magitech / Cyberpunk / Cosmic)
4. Click **"Establish Rift"**

Type `help` for available commands.

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CHALLENGE_PATH` | `challenges/magitech_01.json` | Static challenge to load |
| `DB_PATH` | `challenge.db` | SQLite database path |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama API endpoint |
| `OLLAMA_MODEL` | `llama3` | Ollama model to use |

---

## Running Specific Challenges

### Static Challenge Mode
```bash
# Magitech challenges
CHALLENGE_PATH=challenges/magitech_tier1/magitech_01.json ./sandbox
CHALLENGE_PATH=challenges/magitech_tier1/magitech_04_golem.json ./sandbox

# Cyberpunk challenges  
CHALLENGE_PATH=challenges/cyberpunk_tier1/cyberpunk_01_autodoc.json ./sandbox

# Cosmic challenges
CHALLENGE_PATH=challenges/cosmic_tier1/cosmic_01_airlock.json ./sandbox
```

### Procedural Mode
Connect from Godot with seed + luck + paradigm parameters. The backend will generate an infinite variety of challenges using the Hydrator system.

---

## Command Reference (In-Game)

| Command | Description |
|---------|-------------|
| `help` | Show all commands |
| `clear` | Clear terminal |
| `profile` | Show reputation, luck, unlocked eras |
| `unlock <era>` | Spend reputation to unlock an era |
| `mending` | Trigger AI repair protocol (+30 entropy) |
| `submit_answer:<value>` | Submit answer for recognize challenges |
| `execute_script` | Submit code from the code panel |
| `<event_id>` | Trigger a challenge flaw/event |
| `<event_id>:<payload>` | Trigger with a payload |

---

## Distribution

Pre-built server binaries are in `dist/`:

| Platform | Binary |
|----------|--------|
| macOS (Intel) | `server_darwin_amd64` |
| macOS (Apple Silicon) | `server_darwin_arm64` |
| Linux (x64) | `server_linux_amd64` |
| Windows (x64) | `server_windows_amd64.exe` |

Copy the appropriate binary + the `challenges/` directory to distribute.

---

## Enabling AI Archon (Ollama)

1. Install Ollama: https://ollama.ai
2. Pull a model: `ollama pull llama3`
3. Start Ollama: `ollama serve`
4. Run the game — AI taunts will automatically activate

Without Ollama, the game runs fine — the AI Archon will be silent.

---

## Itch.io Alpha Release

```bash
# Package for Itch.io
cp dist/server_darwin_arm64 ./challenge-to-you-macos
cp -r backend/challenges ./challenges
# + Godot export (File → Export → macOS)
# Upload both to itch.io
```

---

## Troubleshooting

**Port already in use:**
```bash
lsof -i :8080
kill -9 <PID>
```

**WebSocket connection refused in Godot:**
- Ensure the backend is running first
- Check `ws://localhost:8080/rift` is accessible

**Database errors:**
```bash
rm challenge.db  # Reset the database
```

**Go build errors:**
```bash
cd backend && go mod tidy && go build ./...
```

---

*Last updated: 2026-07-11*
*Status: Alpha — Ready for Playtest*
