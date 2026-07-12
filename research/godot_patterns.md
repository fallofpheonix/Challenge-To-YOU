# Godot Patterns

Patterns extracted from `godot_recipes`, `godot-coding-challenge`, `dothop`, and official demos.

---

## 1. Scene Architecture

### Composition over Inheritance
```gdscript
# Good: Compose behaviors
class_name Player
extends CharacterBody2D

@export var movement: MovementComponent
@export var combat: CombatComponent
@export var inventory: InventoryComponent

func _physics_process(delta):
    movement.process(delta)
    combat.process(delta)

# Avoid: Deep inheritance chains
# Player -> Character -> Entity -> Node
```

### Scene Instancing Pattern
```gdscript
# Factory for consistent instantiation
class_name EntityFactory
extends Node

@export var enemy_scenes: Array[PackedScene] = []

static func spawn_enemy(type: String, position: Vector2, parent: Node) -> CharacterBody2D:
    var scene = preload("res://enemies/" + type + ".tscn")
    var instance = scene.instantiate()
    instance.global_position = position
    parent.add_child(instance)
    return instance
```

### Autoload Singletons (Global Systems)
```gdscript
# Store.gd - Global state (from dothop)
class_name Store
extends Node

signal data_changed(key: String)

var _data: Dictionary = {}

func set(key: String, value: Variant):
    _data[key] = value
    emit_signal("data_changed", key)

func get(key: String, default: Variant = null) -> Variant:
    return _data.get(key, default)

func has(key: String) -> bool:
    return _data.has(key)

# Events.gd - Signal bus (from dothop)
class_name Events
extends Node

signal puzzle_completed(puzzle_id: String)
signal era_unlocked(era_id: String)
signal vigilance_changed(level: float)

static var instance: Events

func _ready():
    Events.instance = self
```

---

## 2. Node Patterns

### Controller Pattern (Input Separation)
```gdscript
# PlayerController.gd
class_name PlayerController
extends Node

@export var move_speed: float = 300.0
@export var dash_speed: float = 800.0

var input_vector: Vector2 = Vector2.ZERO
var dash_cooldown: float = 0.0

func _process(delta: float):
    input_vector = Input.get_vector("move_left", "move_right", "move_up", "move_down")
    if Input.is_action_just_pressed("dash") and dash_cooldown <= 0:
        dash_cooldown = 1.0
        emit_signal("dash_requested", input_vector)

func _physics_process(delta: float):
    dash_cooldown = max(0.0, dash_cooldown - delta)
```

### State Machine (from godot_recipes/ai/)
```gdscript
# StateMachine.gd
class_name StateMachine
extends Node

var current_state: State
var states: Dictionary = {}

func add_state(name: String, state: State):
    states[name] = state
    state.fsm = self

func transition_to(state_name: String):
    if current_state:
        current_state.exit()
    current_state = states.get(state_name)
    if current_state:
        current_state.enter()

func _process(delta: float):
    if current_state:
        current_state.update(delta)

# State.gd
class_name State
extends Resource

@export var fsm: StateMachine

func enter(): pass
func update(delta: float): pass
func exit(): pass

# Usage
var idle = IdleState.new()
var run = RunState.new()
var jump = JumpState.new()

state_machine.add_state("idle", idle)
state_machine.add_state("run", run)
state_machine.add_state("jump", jump)
state_machine.transition_to("idle")
```

### Object Pooling (Performance)
```gdscript
# ObjectPool.gd
class_name ObjectPool
extends Node

@export var scene: PackedScene
@export var initial_size: int = 10
@export var max_size: int = 100

var _available: Array = []
var _in_use: Array = []

func _ready():
    for i in range(initial_size):
        _available.append(_create_instance())

func acquire(position: Vector2 = Vector2.ZERO) -> Node:
    var instance: Node
    if _available.size() > 0:
        instance = _available.pop_back()
    else:
        instance = _create_instance()
    _in_use.append(instance)
    instance.global_position = position
    instance.show()
    return instance

func release(instance: Node):
    if instance in _in_use:
        _in_use.erase(instance)
        instance.hide()
        if _available.size() < max_size:
            _available.append(instance)
        else:
            instance.queue_free()

func _create_instance() -> Node:
    var inst = scene.instantiate()
    inst.hide()
    add_child(inst)
    return inst
```

---

## 3. Resource System

### Custom Resources (Data-Driven)
```gdscript
# PuzzleData.gd
class_name PuzzleData
extends Resource

@export var id: String
@export var title: String
@export var description: String
@export var era: String
@export var difficulty: int = 1
@export var puzzle_type: String  # "code", "grid", "logic"
@export var prerequisites: Array[String] = []
@export var unlocks: Array[String] = []

# Type-specific data
@export var starter_code: String = ""
@export var test_cases: Array[Dictionary] = []
@export var grid_data: Dictionary = {}
@export var logic_rules: Array[Dictionary] = []

@export var hints: Array[HintData] = []

# HintData.gd
class_name HintData
extends Resource

@export var tier: int = 1
@export var text: String
@export var unlock_condition: String  # "attempts >= 3"
@export var cost: int = 0  # Luck points
```

### Resource Loading
```gdscript
# PuzzleLoader.gd
class_name PuzzleLoader
extends Node

static const PUZZLE_PATH = "res://data/puzzles/"

static func load_puzzle(id: String) -> PuzzleData:
    var path = PUZZLE_PATH + id + ".tres"
    var resource = ResourceLoader.load(path)
    if resource == null:
        push_error("Failed to load puzzle: ", path)
        return PuzzleData.new()
    return resource

static func load_era_pack(era: String) -> Array[PuzzleData]:
    var dir = DirAccess.open(PUZZLE_PATH + era)
    if dir == null: return []
    
    var puzzles = []
    for file in dir.get_files():
        if file.ends_with(".tres"):
            puzzles.append(load_puzzle(file.get_basename()))
    return puzzles
```

---

## 4. UI Patterns

### Responsive UI with Containers
```gdscript
# HBoxContainer / VBoxContainer / GridContainer
# MarginContainer for padding
# CenterContainer for centering

# CodeEditor layout example:
# VBoxContainer (main)
#   ├── HBoxContainer (toolbar)
#   │   ├── Button (run)
#   │   ├── Button (stop)
#   │   ├── Label (status)
#   │   └── HSpacer
#   ├── MarginContainer (editor area, margin: 4)
#   │   └── TextEdit (code_input)
#   └── PanelContainer (output)
#       ├── Label (output_header)
#       └── TextEdit (output_log, readonly)
```

### Theme Override (Runtime)
```gdscript
# ThemeManager.gd
func apply_era_theme(era: String):
    var theme = load("res://themes/" + era + ".tres")
    for control in get_tree().get_nodes_in_group("themed"):
        if control is Control:
            control.theme = theme
    # Force refresh
    for control in get_tree().get_nodes_in_group("themed"):
        if control is Control:
            control.notify_theme_changed()
```

### Custom TextEdit Syntax Highlighting
```gdscript
# SyntaxHighlighter.gd
class_name SyntaxHighlighter
extends Resource

@export var keywords: Array[String] = []
@export var types: Array[String] = []
@export var functions: Array[String] = []
@export var color_keyword: Color = Color(0.8, 0.4, 1.0)
@export var color_type: Color = Color(0.4, 0.8, 1.0)
@export var color_function: Color = Color(0.4, 1.0, 0.6)
@export var color_string: Color = Color(1.0, 0.8, 0.3)
@export var color_number: Color = Color(1.0, 0.5, 0.5)
@export var color_comment: Color = Color(0.5, 0.5, 0.5)

func highlight(editor: TextEdit):
    var text = editor.text
    var lines = text.split("\n")
    
    for line_idx, line in lines:
        _highlight_line(editor, line_idx, line)

func _highlight_line(editor: TextEdit, line: int, text: String):
    # Simple regex-based highlighting
    # Production: use TreeSitter via GDExtension
    var regex = RegEx.new()
    
    # Keywords
    regex.compile("\\b(" + String(", ").join(keywords) + ")\\b")
    _apply_matches(editor, line, regex, color_keyword)
    
    # Strings
    regex.compile("\"[^\"]*\"")
    _apply_matches(editor, line, regex, color_string)
    
    # Numbers
    regex.compile("\\b\\d+(\\.\\d+)?\\b")
    _apply_matches(editor, line, regex, color_number)
    
    # Comments
    regex.compile("//.*$")
    _apply_matches(editor, line, regex, color_comment)

func _apply_matches(editor: TextEdit, line: int, regex: RegEx, color: Color):
    var match = regex.search(text)
    while match:
        var start = match.get_subject().find(match.get_string(0))
        var end = start + match.get_string(0).length()
        editor.set_syntax_highlighting(line, start, end, color)
        match = regex.search(text, end)
```

---

## 5. Animation & Tweening

### Godot 4 Tween API
```gdscript
# Smooth movement
func move_to(target: Vector2, duration: float = 0.3) -> Tween:
    var tween = create_tween()
    tween.set_trans(Tween.TRANS_QUAD)
    tween.set_ease(Tween.EASE_OUT)
    tween.tween_property(self, "global_position", target, duration)
    return tween

# Pulse animation
func pulse(scale: float = 1.2, duration: float = 0.5) -> Tween:
    var tween = create_tween()
    tween.set_loops()
    tween.tween_property(self, "scale", Vector2(scale, scale), duration/2)
    tween.tween_property(self, "scale", Vector2(1, 1), duration/2)
    return tween

# Shake effect
func shake(intensity: float = 5.0, duration: float = 0.3) -> Tween:
    var tween = create_tween()
    var original = position
    for i in 6:
        tween.tween_property(self, "position", original + Vector2(randf_range(-intensity, intensity), 0), duration/6)
    tween.tween_property(self, "position", original, duration/6)
    return tween

# Parallel animations
func parallel_animations():
    var tween = create_tween()
    tween.parallel().tween_property(sprite, "modulate", Color(1,0,0,1), 0.2)
    tween.parallel().tween_property(sprite, "scale", Vector2(1.2, 1.2), 0.2)
    tween.tween_property(sprite, "modulate", Color(1,1,1,1), 0.2)
    tween.tween_property(sprite, "scale", Vector2(1, 1), 0.2)
```

---

## 6. Signal Patterns

### Typed Signals (Godot 4)
```gdscript
# In class definition
signal puzzle_completed(puzzle_id: String, solution: Variant, time_ms: int)
signal vigilance_changed(level: float, threshold_hit: String)

# Emitting
puzzle_completed.emit("rune_01", my_solution, 1250)
vigilance_changed.emit(0.75, "scan")

# Connecting with Callable
puzzle_completed.connect(_on_puzzle_completed.bind(extra_data))

func _on_puzzle_completed(puzzle_id: String, solution: Variant, time_ms: int, extra: String):
    pass
```

### Signal Bus (Decoupled Communication)
```gdscript
# SignalBus.gd - Global event system
class_name SignalBus
extends Node

# Gameplay
signal puzzle_started(puzzle_id: String)
signal puzzle_completed(puzzle_id: String, result: Dictionary)
signal puzzle_failed(puzzle_id: String, error: String)

# Progression
signal era_unlocked(era: String)
signal mode_unlocked(mode: String)
signal passcode_found(passcode: String, source: String)

# UI
signal show_notification(title: String, body: String, type: int)
signal update_hud(data: Dictionary)

# System
signal save_requested()
signal load_requested(slot: int)
signal settings_changed(key: String, value: Variant)

static var instance: SignalBus

func _ready():
    SignalBus.instance = self
```

---

## 7. Input Handling

### Input Map Actions (Project Settings)
```ini
# project.godot [input]
move_left = {"deadzone": 0.5, "events": [Key(A), Key(Left)]}
move_right = {"deadzone": 0.5, "events": [Key(D), Key(Right)]}
move_up = {"deadzone": 0.5, "events": [Key(W), Key(Up)]}
move_down = {"deadzone": 0.5, "events": [Key(S), Key(Down)]}
dash = {"deadzone": 0.5, "events": [Key(Shift), Key(Space)]}
interact = {"deadzone": 0.5, "events": [Key(E)]}
pause = {"deadzone": 0.5, "events": [Key(Escape)]}
```

### Input Processing
```gdscript
# InputHandler.gd
class_name InputHandler
extends Node

func _unhandled_input(event: InputEvent):
    if event is InputEventKey and event.pressed:
        match event.keycode:
            KEY_F1: SignalBus.instance.show_notification.emit("Debug", "F1 pressed", 0)
            KEY_F2: _dump_debug_info()
            KEY_F3: _reload_current_scene()

func _process(delta):
    # Continuous input
    var move = Input.get_vector("move_left", "move_right", "move_up", "move_down")
    if move.length() > 0:
        SignalBus.instance.player_moved.emit(move)
```

---

## 8. Save/Load System

### Binary Save (Fast)
```gdscript
# SaveSystem.gd
class_name SaveSystem
extends Node

static const SAVE_PATH = "user://saves/"

static func save(slot: int, data: Dictionary) -> Error:
    var file = FileAccess.open(SAVE_PATH + "slot_%d.save" % slot, FileAccess.WRITE)
    if file == null: return ERR_FILE_CANT_OPEN
    
    # Version header
    file.store_32(1)  # version
    file.store_64(Time.get_unix_time_from_system())  # timestamp
    
    # Compressed JSON
    var json = JSON.stringify(data)
    var compressed = json.to_utf8_buffer().compress(FileAccess.COMPRESSION_ZSTD)
    file.store_32(compressed.size())
    file.store_buffer(compressed)
    
    return OK

static func load(slot: int) -> Dictionary:
    var file = FileAccess.open(SAVE_PATH + "slot_%d.save" % slot, FileAccess.READ)
    if file == null: return {}
    
    var version = file.get_32()
    var timestamp = file.get_64()
    var data_size = file.get_32()
    var compressed = file.get_buffer(data_size)
    
    var json = compressed.decompress(FileAccess.COMPRESSION_ZSTD).get_string_from_utf8()
    return JSON.parse_string(json)
```

---

## 9. Debug Overlay

```gdscript
# DebugOverlay.gd (CanvasLayer, top level)
class_name DebugOverlay
extends CanvasLayer

@onready var fps_label = $FPSLabel
@onready var mem_label = $MemLabel
@onready var vm_label = $VMLabel
@onready var vigilance_graph = $VigilanceGraph
@onready var console = $Console

func _process(delta):
    if not visible: return
    fps_label.text = "FPS: %d" % Engine.get_frames_per_second()
    mem_label.text = "Mem: %.1f MB" % (OS.get_static_memory_usage() / 1024.0 / 1024.0)
    if Engine.has_singleton("VM"):
        vm_label.text = "VM Steps: %d" % Engine.get_singleton("VM").get_step_count()

func _unhandled_input(event):
    if event is InputEventKey and event.pressed:
        match event.keycode:
            KEY_F1: visible = not visible
            KEY_F2: _step_vm()
            KEY_F3: _dump_bytecode()
            KEY_F4: _force_passcode()
            KEY_F5: SignalBus.instance.save_requested.emit()
            KEY_F9: SignalBus.instance.load_requested.emit(0)

func log(msg: String):
    console.append_text(msg + "\n")
    console.scroll_to_bottom()
```

---

## 10. Testing Patterns

### GUT (Godot Unit Test)
```gdscript
# test_puzzle_board.gd
extends GutTest

func before_each():
    board = Board.new()
    board.add_thing(Thing.new("a"))
    board.add_thing(Thing.new("b", ["a"]))

func test_dependency_unlock():
    board.get_thing("a").complete()
    assert_true(board.get_thing("b").state == Thing.State.AVAILABLE)

func test_cycle_detection():
    board.add_thing(Thing.new("c", ["d"]))
    board.add_thing(Thing.new("d", ["c"]))
    assert_false(board.validate_graph())
```

### Integration Test (Scene)
```gdscript
# test_vm_integration.gd
extends GutTest

func test_vm_execution():
    var vm = VM.new()
    vm.set_emit_callback(_on_emit.bind())
    
    var source = """
    let x = 10
    let y = 20
    emit(x + y)
    """
    
    var result = vm.execute(source)
    assert_true(result.success)
    assert_eq(last_emitted, "30")

var last_emitted: String = ""

func _on_emit(value: String):
    last_emitted = value
```