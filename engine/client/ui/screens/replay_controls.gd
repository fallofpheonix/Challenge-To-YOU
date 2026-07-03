extends Control

var is_playing: bool = false
var current_tick: int = 0
var total_ticks: int = 0
var is_recording: bool = false

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var tick_label: Label = $VBoxContainer/TimelineBar/HBoxContainer/TickLabel
@onready var seek_slider: HSlider = $VBoxContainer/TimelineBar/HBoxContainer/SeekSlider
@onready var speed_label: Label = $VBoxContainer/ControlsBar/SpeedLabel
@onready var play_btn: Button = $VBoxContainer/ControlsBar/PlayBtn
@onready var event_log: VBoxContainer = $VBoxContainer/ScrollContainer/EventLog

var _speed: float = 1.0

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/TimelineBar, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)
	seek_slider.min_value = 0
	seek_slider.max_value = 1
	seek_slider.step = 1
	seek_slider.value = 0

# Called by game_hub.gd every tick with data["replay"] from EMISSION_SNAPSHOT.
# During live simulation: {recording, total_ticks, current_tick}
# After REPLAY_SEEK:      {recording:false, total_ticks, current_tick, seek_to, events:[...]}
func load_replay(data: Dictionary) -> void:
	is_recording = data.get("recording", false)
	var total = int(data.get("total_ticks", 0))
	if total > 0:
		total_ticks = total
		seek_slider.max_value = float(total)

	current_tick = int(data.get("current_tick", current_tick))
	_update_tick_display()

	if data.has("events"):
		_populate_event_log(data["events"])

func _update_tick_display() -> void:
	var rec_tag = " [REC]" if is_recording else " [REPLAY]"
	tick_label.text = "Tick: %d / %d%s" % [current_tick, total_ticks, rec_tag]
	seek_slider.set_block_signals(true)
	seek_slider.value = float(current_tick)
	seek_slider.set_block_signals(false)

func _populate_event_log(events: Array) -> void:
	for child in event_log.get_children():
		child.queue_free()

	# Events arrive oldest-first; show newest first in the log
	var sorted = events.duplicate()
	sorted.reverse()

	for e in sorted:
		var row = preload("res://ui/components/entity_row.tscn").instantiate()
		var event_type = e.get("type", "unknown")
		var color = _event_color(event_type)
		var drone_tag = "" if e.get("drone_id", -1) < 0 else " #%d" % e.get("drone_id", 0)
		row.set_data(
			"T%d%s" % [e.get("tick", 0), drone_tag],
			_event_icon(event_type),
			color,
			event_type.replace("_", " ").to_upper()
		)
		event_log.add_child(row)

func _get_net() -> Node:
	return get_tree().root.find_child("NetworkBridge", true, false)

func _on_play_pressed() -> void:
	is_playing = not is_playing
	play_btn.text = "⏸" if is_playing else "▶"

func _on_step_forward_pressed() -> void:
	var target = current_tick + 1
	_seek_to(target)

func _on_step_back_pressed() -> void:
	var target = max(0, current_tick - 1)
	_seek_to(target)

func _on_reset_pressed() -> void:
	_seek_to(0)

func _on_seek_slider_changed(value: float) -> void:
	var target = int(value)
	if target != current_tick:
		_seek_to(target)

func _on_save_pressed() -> void:
	var net = _get_net()
	if net:
		net.send_command("REPLAY_SAVE", {})

func _seek_to(tick: int) -> void:
	current_tick = tick
	_update_tick_display()
	var net = _get_net()
	if net:
		net.send_command("REPLAY_SEEK", {"tick": tick})

func _on_speed_up_pressed() -> void:
	_speed = min(_speed * 2.0, 64.0)
	speed_label.text = "%.1fx" % _speed

func _on_speed_down_pressed() -> void:
	_speed = max(_speed / 2.0, 0.25)
	speed_label.text = "%.1fx" % _speed

func _event_color(event_type: String) -> Color:
	match event_type:
		"drone_spawned", "fabricated": return Color(0, 1, 0.62, 1)
		"drone_died":                  return Color(1, 0.2, 0.2, 1)
		"harvested":                   return Color(1, 0.75, 0, 1)
		"deposited":                   return Color(0.2, 0.4, 1, 1)
		"drone_infected":              return Color(0.8, 0, 1, 1)
		"hazard_damage":               return Color(1, 0.4, 0, 1)
		"trust_changed":               return Color(0.5, 0.5, 1, 1)
		"mission_changed":             return Color(1, 1, 0.2, 1)
		_:                             return Color(0.6, 0.6, 0.6, 1)

func _event_icon(event_type: String) -> String:
	match event_type:
		"drone_spawned", "fabricated": return "⊕"
		"drone_died":                  return "⊗"
		"harvested":                   return "◆"
		"deposited":                   return "▲"
		"drone_infected":              return "◆"
		"hazard_damage":               return "⚡"
		"trust_changed":               return "~"
		"mission_changed":             return "★"
		_:                             return "●"
