# Project Phoenix: Toolchain Integration Spec

This document details the configuration specifications for connecting local LLMs, vector search components, and static compilers to the pipeline.

---

## 1. Local Inference Integrations

### 1.1 Ollama API Configuration
- **Host**: `http://localhost:11434`
- **Request Protocol**: HTTP POST `/api/generate` or `/api/chat`
- **System Parameter Defaults**:
  ```json
  {
    "options": {
      "temperature": 0.0,
      "top_p": 0.9,
      "num_predict": 2048,
      "stop": ["```"]
    }
  }
  ```

---

## 2. Code Search & AST Parsing

To locate target variables and function boundaries, the Knowledge Manager interfaces with CLI search libraries:
- **ripgrep (`rg`)**: Used to rapidly search for type names or diagnostic keywords across the workspace directory structure.
  - *Command Example*: `rg "type AxiomaticFabric" --json`
- **tree-sitter**: Used to parse code structures into concrete syntax trees. Allows agents to extract surrounding function bodies without importing full, large files.

---

## 3. Git Hook & CI Triggers

The pipeline runs validation checks prior to commits:
- **Pre-Commit Hook (`.git/hooks/pre-commit`)**:
  ```bash
  #!/bin/bash
  echo "Project Phoenix: Running pre-commit validation checks..."
  go build ./... && go test ./... && golangci-lint run
  if [ $? -ne 0 ]; then
      echo "Validation failed! Starting self-repair pipeline..."
      # Trigger local repair script
      exit 1
  fi
  ```
