# Algorithms Reference

Core algorithms extracted from competitive programming repositories, adapted for game mechanics.

---

## 1. Graph Algorithms (Dependency Resolution)

### From: `competitive-programming/Graph/` + `godot_puzzle_dependencies`

### Topological Sort (Kahn's Algorithm)
```go
// generator/topo.go
func TopologicalSort(nodes map[string]*PuzzleNode) ([]string, error) {
    // Compute in-degrees
    inDegree := make(map[string]int)
    for id, node := range nodes {
        inDegree[id] = 0
    }
    for _, node := range nodes {
        for _, prereq := range node.Prerequisites {
            inDegree[node.ID]++
        }
    }
    
    // Queue nodes with zero in-degree
    queue := []string{}
    for id, deg := range inDegree {
        if deg == 0 {
            queue = append(queue, id)
        }
    }
    
    result := []string{}
    for len(queue) > 0 {
        n := queue[0]
        queue = queue[1:]
        result = append(result, n)
        
        for _, unlock := range nodes[n].Unlocks {
            inDegree[unlock]--
            if inDegree[unlock] == 0 {
                queue = append(queue, unlock)
            }
        }
    }
    
    if len(result) != len(nodes) {
        return nil, ErrCycleDetected
    }
    return result, nil
}
```

### Cycle Detection (DFS)
```go
func DetectCycles(nodes map[string]*PuzzleNode) [][]string {
    var cycles [][]string
    visited := make(map[string]bool)
    recStack := make(map[string]bool)
    path := []string{}
    
    var dfs func(string) bool
    dfs = func(u string) bool {
        visited[u] = true
        recStack[u] = true
        path = append(path, u)
        
        for _, v := range nodes[u].Prerequisites {
            if !visited[v] {
                if dfs(v) { return true }
            } else if recStack[v] {
                // Found cycle - extract it
                idx := slices.Index(path, v)
                cycles = append(cycles, path[idx:])
                return true
            }
        }
        
        recStack[u] = false
        path = path[:len(path)-1]
        return false
    }
    
    for id := range nodes {
        if !visited[id] {
            dfs(id)
        }
    }
    return cycles
}
```

### A* Pathfinding (For Cyberpunk Netrunning)
```go
// analyzer/astar.go
type Node struct {
    ID       string
    Position Vector2
    Cost     float64  // g-cost
    Heuristic float64 // h-cost
    Parent   *Node
}

func AStar(start, goal *Node, graph map[string][]Edge) []*Node {
    openSet := NewPriorityQueue()
    openSet.Push(start, start.Cost+start.Heuristic)
    closedSet := make(map[string]bool)
    
    for !openSet.Empty() {
        current := openSet.Pop()
        if current.ID == goal.ID {
            return reconstructPath(current)
        }
        closedSet[current.ID] = true
        
        for _, edge := range graph[current.ID] {
            neighbor := graph.nodes[edge.To]
            if closedSet[neighbor.ID] { continue }
            
            tentativeG := current.Cost + edge.Weight
            if tentativeG < neighbor.Cost {
                neighbor.Parent = current
                neighbor.Cost = tentativeG
                neighbor.Heuristic = heuristic(neighbor, goal)
                openSet.Push(neighbor, neighbor.Cost+neighbor.Heuristic)
            }
        }
    }
    return nil // No path
}

func heuristic(a, b *Node) float64 {
    return a.Position.DistanceTo(b.Position)
}
```

---

## 2. Dynamic Programming (Optimization Challenges)

### From: `competitive-programming/Dynamic Programming/`

### Longest Increasing Subsequence (LIS) - O(n log n)
```go
// vm/builtin/lis.go
func BuiltinLIS(args ...Object) Object {
    arr := args[0].(*Array).Elements
    n := len(arr)
    if n == 0 { return &Array{} }
    
    tails := make([]Object, n)
    size := 0
    
    for _, x := range arr {
        // Binary search for insertion point
        i, j := 0, size
        for i != j {
            m := (i + j) / 2
            if Less(tails[m], x) {
                i = m + 1
            } else {
                j = m
            }
        }
        tails[i] = x
        if i == size { size++ }
    }
    
    // Reconstruct (simplified - returns length)
    return &Integer{Value: int64(size)}
}
```

### Knapsack (Resource Allocation)
```go
// generator/knapsack.go - For challenge reward optimization
func Knapsack(items []Item, capacity int) (int, []int) {
    n := len(items)
    dp := make([][]int, n+1)
    for i := range dp {
        dp[i] = make([]int, capacity+1)
    }
    
    for i := 1; i <= n; i++ {
        for w := 0; w <= capacity; w++ {
            dp[i][w] = dp[i-1][w]
            if items[i-1].Weight <= w {
                val := dp[i-1][w-items[i-1].Weight] + items[i-1].Value
                if val > dp[i][w] {
                    dp[i][w] = val
                }
            }
        }
    }
    
    // Trace back selected items
    selected := []int{}
    w := capacity
    for i := n; i > 0; i-- {
        if dp[i][w] != dp[i-1][w] {
            selected = append(selected, i-1)
            w -= items[i-1].Weight
        }
    }
    return dp[n][capacity], selected
}
```

### Edit Distance (Code Similarity / Ghost Mode)
```go
// analyzer/editdist.go
func EditDistance(a, b string) int {
    m, n := len(a), len(b)
    dp := make([][]int, m+1)
    for i := range dp {
        dp[i] = make([]int, n+1)
    }
    
    for i := 0; i <= m; i++ { dp[i][0] = i }
    for j := 0; j <= n; j++ { dp[0][j] = j }
    
    for i := 1; i <= m; i++ {
        for j := 1; j <= n; j++ {
            if a[i-1] == b[j-1] {
                dp[i][j] = dp[i-1][j-1]
            } else {
                dp[i][j] = 1 + min(
                    dp[i-1][j],   // delete
                    dp[i][j-1],   // insert
                    dp[i-1][j-1], // replace
                )
            }
        }
    }
    return dp[m][n]
}
```

---

## 3. Bit Manipulation (Flags, Permissions, Crypto)

### From: `competitive-programming/Bit-Manipulations/`

```go
// vm/builtin/bit.go
func BuiltinBitCount(args ...Object) Object {
    n := args[0].(*Integer).Value
    return &Integer{Value: int64(bits.OnesCount64(uint64(n)))}
}

func BuiltinBitMask(args ...Object) Object {
    indices := args[0].(*Array).Elements
    var mask uint64
    for _, idx := range indices {
        bit := idx.(*Integer).Value
        if bit >= 0 && bit < 64 {
            mask |= 1 << bit
        }
    }
    return &Integer{Value: int64(mask)}
}

func BuiltinHasBit(args ...Object) Object {
    n := uint64(args[0].(*Integer).Value)
    bit := args[1].(*Integer).Value
    return &Boolean{Value: (n & (1 << bit)) != 0}
}

func BuiltinSubsetEnumeration(args ...Object) Object {
    mask := uint64(args[0].(*Integer).Value)
    var subsets []Object
    
    for sub := mask; sub > 0; sub = (sub - 1) & mask {
        subsets = append(subsets, &Integer{Value: int64(sub)})
    }
    subsets = append(subsets, &Integer{Value: 0})
    return &Array{Elements: subsets}
}
```

---

## 4. String Algorithms (Pattern Matching, Parsing)

### From: `competitive-programming/String/`

### KMP (Knuth-Morris-Pratt) - O(n + m)
```go
// analyzer/kmp.go
func BuildLPS(pattern string) []int {
    lps := make([]int, len(pattern))
    length := 0
    i := 1
    for i < len(pattern) {
        if pattern[i] == pattern[length] {
            length++
            lps[i] = length
            i++
        } else {
            if length != 0 {
                length = lps[length-1]
            } else {
                lps[i] = 0
                i++
            }
        }
    }
    return lps
}

func KMPSearch(text, pattern string) []int {
    lps := BuildLPS(pattern)
    var matches []int
    i, j := 0, 0
    for i < len(text) {
        if text[i] == pattern[j] {
            i++; j++
        }
        if j == len(pattern) {
            matches = append(matches, i-j)
            j = lps[j-1]
        } else if i < len(text) && text[i] != pattern[j] {
            if j != 0 {
                j = lps[j-1]
            } else {
                i++
            }
        }
    }
    return matches
}
```

### Z-Algorithm (String Matching)
```go
func ZAlgorithm(s string) []int {
    n := len(s)
    z := make([]int, n)
    l, r := 0, 0
    for i := 1; i < n; i++ {
        if i <= r {
            z[i] = min(r-i+1, z[i-l])
        }
        for i+z[i] < n && s[z[i]] == s[i+z[i]] {
            z[i]++
        }
        if i+z[i]-1 > r {
            l, r = i, i+z[i]-1
        }
    }
    z[0] = n
    return z
}
```

---

## 5. Number Theory (Cyberpunk Crypto)

### From: `competitive-programming/Number Theory/` + `competitive-programming/Cryptographic algos/`

### GCD / Extended Euclidean
```go
func GCD(a, b int64) int64 {
    for b != 0 {
        a, b = b, a%b
    }
    return a
}

func ExtendedGCD(a, b int64) (gcd, x, y int64) {
    if a == 0 {
        return b, 0, 1
    }
    gcd, x1, y1 := ExtendedGCD(b%a, a)
    x = y1 - (b/a)*x1
    y = x1
    return gcd, x, y
}

// Modular inverse
func ModInverse(a, m int64) int64 {
    gcd, x, _ := ExtendedGCD(a, m)
    if gcd != 1 {
        return 0 // No inverse
    }
    return (x%m + m) % m
}
```

### Sieve of Eratosthenes
```go
func Sieve(n int) []int {
    isPrime := make([]bool, n+1)
    for i := 2; i <= n; i++ { isPrime[i] = true }
    
    for p := 2; p*p <= n; p++ {
        if isPrime[p] {
            for i := p * p; i <= n; i += p {
                isPrime[i] = false
            }
        }
    }
    
    var primes []int
    for i := 2; i <= n; i++ {
        if isPrime[i] { primes = append(primes, i) }
    }
    return primes
}
```

### Fast Modular Exponentiation
```go
func ModPow(base, exp, mod int64) int64 {
    result := int64(1)
    base %= mod
    for exp > 0 {
        if exp&1 == 1 {
            result = (result * base) % mod
        }
        base = (base * base) % mod
        exp >>= 1
    }
    return result
}
```

---

## 6. Backtracking (Puzzle Generation)

### From: `Sudoku-puzzle-generator` + `competitive-programming/Backtracking/`

### General Backtracking Framework
```go
// generator/backtrack.go
type Backtracker struct {
    State       interface{}
    Constraints []Constraint
    Solutions   []interface{}
    MaxSolutions int
}

type Constraint func(interface{}) bool

func (b *Backtracker) Solve(candidates []Choice, choose func(interface{}, Choice), unchoose func(interface{}, Choice), isComplete func(interface{}) bool) {
    if b.MaxSolutions > 0 && len(b.Solutions) >= b.MaxSolutions {
        return
    }
    
    if isComplete(b.State) {
        b.Solutions = append(b.Solutions, copyState(b.State))
        return
    }
    
    for _, choice := range candidates {
        choose(b.State, choice)
        valid := true
        for _, c := range b.Constraints {
            if !c(b.State) {
                valid = false
                break
            }
        }
        if valid {
            b.Solve(candidates, choose, unchoose, isComplete)
        }
        unchoose(b.State, choice)
    }
}

// Sudoku-specific
func GenerateSudoku(difficulty float64) [][]int {
    grid := make([][]int, 9)
    for i := range grid {
        grid[i] = make([]int, 9)
    }
    
    // 1. Fill diagonal boxes
    fillDiagonalBoxes(grid)
    
    // 2. Fill remaining with backtracking
    backtrack(grid, 0, 3)
    
    // 3. Remove clues based on difficulty
    removeClues(grid, difficulty)
    
    return grid
}

func backtrack(grid [][]int, row, col int) bool {
    if row == 9 { return true }
    if col == 9 { return backtrack(grid, row+1, 0) }
    if grid[row][col] != 0 { return backtrack(grid, row, col+1) }
    
    nums := rand.Perm(9)
    for _, n := range nums {
        n++
        if isValid(grid, row, col, n) {
            grid[row][col] = n
            if backtrack(grid, row, col+1) { return true }
            grid[row][col] = 0
        }
    }
    return false
}
```

---

## 7. Greedy Algorithms (Resource Management)

### From: `competitive-programming/Greedy/`

### Interval Scheduling (Challenge Scheduling)
```go
func MaxNonOverlappingIntervals(intervals []Interval) []Interval {
    // Sort by end time
    sort.Slice(intervals, func(i, j int) bool {
        return intervals[i].End < intervals[j].End
    })
    
    var result []Interval
    lastEnd := -1
    for _, iv := range intervals {
        if iv.Start >= lastEnd {
            result = append(result, iv)
            lastEnd = iv.End
        }
    }
    return result
}
```

### Huffman Coding (Data Compression - Saboteur)
```go
type HuffmanNode struct {
    Char  rune
    Freq  int
    Left  *HuffmanNode
    Right *HuffmanNode
}

func BuildHuffmanTree(freq map[rune]int) *HuffmanNode {
    pq := NewMinHeap()
    for c, f := range freq {
        pq.Push(&HuffmanNode{Char: c, Freq: f}, f)
    }
    
    for pq.Len() > 1 {
        left := pq.Pop().(*HuffmanNode)
        right := pq.Pop().(*HuffmanNode)
        merged := &HuffmanNode{
            Freq:  left.Freq + right.Freq,
            Left:  left,
            Right: right,
        }
        pq.Push(merged, merged.Freq)
    }
    return pq.Pop().(*HuffmanNode)
}
```

---

## 8. Randomized Algorithms (Procedural Generation)

### Fisher-Yates Shuffle
```go
func Shuffle[T any](slice []T, rng *rand.Rand) {
    for i := len(slice) - 1; i > 0; i-- {
        j := rng.Intn(i + 1)
        slice[i], slice[j] = slice[j], slice[i]
    }
}
```

### Reservoir Sampling (Stream Sampling)
```go
func ReservoirSample[T any](stream <-chan T, k int, rng *rand.Rand) []T {
    reservoir := make([]T, k)
    i := 0
    for item := range stream {
        if i < k {
            reservoir[i] = item
        } else {
            j := rng.Intn(i + 1)
            if j < k {
                reservoir[j] = item
            }
        }
        i++
    }
    return reservoir[:min(k, i)]
}
```

### Weighted Random Selection
```go
func WeightedChoice[T any](items []T, weights []float64, rng *rand.Rand) T {
    total := 0.0
    for _, w := range weights { total += w }
    
    r := rng.Float64() * total
    for i, w := range weights {
        r -= w
        if r <= 0 { return items[i] }
    }
    return items[len(items)-1]
}
```

---

## 9. Geometry (Spatial Puzzles)

### From: `competitive-programming/` (Geometry section)

### Convex Hull (Graham Scan)
```go
type Point struct { X, Y float64 }

func Cross(o, a, b Point) float64 {
    return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X)
}

func ConvexHull(points []Point) []Point {
    sort.Slice(points, func(i, j int) bool {
        return points[i].X < points[j].X || (points[i].X == points[j].X && points[i].Y < points[j].Y)
    })
    
    lower := []Point{}
    for _, p := range points {
        for len(lower) >= 2 && Cross(lower[len(lower)-2], lower[len(lower)-1], p) <= 0 {
            lower = lower[:len(lower)-1]
        }
        lower = append(lower, p)
    }
    
    upper := []Point{}
    for i := len(points) - 1; i >= 0; i-- {
        p := points[i]
        for len(upper) >= 2 && Cross(upper[len(upper)-2], upper[len(upper)-1], p) <= 0 {
            upper = upper[:len(upper)-1]
        }
        upper = append(upper, p)
    }
    
    // Remove duplicate endpoints
    lower = lower[:len(lower)-1]
    upper = upper[:len(upper)-1]
    return append(lower, upper...)
}
```

---

## 10. Algorithm Complexity Cheatsheet

| Algorithm | Time | Space | Use Case |
|-----------|------|-------|----------|
| Topo Sort | O(V+E) | O(V) | Puzzle dependencies |
| Cycle Detect | O(V+E) | O(V) | Validation |
| A* | O(b^d) | O(b^d) | Netrunning pathfinding |
| LIS | O(n log n) | O(n) | Sequence optimization |
| Knapsack | O(nW) | O(nW) | Reward allocation |
| Edit Distance | O(mn) | O(mn) | Code similarity |
| KMP | O(n+m) | O(m) | Pattern search |
| Z-Algorithm | O(n) | O(n) | String matching |
| GCD | O(log min) | O(1) | Crypto |
| ModPow | O(log e) | O(1) | Crypto |
| Backtrack | O(b^d) | O(d) | Puzzle generation |
| Huffman | O(n log n) | O(n) | Compression |
| Fisher-Yates | O(n) | O(1) | Shuffling |
| Reservoir | O(n) | O(k) | Stream sampling |
| Graham Scan | O(n log n) | O(n) | Convex hull |

---

## 11. Integration into VM Builtins

```go
// vm/builtin/algorithms.go
var AlgorithmBuiltins = []*Builtin{
    {Name: "topo_sort", Fn: BuiltinTopoSort},
    {Name: "has_cycle", Fn: BuiltinHasCycle},
    {Name: "a_star", Fn: BuiltinAStar},
    {Name: "lis", Fn: BuiltinLIS},
    {Name: "knapsack", Fn: BuiltinKnapsack},
    {Name: "edit_distance", Fn: BuiltinEditDistance},
    {Name: "kmp_search", Fn: BuiltinKMPSearch},
    {Name: "gcd", Fn: BuiltinGCD},
    {Name: "mod_inv", Fn: BuiltinModInverse},
    {Name: "mod_pow", Fn: BuiltinModPow},
    {Name: "sieve", Fn: BuiltinSieve},
    {Name: "backtrack", Fn: BuiltinBacktrack},
    {Name: "shuffle", Fn: BuiltinShuffle},
    {Name: "weighted_choice", Fn: BuiltinWeightedChoice},
    {Name: "convex_hull", Fn: BuiltinConvexHull},
}
```

### Pscript Usage Examples
```pscript
# Dependency resolution
let order = topo_sort(puzzle_graph)
if has_cycle(puzzle_graph) { error("Circular dependency!") }

# Optimization
let best_path = a_star(start_node, goal_node, graph)

# Sequence analysis
let len = lis(sequence)

# Resource allocation
let (value, items) = knapsack(items, capacity)

# Code comparison (Ghost mode)
let similarity = edit_distance(original, modified)

# Pattern search
let matches = kmp_search(log_data, "ERROR")

# Crypto (Cyberpunk)
let inv = mod_inv(key, prime)
let cipher = mod_pow(plaintext, exp, mod)
```