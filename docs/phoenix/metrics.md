# Project Phoenix: Performance & Health Metrics

This document defines the Key Performance Indicators (KPIs) and repository health equations used to track the efficiency of the self-repair pipeline.

---

## 1. Key Performance Indicators (KPIs)

The pipeline measures repair metrics across execution cycles:

- **Repair Success Rate**: Percentage of generated patches that compile, pass all unit test assertions, and are successfully merged.
  $$\text{Success Rate} = \frac{\text{Merged Patches}}{\text{Proposed Patches}} \times 100$$
- **Mean Time to Repair (MTTR)**: The average duration (in seconds) between detecting a failure and committing the valid patch.
- **Regression Rate**: The percentage of patches that solve the targeted error but introduce secondary lint violations or test failures.
- **Inference Density**: Average CPU/GPU cycle seconds spent per successful line of code fixed.

---

## 2. Repository Health Score (RHS)

A global health metric is computed nightly to track codebase stability:

$$\text{RHS} = (\text{Test Coverage} \times 0.4) + (\text{Lint Cleanness} \times 0.3) - (\text{Open Vulns} \times 0.3)$$

- **Test Coverage**: Percentage of lines covered by unit tests.
- **Lint Cleanness**: Ratio of files with zero linter findings to total files.
- **Open Vulns**: Count of active security vulnerabilities flagged by `govulncheck`.

---

## 3. Metrics Visualization & Reports

Nightly maintenance scripts generate markdown summaries under `brain/performance/`:
- **Coverage Trends**: Plots of line coverage changes over historical git commits.
- **Agent Utilization**: Charts tracking VRAM allocation and prompt token throughput per agent model (e.g. Qwen vs DeepSeek).
