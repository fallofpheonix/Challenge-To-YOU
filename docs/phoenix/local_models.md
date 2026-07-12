# Project Phoenix: Local Model Configurations

This document details the recommended local model allocations, system hardware limits, and context routing parameters.

---

## 1. Local Model Matrix

| Pipeline Role | Target Model | Context Window | System RAM / VRAM |
| :--- | :--- | :--- | :--- |
| **Planning** | Qwen2.5-32B-Instruct | 32k tokens | 32 GB RAM / 16 GB VRAM |
| **Coding & Repair** | DeepSeek-Coder-V2-Lite | 64k tokens | 64 GB RAM / 24 GB VRAM |
| **Fast Fixes & Lint** | Qwen2.5-Coder-7B | 16k tokens | 16 GB RAM / 8 GB VRAM |
| **Review & Security** | Qwen2.5-Coder-32B | 32k tokens | 32 GB RAM / 16 GB VRAM |
| **Embeddings** | mxbai-embed-large | 512 tokens | < 4 GB VRAM |

---

## 2. Model Routing & Load Balancing

To optimize local inference times on a single workstation:
- **Task Gating**: Simple syntax and formatting corrections bypass the 32B model path, routing directly to the lightweight 7B model.
- **Concurrent Execution Limit**: The pipeline locks inference requests to one active execution thread, preventing CPU and VRAM bottlenecks.
- **Context Pruning**: The Knowledge Manager restricts code file injections to the immediate function coordinates and imports, keeping prompt lengths under 8,000 tokens for standard runs.
