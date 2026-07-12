# Discarded Repositories

These repositories were cloned but provide no reusable code—only curated link lists.

| Repository | Reason |
|------------|--------|
| `awesome-godot` | 50KB README.md with categorized links to plugins, templates, tutorials. No source code. |
| `awesome-competitive-programming` | 53KB markdown index of CP resources (books, sites, OJs). No implementations. |
| `awesome-repos` | Generic "awesome list" of GitHub repos across many topics. No Godot/Go relevance. |

## Cleanup

```bash
rm -rf /Users/fallofpheonix/Project/research/awesome-godot
rm -rf /Users/fallofpheonix/Project/research/awesome-competitive-programming
rm -rf /Users/fallofpheonix/Project/research/awesome-repos
```

## Verification

After cleanup, `research/` should contain only:

```
research/
├── README.md
├── repository_summary.md
├── architecture_notes.md
├── reusable_components.md
├── gameplay_mechanics.md
├── ui_patterns.md
├── algorithms.md
├── go_patterns.md
├── implementation_plan.md
├── recommendation_table.md
├── discard.md
├── godot-go/                    # Critical
├── godot-go-demo-projects/      # Critical
├── gdextension/                 # Medium (reference)
├── godot-coding-challenge/      # High
├── dothop/                      # Critical
├── GodotSudoku/                 # High
├── Sudoku-puzzle-generator/     # High
├── godot_puzzle_dependencies/   # Critical
├── godot_recipes/               # High
├── coding-challenges/           # High
└── competitive-programming/     # Critical
```

Total: **11 keep / 3 discard**