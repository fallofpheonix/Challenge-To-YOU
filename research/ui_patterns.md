# UI Patterns

Godot UI patterns and reusable components from research.

---

## 1. Terminal / Code Editor (Primary Interface)

### From: `godot-go-demo-projects` + `godot_recipes` (UI patterns)

### Editor Component Structure
```
TerminalEditor (Control)
├── LineNumbers (RichTextLabel)
├── CodeArea (TextEdit)
│   └── SyntaxHighlighter (extends TextEdit)
├── OutputTerminal (RichTextLabel)
│   ├── ANSI color support
│   ├── Scroll lock toggle
│   └── Clear button
├── StatusBar (HBoxContainer)
│   ├── Era indicator
│   ├── Mode indicator
│   ├── Vigilance meter (ProgressBar)
│   ├── Luck display
│   └── Passcode display
└── CommandPalette (PopupPanel)
    └── Fuzzy search for commands
```

### Syntax Highlighting (TextEdit + SyntaxHighlighter)
```gdscript
# SyntaxHighlighter.gd
extends TextEdit
class_name SyntaxHighlighter

@tool

var highlighter: SyntaxHighlighter

func _ready():
    highlighter = SyntaxHighlighter.new()
    syntax_highlighter = highlighter
    
    # Pscript (Magitech) keywords
    highlighter.add_keyword_color("rune", Color.CYAN)
    highlighter.add_keyword_color("bind", Color.YELLOW)
    highlighter.add_keyword_color("channel", Color.YELLOW)
    highlighter.add_keyword_color("release", Color.YELLOW)
    highlighter.add_keyword_color("if", Color.ORANGE)
    highlighter.add_keyword_color("then", Color.ORANGE)
    highlighter.add_keyword_color("else", Color.ORANGE)
    highlighter.add_keyword_color("end", Color.ORANGE)
    highlighter.add_keyword_color("fn", Color.GREEN)
    highlighter.add_keyword_color("return", Color.GREEN)
    
    # Comments
    highlighter.set_comment_color(Color.GRAY)
    
    # Strings
    highlighter.set_string_color(Color.LIME)
    
    # Numbers
    highlighter.set_number_color(Color.GOLD)
    
    # Operators
    highlighter.add_symbol_color(":", Color.WHITE)
    highlighter.add_symbol_color("=>", Color.CYAN)
```

### Code Editor Features
| Feature | Implementation |
|---------|----------------|
| Line numbers | Separate `RichTextLabel` synced to `TextEdit` scroll |
| Bracket matching | `TextEdit` built-in `matching_bracket` |
| Auto-indent | `TextEdit` `auto_indent` property |
| Code folding | `TextEdit` `fold_all`/`unfold_all` |
| Minimap | Custom `Control` with scaled text preview |
| Multi-cursor | `TextEdit` `caret_multiple` |
| Find/Replace | Custom `PopupPanel` with regex support |

---

## 2. Vigilance Meter (Ghost Mode)

### From: `CHALLENGE-TO-YOU-PLAN.md` + `dothop` (StatLogger)

```
VigilanceMeter (HBoxContainer)
├── Icon (TextureRect) - Eye icon
├── Bar (ProgressBar)
│   ├── Green (0-30%)
│   ├── Yellow (30-60%)
│   ├── Orange (60-85%)
│   └── Red (85-100%) + pulse animation
├── Label (Label) - "45%"
└── Tooltip (PopupPanel) - Details on hover
```

```gdscript
# VigilanceMeter.gd
extends HBoxContainer
class_name VigilanceMeter

@onready var bar = %Bar
@onready var label = %Label
@onready var icon = %Icon

func set_vigilance(value: float) -> void:
    value = clamp(value, 0.0, 1.0)
    bar.value = value * 100
    label.text = "%d%%" % (value * 100)
    
    # Color transitions
    if value < 0.3:
        bar.set_theme_color_override("fill_color", Color.GREEN)
        icon.modulate = Color.GREEN
    elif value < 0.6:
        bar.set_theme_color_override("fill_color", Color.YELLOW)
        icon.modulate = Color.YELLOW
    elif value < 0.85:
        bar.set_theme_color_override("fill_color", Color.ORANGE)
        icon.modulate = Color.ORANGE
    else:
        bar.set_theme_color_override("fill_color", Color.RED)
        icon.modulate = Color.RED
        # Pulse animation
        if not icon.is_playing():
            icon.play("pulse")
```

---

## 3. Puzzle Dependency Graph (Map View)

### From: `godot_puzzle_dependencies` + `dothop`

```
PuzzleMap (Control)
├── GraphView (Control) - Custom drawing
│   ├── Node (Button) - Puzzle node
│   │   ├── State: LOCKED (gray)
│   │   ├── State: AVAILABLE (glowing)
│   │   ├── State: ACTIVE (pulsing)
│   │   └── State: COMPLETED (checkmark)
│   ├── Edge (Line2D) - Prerequisite connection
│   │   ├── Solid: Direct prerequisite
│   │   └── Dashed: Optional/unlocks
│   └── MiniMap (TextureRect) - Overview
├── Panel (PanelContainer) - Details
│   ├── Title
│   ├── Description
│   ├── Difficulty stars
│   ├── Mode badges [A][G][S]
│   └── Action: "Enter Challenge"
└── Filters (HBoxContainer)
    ├── Era tabs
    ├── Mode filter
    └── Search box
```

### GraphView Drawing
```gdscript
# GraphView.gd
extends Control
class_name GraphView

var nodes: Dictionary = {}  # id -> NodeData
var edges: Array = []       # {from, to, type}

func _draw():
    # Draw edges first (behind nodes)
    for edge in edges:
        var from_pos = nodes[edge.from].position
        var to_pos = nodes[edge.to].position
        var color = edge.type == "prereq" ? Color.WHITE : Color.YELLOW
        var style = edge.type == "prereq" ? LINE_SOLID : LINE_DASHED
        draw_line(from_pos, to_pos, color, 2, style)
    
    # Draw nodes
    for node in nodes.values():
        var color = node.get_state_color()
        draw_circle(node.position, 20, color)
        if node.state == NodeState.AVAILABLE:
            draw_circle(node.position, 24, Color.YELLOW, false, 2)

func get_node_at(pos: Vector2) -> String:
    for id, node in nodes:
        if pos.distance_to(node.position) < 24:
            return id
    return ""
```

---

## 4. Era Theme System

### From: `godot-coding-challenge` (per-project themes) + `dothop` (themes folder)

```
ThemeManager (Autoload)
├── current_theme: Theme
├── era_themes: Dictionary
│   ├── magitech: Theme
│   │   ├── colors: parchment, ink, gold, blood
│   │   ├── fonts: medieval, runic
│   │   ├── textures: paper, vellum, wax_seal
│   │   └── sounds: quill, candle, chant
│   ├── cyberpunk: Theme
│   │   ├── colors: neon_green, hot_pink, deep_blue, black
│   │   ├── fonts: monospace, terminal
│   │   ├── textures: scanlines, glitch, circuit
│   │   └── sounds: synth, keyboard, hum
│   └── ...
└── apply_theme(era: String)
```

```gdscript
# ThemeManager.gd
extends Node
class_name ThemeManager

var themes: Dictionary = {}

func _ready():
    themes["magitech"] = load("res://themes/magitech/theme.tres")
    themes["cyberpunk"] = load("res://themes/cyberpunk/theme.tres")
    # ...
    apply_theme("magitech")

func apply_theme(era: String):
    if themes.has(era):
        current_theme = themes[era]
        # Apply to entire scene tree
        for canvas in get_viewport().canvas_layers:
            apply_theme_recursive(canvas, current_theme)

func apply_theme_recursive(node: Node, theme: Theme):
    if node is Control:
        node.theme = theme
    for child in node.get_children():
        apply_theme_recursive(child, theme)
```

### Theme.tres Structure
```gdscript
# magitech/theme.tres
[gd_resource type="Theme" load_steps=4 format=3]

[sub_resource type="StyleBoxFlat" id="StyleBoxFlat_panel"]
bg_color = Color(0.12, 0.08, 0.04, 1)
border_width_top = 2
border_width_bottom = 2
border_width_left = 2
border_width_right = 2
border_color = Color(0.6, 0.4, 0.1, 1)
corner_radius_top_left = 4
corner_radius_top_right = 4
corner_radius_bottom_left = 4
corner_radius_bottom_right = 4

[sub_resource type="FontFile" id="FontFile_main"]
data = ExtResource("res://fonts/medieval.ttf")

[resource]
default_font = SubResource("FontFile_main")
default_font_size = 16
default_font_color = Color(0.9, 0.8, 0.6, 1)

colors = {
    "font_color": Color(0.9, 0.8, 0.6, 1),
    "font_color_hover": Color(1, 0.9, 0.4, 1),
    "font_color_pressed": Color(1, 0.7, 0.2, 1),
    "font_color_disabled": Color(0.4, 0.3, 0.2, 1),
    "button_normal": Color(0.15, 0.1, 0.05, 1),
    "button_hover": Color(0.2, 0.15, 0.08, 1),
    "button_pressed": Color(0.25, 0.2, 0.1, 1),
}
styles = {
    "panel": SubResource("StyleBoxFlat_panel"),
}
constants = {
    "separation": 8,
}
```

---

## 5. Hint / Archive Panel

### From: `CHALLENGE-TO-YOU-PLAN.md` + `godot_recipes` (UI patterns)

```
HintArchive (PopupPanel)
├── Header (HBoxContainer)
│   ├── Title: "Hint Archive"
│   ├── Search box
│   └── Close button
├── CategoryTree (Tree)
│   ├── Era: Magitech
│   │   ├── Tier 1: Rune Binding
│   │   │   ├── [🔒] Hint 1: "Mana flows downhill"
│   │   │   └── [🔓] Hint 2: "Bind to the well first"
│   │   └── Tier 2: Sigil Chains
│   └── Era: Cyberpunk
├── DetailPanel (VBoxContainer) - Selected hint
│   ├── Hint text (RichTextLabel)
│   ├── Tier indicator (1-3 stars)
│   ├── Cost: "5 Luck" or "Free"
│   └── Unlock button
└── Footer
    └── Total hints: 23/47
```

---

## 6. Passcode Display

```
PasscodeDisplay (HBoxContainer)
├── Label: "PASSCODE:"
├── Code (RichTextLabel) - Monospace, animated reveal
│   "LOGOS-AXIOM-PRIME"
├── Copy Button (Button)
├── Source Badge (TextureRect) - How it was found
│   [Glitch] [Error] [Timing] [Direct] [Chain]
└── Animation Player - Typewriter effect
```

```gdscript
# PasscodeDisplay.gd
func show_passcode(code: String, source: PasscodeSource):
    visible = true
    $Code.text = ""
    $Code.add_text("[color=#00ff00][font=monospace]")
    
    var tween = create_tween()
    for char in code:
        tween.tween_callback(Callable(self, "_append_char").bind(char))
        tween.tween_interval(0.05)
    tween.tween_callback(Callable(self, "_append_char").bind("[/font][/color]"))
    
    $SourceBadge.texture = source_icons[source]

func _append_char(char: String):
    $Code.append_text(char)
```

---

## 7. Save/Load System

### From: `dothop` (SaveGame.gd) + `godot_recipes`

```
SaveManager (Autoload)
├── save_slots: Array[SaveSlot]
├── current_slot: int
├── auto_save_interval: 60.0
├── save(path) -> Error
├── load(path) -> SaveData
├── auto_save()
└── quick_save/quick_load (F5/F9)
```

```gdscript
# SaveGame.gd (from dothop)
class_name SaveGame
extends Resource

var version: int = 1
var timestamp: int
var player_data: Dictionary = {
    "era": "magitech",
    "luck": 0.5,
    "completed_puzzles": [],
    "hint_archive": {},
    "statistics": {}
}
var world_state: Dictionary = {}
var rng_state: Array = []  # For reproducibility
```

---

## 8. Accessibility Patterns

### From: `godot_recipes` (UI basics)

| Feature | Implementation |
|---------|----------------|
| High contrast | Theme variant `theme_high_contrast.tres` |
| Font scaling | `@export var font_scale: float = 1.0` applied to all controls |
| Screen reader | `Control.accessibility_*` properties |
| Color blind | Avoid red/green only; use patterns + color |
| Key remapping | `InputMap` customization menu |
| Reduced motion | `@export var reduced_motion: bool` disables tweens |

---

## 9. Responsive Layout

```gdscript
# ResponsiveContainer.gd
extends Container
class_name ResponsiveContainer

@export var breakpoints: Dictionary = {
    "mobile": 600,
    "tablet": 1024,
    "desktop": 1920
}

func _notification(what):
    if what == NOTIFICATION_RESIZED:
        _update_layout()

func _update_layout():
    var width = get_viewport_rect().size.x
    if width < breakpoints["mobile"]:
        _apply_layout("mobile")
    elif width < breakpoints["tablet"]:
        _apply_layout("tablet")
    else:
        _apply_layout("desktop")

func _apply_layout(layout: String):
    # Reorganize children based on layout
    match layout:
        "mobile":
            # Stack vertically, hide sidebar
        "tablet":
            # Two column
        "desktop":
            # Three column with full graph
```

---

## 10. Animation Library

```gdscript
# UIAnimations.gd (Autoload)
static func fade_in(node: Control, duration: float = 0.2) -> Tween:
    node.modulate = Color(1, 1, 1, 0)
    node.show()
    var tween = node.create_tween()
    tween.tween_property(node, "modulate:a", 1.0, duration)
    return tween

static func fade_out(node: Control, duration: float = 0.2) -> Tween:
    var tween = node.create_tween()
    tween.tween_property(node, "modulate:a", 0.0, duration)
    tween.finished.connect(Callable(node, "hide"))
    return tween

static func slide_in(node: Control, from: Vector2, duration: float = 0.3) -> Tween:
    node.position = from
    var tween = node.create_tween()
    tween.tween_property(node, "position", Vector2.ZERO, duration).set_trans(Tween.TRANS_QUAD).set_ease(Tween.EASE_OUT)
    return tween

static func shake(node: Control, intensity: float = 5.0, duration: float = 0.3) -> Tween:
    var tween = node.create_tween()
    var original = node.position
    for i in range(6):
        tween.tween_property(node, "position", original + Vector2(randf_range(-intensity, intensity), 0), duration/6)
    tween.tween_property(node, "position", original, duration/6)
    return tween

static func pulse(node: Control, scale: float = 1.1, duration: float = 0.5) -> Tween:
    var tween = node.create_tween()
    tween.set_loops()
    tween.tween_property(node, "scale", Vector2(scale, scale), duration/2).set_trans(Tween.TRANS_SINE)
    tween.tween_property(node, "scale", Vector2(1, 1), duration/2).set_trans(Tween.TRANS_SINE)
    return tween
```

---

## 11. Debug Overlay (Dev Only)

```
DebugOverlay (CanvasLayer - top)
├── FPS Counter (Label)
├── Memory Usage (Label)
├── VM Step Counter (Label)
├── Vigilance Debug (Graph)
├── Luck Value (Label)
├── RNG Seed (Label)
├── Active Challenge (Label)
├── Hotkeys:
│   F1: Toggle debug
│   F2: Step VM
│   F3: Dump bytecode
│   F4: Force passcode
│   F5: Quick save
│   F9: Quick load
└── Console (TextEdit - read only)
    └── Log output
```