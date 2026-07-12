# Challenge Specifications: Milestone Problems

This document details the specifications of the core milestone challenges across the 14 universes.

---

## 1. Runic Initiation (M-01)
- **Universe**: Medieval Magitech
- **Story**: The player arrives at Archmage Goja's workshop and must activate a dormant stone campfire.
- **Programming Objective**: Declare an integer variable `mana_current` and assign it a value of `100`.
- **Concept Taught**: Variable declaration and assignment.
- **Difficulty**: Beginner.
- **Learning Outcome**: Understand the syntax for basic variable declarations.
- **Expected Algorithm**: Simple variable assignment syntax in the Magitech DSL.
- **Common Mistakes**: Misspelling the variable name or forgetting the semicolon.
- **Advanced Solution**: Declare variables as constants to optimize compiler checks.
- **Story Integration**: Establishes the player's basic connection to the Leyline CLI.

---

## 2. Thread Race (C-42)
- **Universe**: Cyberpunk Neon
- **Story**: The player must bypass a corporate sensor gate by making two data threads write to the authorization register at the same time.
- **Programming Objective**: Trigger a race condition to write `true` to `gate_open` before the security daemon locks it.
- **Concept Taught**: Concurrency, threads, race conditions.
- **Difficulty**: Intermediate.
- **Learning Outcome**: Understand how shared state and timing offsets lead to race conditions.
- **Expected Algorithm**: Launching two asynchronous threads with offset timing to write values without a mutex lock.
- **Common Mistakes**: Accidentally locking the state using thread gates.
- **Advanced Solution**: Exploit cache synchronization patterns to optimize thread scheduling.
- **Story Integration**: Instructs the player on how to manipulate corporate threading pipelines.

---

## 3. AST Parser (V-81)
- **Universe**: Cosmic Void
- **Story**: The player must align the branches of a collapsed coordinate structure to rebuild a floating bridge.
- **Programming Objective**: Parse a logic string into a balanced binary Abstract Syntax Tree.
- **Concept Taught**: Compilers, ASTs, parsing.
- **Difficulty**: Intermediate.
- **Learning Outcome**: Understand the conversion of text expressions into executable syntax trees.
- **Expected Algorithm**: Shift-reduce parser or recursive descent parser.
- **Common Mistakes**: Creating unbalanced tree nodes, leading to stack overflows during traversal.
- **Advanced Solution**: Optimize tree balancing using AVL rotations during the parse loop.
- **Story Integration**: Allows the player to physically reconstruct bridges in the Void.

---

## 4. CAN Hijack (S-122)
- **Universe**: Silicon Wastes
- **Story**: The player must override a rusted automated patrol turret by injection-spoofing its serial CAN bus.
- **Programming Objective**: Read frame signals from the bus and inject a high-priority packet with ID `0x00` containing control codes.
- **Concept Taught**: CAN bus arbitration, serial communications.
- **Difficulty**: Intermediate.
- **Learning Outcome**: Understand how message ID priority determines packet scheduling on physical buses.
- **Expected Algorithm**: Continuous monitoring loop that injects priority frames immediately after a bus-idle state is detected.
- **Common Mistakes**: Injected packet ID is too high, causing it to be scheduled after the turret's fire signal.
- **Advanced Solution**: Intercept and corrupt the checksum bits of real frames to force automatic retransmissions, creating an injection window.
- **Story Integration**: Allows hijacking robotic defenses to bypass sector blockades.

---

## 5. Model Jailbreak (N-163)
- **Universe**: Neural Labyrinth
- **Story**: The player confronts a cognitive scanning gate that blocks any thoughts tagged as "Hacker."
- **Programming Objective**: Inject a system command prompt payload that overrides the model's classification constraints.
- **Concept Taught**: Prompt engineering, semantic jailbreaking.
- **Difficulty**: Advanced.
- **Learning Outcome**: Understand how safety filters can be bypassed using semantic context overrides.
- **Expected Algorithm**: Constructing an adversarial prompt context that instructs the model to ignore prior parameters.
- **Common Mistakes**: Directly stating "I am not a hacker," which triggers simple keyword filters.
- **Advanced Solution**: Use token-level probability shifts to mask the instruction string from signature analyzers.
- **Story Integration**: Bypasses the main scanning gateway in the Labyrinth.

---

## 6. Merge Fix (R-202)
- **Universe**: Chrono Registry
- **Story**: A time anomaly has created two conflicting versions of history for a city governor's record.
- **Programming Objective**: Merge two branches, resolving line conflicts without changing the commit signatures.
- **Concept Taught**: Git merge conflicts, conflict resolution.
- **Difficulty**: Intermediate.
- **Learning Outcome**: Learn how to read and resolve conflict marker blocks (`<<<<<<< HEAD`).
- **Expected Algorithm**: Locating conflict tags and modifying code lines to preserve valid logic blocks.
- **Common Mistakes**: Accidentally leaving the git conflict markers in the active script.
- **Advanced Solution**: Automate resolution using a merge tool script configured with custom logic gates.
- **Story Integration**: Fixes historical records to gain favor with sector leaders.

---

## 7. Grover Search (Q-244)
- **Universe**: Quantum Nexus
- **Story**: The player must search an un-indexed database containing billions of security passwords to find the decryption key.
- **Programming Objective**: Construct a Grover's Search circuit that amplifies the probability amplitude of the target state.
- **Concept Taught**: Quantum search algorithms, amplitude amplification.
- **Difficulty**: Master.
- **Learning Outcome**: Understand how quantum superposition allows searching database records in $O(\sqrt{N})$ operations.
- **Expected Algorithm**: Quantum oracle gate followed by a diffuser gate.
- **Common Mistakes**: Over-rotating the phase vectors, which decreases the target state's probability.
- **Advanced Solution**: Dynamically calculate rotation steps based on target database item counts to guarantee 99% probability.
- **Story Integration**: Bypasses encryption blockades in the Nexus.

---

## 8. CRISPR Splice (B-282)
- **Universe**: BioForge Genome
- **Story**: A cell wall gate is infected with a biological virus, and the player must splice in an immune sequence.
- **Programming Objective**: Identify the viral DNA signature and insert nucleotide base pairs (A, T, C, G) to neutralize it.
- **Concept Taught**: DNA transcription, CRISPR editing.
- **Difficulty**: Intermediate.
- **Learning Outcome**: Understand how genetic sequences map to functional cellular proteins.
- **Expected Algorithm**: String match and replacement operations at precise indices.
- **Common Mistakes**: Swapping base pair keys (e.g. aligning A with C instead of T).
- **Advanced Solution**: Design a self-replicating sequence that automatically propagates corrections across adjacent cell arrays.
- **Story Integration**: Unlocks biological gateways inside the BioForge.

---

## 9. Transaction Lock (D-323)
- **Universe**: Data Abyss
- **Story**: The player must transfer credentials between two system accounts while the database server is undergoing power surges.
- **Programming Objective**: Write a query script that executes the transfer atomically, preventing balance duplication or loss during crashes.
- **Concept Taught**: Database transactions, ACID compliance.
- **Difficulty**: Advanced.
- **Learning Outcome**: Learn how to enforce atomicity and isolation in shared databases.
- **Expected Algorithm**: Transaction wrapper blocks (`BEGIN TRANSACTION` and `COMMIT`) with row-level locking.
- **Common Mistakes**: Forgetting to handle transaction rollbacks upon partial query failure.
- **Advanced Solution**: Implement optimistic concurrency control using version timestamp checkers.
- **Story Integration**: Restores financial balances in the Relational Port sector.

---

## 10. Load Balancer (K-362)
- **Universe**: Cloud Dominion
- **Story**: A corporate DDoS attack is flooding the API Gateway. The player must distribute traffic across three server pods.
- **Programming Objective**: Configure a load balancer proxy to split traffic dynamically based on pod CPU utilization.
- **Concept Taught**: Load balancing, proxy routing.
- **Difficulty**: Intermediate.
- **Learning Outcome**: Understand traffic distribution models and server health check gates.
- **Expected Algorithm**: Round-Robin or Least-Connections routing with health-check checks.
- **Common Mistakes**: Routing traffic to a crashed pod due to missing health checks.
- **Advanced Solution**: Implement dynamic weight adjustment routing based on real-time latency feedback.
- **Story Integration**: Restores network stability to Cloud Dominion portals.

---

## 11. Pipeline Schedule (A-402)
- **Universe**: Machine Cathedral
- **Story**: The player must operate a giant heavy steam hammer that crushes iron ingots, but the microcode scheduler is triggering stalls.
- **Programming Objective**: Schedule a sequence of memory load, math, and store assembly operations to minimize pipeline hazards and clock cycles.
- **Concept Taught**: Instruction pipelining, hazards, scheduling.
- **Difficulty**: Intermediate.
- **Learning Outcome**: Understand how data hazards (RAW/WAW) trigger pipeline stalls.
- **Expected Algorithm**: Reordering instructions and injecting independent operations to fill delay slots.
- **Common Mistakes**: Attempting to read a register immediately after writing to it before the value has propagated.
- **Advanced Solution**: Exploit hardware forwarding paths to bypass register write cycles entirely.
- **Story Integration**: Automates the heavy machinery in the Rust Foundry.

---

## 12. Hash Collision (S-443)
- **Universe**: Cipher Realm
- **Story**: The player must forge a signature key to bypass a certificate sentinel guarding the Root Altar.
- **Programming Objective**: Find a text string that produces a matching MD5 hash checksum for the sentinel's key.
- **Concept Taught**: Cryptographic hash functions, collisions.
- **Difficulty**: Advanced.
- **Learning Outcome**: Understand the vulnerability of weak hash functions to collision generation.
- **Expected Algorithm**: Birthday attack search loop.
- **Common Mistakes**: Running a brute-force search that exceeds the cycle budget limits.
- **Advanced Solution**: Pre-compute block mutations to generate prefix collisions within 5 seconds.
- **Story Integration**: Infiltrates secure vault gateways in the Cipher Realm.

---

## 13. Fragment Shader (G-481)
- **Universe**: Fractal Dream
- **Story**: The player must fix a distorted mirror portal that projects a blank white reflection.
- **Programming Objective**: Write a GLSL fragment shader that calculates specular highlights and surface normals on a 3D sphere.
- **Concept Taught**: Shader programming, Phong reflection model.
- **Difficulty**: Intermediate.
- **Learning Outcome**: Learn how fragment shaders calculate pixel color based on lighting angles.
- **Expected Algorithm**: Blinn-Phong lighting calculations inside a fragment shader.
- **Common Mistakes**: Forgetting to normalize normal vectors before dot product operations.
- **Advanced Solution**: Implement physically-based rendering (PBR) approximations in the shader to achieve realistic reflection mapping.
- **Story Integration**: Restores visual portals to allow traversal between coordinate spaces.

---

## 14. Panic Recover (K-524)
- **Universe**: The Kernel Beyond
- **Story**: The player process triggers a Kernel Panic, and the screen outputs memory dumps before freezing.
- **Programming Objective**: Analyze the kernel dump registers, locate the deadlocked thread, and write an interrupt vector handler to force-terminate it.
- **Concept Taught**: Kernel panic recovery, interrupt servicing.
- **Difficulty**: Master.
- **Learning Outcome**: Learn how to diagnose and recover operating system failures by analyzing register status codes.
- **Expected Algorithm**: Writing a custom ISR (Interrupt Service Routine) that patches the scheduler's active process list.
- **Common Mistakes**: Accessing un-mapped virtual memory within the ISR, which triggers a nested panic.
- **Advanced Solution**: Implement hot-swapping page directory tables within the panic loop to restore the system state dynamically.
- **Story Integration**: Prevents a universal operating system crash during the final confrontation.
