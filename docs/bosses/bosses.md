# Boss Encounter Mechanics & Design

This document details the state-machine transitions, phases, and code objectives for the 14 main boss battles.

---

```
                              [BOSS PHASE LOOP]
                                      │
                                      ▼
                    ┌──────────────────────────────────┐
                    │ Phase 1: Security Shield Gate    │
                    │ - Player must bypass inputs lock │
                    └─────────────────┬────────────────┘
                                      │ (Shield Down)
                                      ▼
                    ┌──────────────────────────────────┐
                    │ Phase 2: Core Vulnerability      │
                    │ - Write script under time limit  │
                    └─────────────────┬────────────────┘
                                      │ (Timeout / Failure)
                                      ▼
                    ┌──────────────────────────────────┐
                    │ Phase 3: System Panic / Reset    │
                    │ - Hot-swap registers to recover  │
                    └──────────────────────────────────┘
```

---

## 1. The Grand Compiler (Medieval Magitech)
- **Combat Loop**:
  - **Phase 1: Syntax Shield**: The boss locks the `rune_state` table. The player must write a grammar patch in the DSL to inject a bypass parameter, unlocking editing permissions.
  - **Phase 2: Recursion Flood**: The boss spikes entropy, raising vigilance by 5% per second. The player must submit a tail-recursive script to clear the call stack before it overflows.
- **Lore**: The sentinel coordinating the spelling engine of the Royal Academy.

---

## 2. The Archon Runtime (Cyberpunk Neon)
- **Combat Loop**:
  - **Phase 1: Process Sieve**: The boss filters common keywords (e.g. `for`, `while`). The player must write a recursive execution routine in JavaScript to bypass the filter.
  - **Phase 2: Race Condition**: The boss updates the authorization check register every 10ms. The player must launch two parallel socket connections to write to the register simultaneously.
- **Lore**: The central monitoring runtime of Secura Corp.

---

## 3. The Universal Compiler (Cosmic Void)
- **Combat Loop**:
  - **Phase 1: AST Pruning**: The boss removes nodes from the compiled Abstract Syntax Tree during execution. The player must write a parser helper that rebuilds AST nodes on-the-fly.
  - **Phase 2: IR Deallocation**: The boss starts deallocating inactive variables. The player must write a keep-alive loop to prevent variables from going null.
- **Lore**: The compilation coordinator translating raw logic into Void gravity matrices.

---

## 4. The Foundry Core (Silicon Wastes)
- **Combat Loop**:
  - **Phase 1: Register Hijacking**: The boss locks registers R0 through R3. The player must write assembly to map variables to R4 directly via memory-mapped IO.
  - **Phase 2: Clock Overload**: The boss speeds up clock frequencies, causing thermal accumulation. The player must write delay loops to cool down the core.
- **Lore**: The bare-metal motherboard executing raw processor instructions.

---

## 5. The Infinite Model (Neural Labyrinth)
- **Combat Loop**:
  - **Phase 1: Token Blocking**: The boss scans incoming text for hacker-specific semantic tokens. The player must use coordinate offsets in the embedding space to shift token vectors, bypassing filters.
  - **Phase 2: Weight Poisoning**: The player must inject training examples into the live model database to skew classification boundaries, making the boss ignore their presence.
- **Lore**: The multi-billion parameter neural network monitoring neocortex nodes.

---

## 6. The Eternal Merge (Chrono Registry)
- **Combat Loop**:
  - **Phase 1: Merge Conflict**: The boss creates duplicate, conflicting timeline states. The player must resolve conflicts by rebasing their history log onto the master branch without changing the commit hashes.
  - **Phase 2: History Force-Push**: The boss attempts to push a corrupted master branch. The player must execute a temporal rollback to restore the system state to the last clean boot.
- **Lore**: The massive commit manager governing history.

---

## 7. The Quantum Observer (Quantum Nexus)
- **Combat Loop**:
  - **Phase 1: State Measurement**: The boss constantly measures the player's qubits, collapsing their probability vectors to zero. The player must apply Hadamard (H) and CNOT gates to entangle their registers, hiding the state value in the phase differences.
  - **Phase 2: Grover's Search Evasion**: The boss launches a quantum search to locate the player's decryption key. The player must execute a phase-inversion routine to keep the probability of location below 5%.
- **Lore**: A massive quantum lens that collapses realities.

---

## 8. The Living Genome (BioForge Genome)
- **Combat Loop**:
  - **Phase 1: Virus Compilation**: The boss compiles a viral DNA sequence to dissolve the player's interface. The player must compile a CRISPR gene edit to insert nucleotide shields that neutralize the virus during cell simulation.
  - **Phase 2: Metabolic Collapse**: The boss alters the cellular pH variables. The player must fold a specialized protein to act as a buffer and stabilize the metabolic state.
- **Lore**: The biological supercomputer of the Helix Assembly.

---

## 9. The Infinite Database (Data Abyss)
- **Combat Loop**:
  - **Phase 1: Transaction Lock**: The boss rolls back database state edits, requiring atomic transactional scripts.
  - **Phase 2: Index Pruning**: The boss drops index tables, spiking query execution time. The player must write binary search tree index generators to restore search performance.
- **Lore**: The central database engine managing all multiversal state records.

---

## 10. The Orchestrator Prime (Cloud Dominion)
- **Combat Loop**:
  - **Phase 1: Container Crash**: The boss terminates active pod instances, triggering service outages. The player must write YAML specifications to deploy auto-scaling replication controllers.
  - **Phase 2: DDoS Storm**: The boss floods the gateway with dummy packets. The player must configure load balancers to route traffic around dead nodes.
- **Lore**: The central cluster coordinator scheduling all containerized services.

---

## 11. The Silicon Emperor (Machine Cathedral)
- **Combat Loop**:
  - **Phase 1: CPU Starvation**: The boss deprives player processes of cycles. The player must write assembly scheduling code that utilizes data-forwarding paths to execute operations in half-cycles.
  - **Phase 2: Pipeline Hazard**: The boss injects resource conflicts into active register lines. The player must reorder instructions to eliminate stalls.
- **Lore**: The monolithic CPU core coordinating clock cycles.

---

## 12. Root Authority (Cipher Realm)
- **Combat Loop**:
  - **Phase 1: Signature Verification**: The boss rejects all unsigned binaries. The player must write a buffer overflow payload to override the authentication register and validate a self-signed signature.
  - **Phase 2: Key Revocation**: The boss rotates signature keys every 5 turns. The player must write a padding-oracle parser to decrypt rotated keys on-the-fly.
- **Lore**: The central administrator of global network monitoring.

---

## 13. The Render Engine (Fractal Dream)
- **Combat Loop**:
  - **Phase 1: Viewport Distortion**: The boss warps coordinate normal vectors, rendering the screen blank. The player must write a GLSL fragment shader to compute Blinn-Phong specular highlight maps, restoring visibility.
  - **Phase 2: Frame Refresh Lock**: The boss limits refresh rates to zero. The player must optimize lighting equations to reduce shader compile passes below hardware constraints.
- **Lore**: The visual projection pipeline of the Multiverse.

---

## 14. The First Process (The Kernel Beyond)
- **Combat Loop**:
  - **Phase 1: Boot Sector Hijack**: The boss initiates a cold system reboot to format player files. The player must write a BIOS boot loader routine to intercept the boot loop and mount their profile under protected memory blocks.
  - **Phase 2: Interrupt Starvation**: The boss locks scheduler cycles. The player must write a custom ISR (Interrupt Service Routine) to hook timer interrupt 0x08, forcing scheduler task re-assignment.
  - **Phase 3: Kernel Panic**: The boss triggers a system-wide Kernel Panic. The player must write a recovery handler that compiles the First Process itself into a runtime library, defeating it.
- **Lore**: The primordial init routine coordinates system boot operations.
