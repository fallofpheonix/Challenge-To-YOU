extends Control

signal play_pressed
signal pause_pressed
signal step_forward_pressed
signal step_backward_pressed
signal reset_pressed
signal seek_requested(tick: int)

var replay_data: Dictionary = {}
var is_playing: bool = false
var current_tick: int = 0
var total_ticks: int = 0

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var tick_label: Label = $VBoxContainer/TimelineBar/HBoxContainer/TickLabel
@onready var seek_slider: HSlider = $VBoxContainer/TimelineBar/HBoxContainer/SeekSlider
@onready var speed_label: Label = $VBoxContainer/ControlsBar/SpeedLabel
@onready var play_btn: Button = $VBoxContainer/ControlsBar/PlayBtn
@onready var event_log: VBoxContainer = $VBoxContainer/ScrollContainer/EventLog

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/TimelineBar, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)
	seek_slider.min_value = 0
	seek_slider.max_value = 1
	seek_slider.step = 0.001
	seek_slider.value = 0

func load_replay(data: Dictionary) -> void:
	replay_data = data
	total_ticks = data.get("total_ticks", 0)
	current_tick = data.get("current_tick", 0)
	seek_slider.max_value = max(total_ticks, 1)
	seek_slider.value = current_tick
	_update_tick_display()
	_populate_event_log()

func _update_tick_display() -> void:
	tick_label.text = "Tick: %d / %d" % [current_tick, total_ticks]
	seek_slider.value = current_tick

func _populate_event_log() -> void:
	for child in event_log.get_children():
		child.queue_free()

	var events = replay_data.get("events", [])
	for e in events:
		var row = preload("res://ui/components/entity_row.tscn").instantiate()
		var event_type = e.get("type", "unknown")
		var color = _event_color(event_type)
		row.set_data(
			"@ Tick %d" % e.get("tick", 0),
			_event_icon(event_type),
			color,
			e.get("description", "")
		)
		event_log.add_child(row)

func _event_color(event_type: String) -> Color:
	match event_type:
		"drone_spawned": return Color(0, 1, 0.62, 1)
		"drone_died": return Color(1, 0.2, 0.2, 1)
		"resource_depleted": return Color(1, 0.75, 0, 1)
		"structure_built": return Color(0.2, 0.4, 1, 1)
		"hazard_triggered": return Color(1, 0.4, 0, 1)
		"alien_activity": return Color(1, 0, 0.5, 1)
		_: return Color(0.5, 0.5, 0.5, 1)

func _event_icon(event_type: String) -> String:
	match event_type:
		"drone_spawned": return "⊕"
		"drone_died": return "⊗"
		"resource_depleted": return "∅"
		"structure_built": return "▲"
		"hazard_triggered": return "⚡"
		"alien_activity": return "◆"
		_: return "●"

func _on_play_pressed() -> void:
	is_playing = not is_playing
	play_btn.text = "⏸" if is_playing else "▶"
	if is_playing:
		play_pressed.emit()
	else:
		pause_pressed.emit()

func _on_step_forward_pressed() -> void:
	step_forward_pressed.emit()

func _on_step_back_pressed() -> void:
	step_backward_pressed.emit()

func _on_reset_pressed() -> void:
	reset_pressed.emit()
	current_tick = 0
	_update_tick_display()

func _on_seek_slider_changed(value: float) -> void:
	current_tick = int(value)
	_update_tick_display()
	seek_requested.emit(current_tick)

func _on_speed_up_pressed() -> void:
	var speed = replay_data.get("speed", 1.0)
	speed = min(speed * 2, 64.0)
	replay_data["speed"] = speed
	speed_label.text = "%.1fx" % speed

func _on_speed_down_pressed() -> void:
	var speed = replay_data.get("speed", 1.0)
	speed = max(speed / 2, 0.25)
	replay_data["speed"] = speed
	speed_label.text = "%.1fx" % speed
