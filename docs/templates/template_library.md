# Reusable Code Template Library

This document provides reusable design specifications for the 13 core problem structures in **Challenge To YOU**.

---

## 1. Array Template (`template_array_ops`)
- **Narrative Wrapper**: Stabilizing power routing array values along memory buses.
- **Parameter Ranges**: Array size: 10 to 1,000 integers. Input range: `-10^5` to `10^5`.
- **Difficulty Knobs**: Array sorted (low) vs un-sorted (medium), sliding window checks required (high).
- **Generator Inputs**: Seed, target integer value.
- **Hidden Tests**: Empty list arrays, arrays with all identical values, boundary integer overflows.
- **Failure Cases**: Index out of bounds, execution timeout on un-optimized search loops.
- **Replay Metadata**: Track total comparison operations and array mutations.

---

## 2. Tree Template (`template_tree_balance`)
- **Narrative Wrapper**: Balancing abstract syntax tree branches in the Aether Library.
- **Parameter Ranges**: Node counts: 7 to 127 nodes. Tree depth: 3 to 10.
- **Difficulty Knobs**: Binary search tree (low) vs unbalanced AVL/Red-Black rotation requirements (high).
- **Generator Inputs**: List of node values to insert.
- **Hidden Tests**: Shuffled input sequences, nodes with duplicate values, deep linear branch trees (linked list structure).
- **Failure Cases**: Stack overflow on deep recursion, incorrect root assignment during rotation.
- **Replay Metadata**: Total tree depth after balances.

---

## 3. Graph Template (`template_graph_shortest`)
- **Narrative Wrapper**: Finding the shortest route between connection portals in the dream space.
- **Parameter Ranges**: Node vertices: 5 to 50 nodes. Edges: 10 to 200 weights.
- **Difficulty Knobs**: Acyclic positive weights (low) vs negative weights requiring cycle checks (high).
- **Generator Inputs**: Node adjacency list, start vertex, destination vertex.
- **Hidden Tests**: Disconnected graphs, loops containing negative weight cycles, multi-path overlaps.
- **Failure Cases**: Infinite loop search states, routing to non-existent nodes.
- **Replay Metadata**: Total nodes traversed, memory footprint of stack queue.

---

## 4. Simulation Template (`template_simulation_pid`)
- **Narrative Wrapper**: Configuring proportional-integral-derivative (PID) feedback loop for factory furnaces.
- **Parameter Ranges**: Temperature targets: 100C to 2000C. Sensor update frequency: 10ms to 100ms.
- **Difficulty Knobs**: Static target (low) vs varying target fluctuations under wind noise (high).
- **Generator Inputs**: Target value, P/I/D scalar constants.
- **Hidden Tests**: High environmental noise, sensor delays.
- **Failure Cases**: System oscillation limits exceeded, thermal safety shutdown trigger.
- **Replay Metadata**: Total time steps taken to stabilize target value.

---

## 5. Optimization Template (`template_opt_pipeline`)
- **Narrative Wrapper**: Reordering assembly instructions inside the CPU queue.
- **Parameter Ranges**: Operation queue size: 5 to 25. Registers available: R0 to R4.
- **Difficulty Knobs**: Zero branch predictions (low) vs dynamic branch targets with data hazards (high).
- **Generator Inputs**: Un-optimized instruction sequence list.
- **Hidden Tests**: RAW (Read-After-Write) hazard states, register write-lock blocks.
- **Failure Cases**: CPU pipeline stalls, writing to read-only register addresses.
- **Replay Metadata**: Execution clock cycles saved.

---

## 6. Debugging Template (`template_debug_trace`)
- **Narrative Wrapper**: Resolving off-by-one and syntax pointer faults in memory logs.
- **Parameter Ranges**: Code lines: 15 to 80 lines. Injected bug counts: 1 to 3.
- **Difficulty Knobs**: Single variable scopes (low) vs nested loop closures with pointer adjustments (high).
- **Generator Inputs**: Broken code template string, bug coordinate index.
- **Hidden Tests**: Empty string inputs, out-of-bound array queries.
- **Failure Cases**: Compilation failure, incomplete bug resolution.
- **Replay Metadata**: Time elapsed, editor compilations compiled.

---

## 7. Networking Template (`template_net_routing`)
- **Narrative Wrapper**: Configuring gossip protocol routes in mesh sensor grids.
- **Parameter Ranges**: Node counts: 10 to 100. Packet drop rates: 5% to 40%.
- **Difficulty Knobs**: Round-robin routing (low) vs Byzantine consensus agreements (high).
- **Generator Inputs**: Signal maps, node connection lists.
- **Hidden Tests**: Network partitioning (decoupled groups), packet data corruption.
- **Failure Cases**: Network packet loops, failure to reach 66% agreement.
- **Replay Metadata**: Total bytes sent to reach consensus.

---

## 8. Security Template (`template_sec_overflow`)
- **Narrative Wrapper**: Overwriting return pointers in gateway authentication buffers.
- **Parameter Ranges**: Buffer allocation size: 16 to 128 bytes. Key register addresses: 4-byte hex offsets.
- **Difficulty Knobs**: Single string inputs (low) vs bad-character filtering blocks (high).
- **Generator Inputs**: Target binary code, maximum payload size.
- **Hidden Tests**: Input filtration checkers (e.g. no null characters or loop symbols).
- **Failure Cases**: Segmentation fault reset, firewall signature detect block.
- **Replay Metadata**: Payload string size in bytes.

---

## 9. Database Template (`template_db_query`)
- **Narrative Wrapper**: Writing optimized queries to retrieve library schema records.
- **Parameter Ranges**: Table counts: 2 to 5. Records count: 1,000 to 100,000.
- **Difficulty Knobs**: Single table SELECTs (low) vs multi-JOIN aggregations with index requirements (high).
- **Generator Inputs**: SQL database tables, target query spec.
- **Hidden Tests**: Empty database tables, fields containing null values.
- **Failure Cases**: Query timeout (exceeding operation limits), incorrect record outputs.
- **Replay Metadata**: Total database rows scanned during execution.

---

## 10. Compiler Template (`template_compiler_parse`)
- **Narrative Wrapper**: Building grammar lexers for ancient runic text lines.
- **Parameter Ranges**: Token dictionary size: 5 to 20. Expression length: 10 to 100 characters.
- **Difficulty Knobs**: Simple regular expressions (low) vs recursive parser trees (high).
- **Generator Inputs**: Token dictionary, grammar rules (EBNF format).
- **Hidden Tests**: Nested parentheses, unbalanced tags, unmatched expressions.
- **Failure Cases**: Syntax parsing loops, memory overflow on deep trees.
- **Replay Metadata**: Total nodes parsed, AST depth.

---

## 11. AI Template (`template_ai_jailbreak`)
- **Narrative Wrapper**: Poisoning sentinel network weights to alter classification categories.
- **Parameter Ranges**: Feature vector dimensions: 2 to 10. Dataset sample size: 100 to 1,000 points.
- **Difficulty Knobs**: Prompt constraints (low) vs active coordinate adjustments in semantic spaces (high).
- **Generator Inputs**: Target classification model weights, input boundaries.
- **Hidden Tests**: Keyword filters, safety guard rails.
- **Failure Cases**: Target classification not achieved, safety flag raised.
- **Replay Metadata**: Weight adjustment steps taken, prompt length.

---

## 12. Distributed Systems Template (`template_dist_consensus`)
- **Narrative Wrapper**: Syncing state logs across three parallel database replicas.
- **Parameter Ranges**: Server replicas: 3 to 7 nodes. Network latency: 50ms to 500ms.
- **Difficulty Knobs**: Static master nodes (low) vs dynamic leader election loops (high).
- **Generator Inputs**: Node status logs, message queues.
- **Hidden Tests**: Sudden leader disconnects, network division splits.
- **Failure Cases**: State divergency between nodes, split-brain condition.
- **Replay Metadata**: Total messages sent to resolve conflicts.

---

## 13. Concurrency Template (`template_thread_sync`)
- **Narrative Wrapper**: Resolving thread access locks on a shared file system.
- **Parameter Ranges**: Active threads: 2 to 16. Shared locks: 1 to 5 resource channels.
- **Difficulty Knobs**: Basic mutex locks (low) vs atomic compare-and-swap (CAS) register updates (high).
- **Generator Inputs**: Concurrent code loop templates.
- **Hidden Tests**: Resource allocation deadlocks, thread context swaps.
- **Failure Cases**: Deadlock state (process freezes), variable value corruption.
- **Replay Metadata**: Thread wait-time duration in clock cycles.
