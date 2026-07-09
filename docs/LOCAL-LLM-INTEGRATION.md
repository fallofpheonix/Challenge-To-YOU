# Local LLM Integration — Technical Specification

## Overview

The **Local LLM** is a highly specialized, quantized model that runs on player machines to:
1. Analyze player code/style
2. Generate dynamic passcodes
3. Detect creative solutions
4. Provide contextual hints

**Goal**: Sub-100ms inference, <3GB RAM, works on consumer hardware.

---

## Model Selection

### Recommended Base Models

| Model | Parameters | Size (Q4) | Speed | Recommendation |
|-------|------------|-----------|-------|----------------|
| **Qwen-2.5-1.5B** | 1.5B | ~1GB | Blazing | **Best for MVP** |
| **Qwen-2.5-3B** | 3B | ~2GB | Fast | Good balance |
| **Llama-3-8B** | 8B | ~4GB | Slower | Overkill for MVP |
| **Phi-3-mini** | 3.8B | ~2GB | Fast | Alternative |

**MVP Choice**: `Qwen-2.5-1.5B` — Smallest, fastest, smartest for its size.

---

## Optimization Pipeline

### Step 1: Dataset Creation

```json
[
  {
    "instruction": "Analyze player code and generate passcode.",
    "input": {
      "code": "RUNE fire = IGNITE(power: 100)\nRUNE ice = FLOW(direction: NORTH)\nEFFECT { fire.COMBINE(ice) }",
      "theme": "magitech",
      "mode": "architect",
      "glitches": ["rune_interference"]
    },
    "output": {
      "passcode": "a1b2c3d4e5f6g7h8",
      "style": "creative_combiner",
      "complexity": 0.7,
      "hints": ["Opposite elements create unexpected reactions"]
    }
  },
  {
    "instruction": "Detect emergent solution from player actions.",
    "input": {
      "actions": ["connect_ice_rune_to_fire_totem", "activate_near_fireplace"],
      "theme": "magitech",
      "expected_glitch": "thermal_explosion"
    },
    "output": {
      "glitch_detected": true,
      "glitch_type": "thermal_shock",
      "creativity_score": 0.85,
      "passcode": "x9y8z7w6v5u4t3s2"
    }
  }
]
```

**Dataset Size**: 100-500 examples for MVP

---

### Step 2: Fine-Tuning with Unsloth

```python
# fine_tune.py
from unsloth import FastLanguageModel
from trl import SFTTrainer
from transformers import TrainingArguments

# 1. Load base model
model, tokenizer = FastLanguageModel.from_pretrained(
    model_name="unsloth/Qwen2.5-1.5B-Instruct",
    max_seq_length=2048,
    dtype=None,  # Auto-detect
    load_in_4bit=True,
)

# 2. Add LoRA adapters
model = FastLanguageModel.get_peft_model(
    model,
    r=16,  # LoRA rank
    target_modules=["q_proj", "k_proj", "v_proj", "o_proj",
                     "gate_proj", "up_proj", "down_proj"],
    lora_alpha=16,
    lora_dropout=0,
    bias="none",
    use_gradient_checkpointing="unsloth",
)

# 3. Load dataset
dataset = load_dataset("json", data_files="game_logic_dataset.json")

# 4. Train
trainer = SFTTrainer(
    model=model,
    tokenizer=tokenizer,
    train_dataset=dataset["train"],
    max_seq_length=2048,
    args=TrainingArguments(
        per_device_train_batch_size=2,
        gradient_accumulation_steps=4,
        warmup_steps=5,
        max_steps=100,  # Quick training for MVP
        learning_rate=2e-4,
        fp16=not torch.cuda.is_bf16_supported(),
        bf16=torch.cuda.is_bf16_supported(),
        logging_steps=1,
        output_dir="outputs",
    ),
)

trainer.train()

# 5. Save model
model.save_pretrained("challenge-to-you-llm")
tokenizer.save_pretrained("challenge-to-you-llm")
```

**Training Time**: ~30-60 minutes on free Colab GPU

---

### Step 3: Quantization to GGUF

```bash
# 1. Clone llama.cpp
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp

# 2. Convert to GGUF
python convert_hf_to_gguf.py ../challenge-to-you-llm \
    --outfile challenge-to-you.gguf \
    --outtype f16

# 3. Quantize to Q4_K_M (4-bit)
./llama-quantize challenge-to-you.gguf \
    challenge-to-you-Q4_K_M.gguf \
    Q4_K_M

# Result: ~1GB file, runs on CPU
```

**Output**: `challenge-to-you-Q4_K_M.gguf` (~1GB)

---

### Step 4: Local Runtime Setup

#### Option A: Ollama (Easiest)

```bash
# Create Modelfile
cat > Modelfile << 'EOF'
FROM ./challenge-to-you-Q4_K_M.gguf

PARAMETER temperature 0.3
PARAMETER top_p 0.9
PARAMETER num_ctx 512
PARAMETER num_thread 4

SYSTEM "You are a game logic analyzer. Output ONLY valid JSON. No explanations."
EOF

# Create Ollama model
ollama create challenge-to-you -f Modelfile

# Run
ollama run challenge-to-you
```

#### Option B: llama.cpp Server

```bash
# Start server
./llama-server \
    -m challenge-to-you-Q4_K_M.gguf \
    --host 127.0.0.1 \
    --port 8080 \
    -c 512 \
    -t 4 \
    --grammar-file json.gbnf
```

---

## Go Backend Integration

### Ollama Client

```go
package llm

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"
)

type OllamaClient struct {
    baseURL    string
    model      string
    httpClient *http.Client
}

type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Options  Options   `json:"options"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Options struct {
    Temperature float64 `json:"temperature"`
    TopP        float64 `json:"top_p"`
    NumCtx      int     `json:"num_ctx"`
    NumThread   int     `json:"num_thread"`
}

type ChatResponse struct {
    Message struct {
        Content string `json:"content"`
    } `json:"message"`
    Done bool `json:"done"`
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
    return &OllamaClient{
        baseURL: baseURL,
        model:   model,
        httpClient: &http.Client{
            Timeout: 5 * time.Second,
        },
    }
}

func (c *OllamaClient) AnalyzeCode(code, theme, mode string) (*AnalysisResult, error) {
    prompt := fmt.Sprintf(`Analyze this code and return ONLY valid JSON:
Theme: %s
Mode: %s
Code:
%s

Return JSON with: passcode, style, complexity, glitches`, theme, mode, code)
    
    req := ChatRequest{
        Model: c.model,
        Messages: []Message{
            {Role: "system", Content: "Output ONLY valid JSON. No explanations."},
            {Role: "user", Content: prompt},
        },
        Options: Options{
            Temperature: 0.3,
            TopP:        0.9,
            NumCtx:      512,
            NumThread:   4,
        },
    }
    
    resp, err := c.httpClient.Post(
        c.baseURL+"/api/chat",
        "application/json",
        bytes.NewBuffer(marshal(req)),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var chatResp ChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
        return nil, err
    }
    
    var result AnalysisResult
    if err := json.Unmarshal([]byte(chatResp.Message.Content), &result); err != nil {
        return nil, err
    }
    
    return &result, nil
}
```

---

### AST Fallback (No LLM Required)

```go
package analyzer

import (
    "go/ast"
    "go/parser"
    "go/token"
)

type ASTAnalyzer struct{}

type AnalysisResult struct {
    Complexity  float64  `json:"complexity"`
    Patterns    []string `json:"patterns"`
    Glitches    []string `json:"glitches"`
    Style       string   `json:"style"`
    Readability float64  `json:"readability"`
}

func (a *ASTAnalyzer) Analyze(code string) *AnalysisResult {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, "", code, 0)
    if err != nil {
        return &AnalysisResult{Complexity: 0.5}
    }
    
    result := &AnalysisResult{
        Patterns:    []string{},
        Glitches:    []string{},
        Style:       "unknown",
        Readability: 0.5,
    }
    
    // Count AST nodes for complexity
    ast.Inspect(node, func(n ast.Node) bool {
        if n != nil {
            result.Complexity += 0.1
        }
        return true
    })
    
    // Detect patterns
    if result.Complexity > 5 {
        result.Patterns = append(result.Patterns, "complex_logic")
    }
    
    // Normalize complexity
    if result.Complexity > 1 {
        result.Complexity = 1
    }
    
    return result
}
```

---

### Hybrid Analyzer (LLM + AST)

```go
package analyzer

type HybridAnalyzer struct {
    llmClient *llm.OllamaClient
    astAnalyzer *ASTAnalyzer
    llmAvailable bool
}

func NewHybridAnalyzer(llmURL, model string) *HybridAnalyzer {
    client := llm.NewOllamaClient(llmURL, model)
    
    // Check if LLM is available
    llmAvailable := client.Ping() == nil
    
    return &HybridAnalyzer{
        llmClient:   client,
        astAnalyzer: &ASTAnalyzer{},
        llmAvailable: llmAvailable,
    }
}

func (h *HybridAnalyzer) Analyze(code, theme, mode string) *AnalysisResult {
    // Always run AST analysis (fast, deterministic)
    astResult := h.astAnalyzer.Analyze(code)
    
    // If LLM available, enhance with AI analysis
    if h.llmAvailable {
        llmResult, err := h.llmClient.AnalyzeCode(code, theme, mode)
        if err == nil {
            return h.mergeResults(astResult, llmResult)
        }
    }
    
    // Fallback to AST only
    return astResult
}

func (h *HybridAnalyzer) mergeResults(ast *AnalysisResult, llm *AnalysisResult) *AnalysisResult {
    return &AnalysisResult{
        Complexity:  (ast.Complexity + llm.Complexity) / 2,
        Patterns:    append(ast.Patterns, llm.Patterns...),
        Glitches:    append(ast.Glitches, llm.Glitches...),
        Style:       llm.Style, // LLM is better at style detection
        Readability: (ast.Readability + llm.Readability) / 2,
    }
}
```

---

## JSON Grammar Constraint

### llama.cpp Grammar File (json.gbnf)

```gbnf
root   ::= object
value  ::= object | array | string | number | "true" | "false" | "null"
object ::= "{" ws (string ":" ws value ("," ws string ":" ws value)*)? "}"
array  ::= "[" ws (value ("," ws value)*)? "]"
string ::= "\"" char* "\""
char   ::= "\\"["\\/bfnrt] | "\\u" hex hex hex hex | [^"\\]
number ::= "-"? ("0" | [1-9] [0-9]*) ("." [0-9]+)? ([eE] [-+]? [0-9]+)?
ws     ::= ([ \t\n])*
hex    ::= [0-9a-fA-F]
```

**Effect**: Model can ONLY output valid JSON, no explanations.

---

## Performance Targets

### Latency Requirements

| Operation | Target | Method |
|-----------|--------|--------|
| Code Analysis | <100ms | LLM + AST |
| Passcode Generation | <50ms | Hash only |
| Glitch Detection | <200ms | Rule matrix |
| Hint Generation | <150ms | LLM (optional) |

### Resource Limits

| Resource | Limit | Enforcement |
|----------|-------|-------------|
| RAM | 3GB | OS monitoring |
| CPU | 4 threads | Thread pool |
| Context Window | 512 tokens | Hard limit |
| Max Response | 256 tokens | Truncation |

### Fallback Behavior

```
LLM Available?
├── Yes → Use LLM + AST hybrid
└── No → Use AST only (deterministic)
```

---

## Implementation Timeline

### Week 1
- [ ] Create dataset (100 examples)
- [ ] Set up Unsloth environment
- [ ] Fine-tune Qwen-2.5-1.5B

### Week 2
- [ ] Quantize to GGUF (Q4_K_M)
- [ ] Set up Ollama integration
- [ ] Implement Go client

### Week 3
- [ ] Build AST fallback
- [ ] Create hybrid analyzer
- [ ] Add JSON grammar constraint

### Week 4
- [ ] Optimize latency
- [ ] Test on low-end hardware
- [ ] Bundle with game installer

---

## Distribution Options

### Option A: Bundle Ollama (Recommended)
- Include Ollama installer in game setup
- Auto-download model on first launch
- ~1GB download

### Option B: Bundle Model Only
- Include GGUF file in game assets
- Use llama.cpp library directly
- ~1GB game size increase

### Option C: Download on First Launch
- Detect if Ollama installed
- If not, prompt to install
- Download model from CDN

---

*Last updated: 2026-07-10*
