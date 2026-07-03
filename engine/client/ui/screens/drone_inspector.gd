extends Control

signal view_swarm_requested
signal view_entity_requested(entity_id: int)

var swarm_registry: Dictionary = {}
var selected_drone_id: int = -1
var _recent_events: Array = []
const MAX_EVENTS_SHOWN = 50

var _detail_vbox: VBoxContainer
var _trace_list: VBoxContainer
var _events_list: VBoxContainer

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var drone_count_label: Label = $VBoxContainer/SummaryBar/HBoxContainer/DroneCount
@onready var total_battery_label: Label = $VBoxContainer/SummaryBar/HBoxContainer/TotalBattery
@onready var active_count_label: Label = $VBoxContainer/SummaryBar/HBoxContainer/ActiveCount
@onready var drone_list: VBoxContainer = $VBoxContainer/ScrollContainer/DroneList
@onready var detail_panel: PanelContainer = $VBoxContainer/DetailPanel
@onready var detail_id: Label = $VBoxContainer/DetailPanel/MarginContainer/VBoxContainer/DetailId
@onready var detail_pos: Label = $VBoxContainer/DetailPanel/MarginContainer/VBoxContainer/DetailPos
@onready var detail_battery: Label = $VBoxContainer/DetailPanel/MarginContainer/VBoxContainer/DetailBattery
@onready var detail_state: Label = $VBoxContainer/DetailPanel/MarginContainer/VBoxContainer/DetailState
@onready var detail_payload: Label = $VBoxContainer/DetailPanel/MarginContainer/VBoxContainer/DetailPayload
@onready var detail_protocol: Label = $VBoxContainer/DetailPanel/MarginContainer/VBoxContainer/DetailProtocol

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/SummaryBar, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)
	theme_ctrl.apply_panel_style(detail_panel, theme_ctrl.colors.surface, theme_ctrl.colors.border)
	detail_panel.hide()
	_detail_vbox = $VBoxContainer/DetailPanel/MarginContainer/VBoxContainer
	_build_inspector_panels()

func _build_inspector_panels() -> void:
	var sep1 = HSeparator.new()
	_detail_vbox.add_child(sep1)

	var trace_header = Label.new()
	trace_header.text = "DECISION TRACE"
	trace_header.add_theme_font_size_override("font_size", 10)
	trace_header.add_theme_color_override("font_color", Color(0, 1, 1, 1))
	_detail_vbox.add_child(trace_header)

	var trace_scroll = ScrollContainer.new()
	trace_scroll.custom_minimum_size = Vector2(0, 120)
	trace_scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	_detail_vbox.add_child(trace_scroll)

	_trace_list = VBoxContainer.new()
	_trace_list.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	trace_scroll.add_child(_trace_list)

	var sep2 = HSeparator.new()
	_detail_vbox.add_child(sep2)

	var events_header = Label.new()
	events_header.text = "EVENT TIMELINE"
	events_header.add_theme_font_size_override("font_size", 10)
	events_header.add_theme_color_override("font_color", Color(0, 1, 1, 1))
	_detail_vbox.add_child(events_header)

	var events_scroll = ScrollContainer.new()
	events_scroll.custom_minimum_size = Vector2(0, 100)
	_detail_vbox.add_child(events_scroll)

	_events_list = VBoxContainer.new()
	_events_list.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	events_scroll.add_child(_events_list)

func load_swarm(data: Dictionary) -> void:
	swarm_registry = data
	_update_summary()
	_populate_list()

func _update_summary() -> void:
	var drones = swarm_registry.get("drones", [])
	var total = drones.size()
	var active = 0
	var total_bat = 0
	for d in drones:
		if d.get("state", "inert") != "inert":
			active += 1
		total_bat += d.get("bat", 0)

	drone_count_label.text = "Drones: %d" % total
	active_count_label.text = "Active: %d" % active
	total_battery_label.text = "Battery: %d" % total_bat

func _populate_list() -> void:
	for child in drone_list.get_children():
		child.queue_free()

	for d in swarm_registry.get("drones", []):
		var row = preload("res://ui/components/entity_row.tscn").instantiate()
		var state_idx = d.get("state", 2)
		var is_comp = d.get("comp", false)

		var color = _state_color(state_idx, is_comp)
		var x = d.get("x", 0) / 1000000
		var y = d.get("y", 0) / 1000000

		var is_selected = d.get("id", -1) == selected_drone_id
		if is_selected:
			color = color.lightened(0.3)

		row.set_data(
			"Drone #%d" % d.get("id", 0),
			_state_name(state_idx),
			color,
			"(%d, %d)" % [x, y]
		)
		var drone_id = d.get("id", -1)
		row.clicked.connect(func(): _select_drone(drone_id))
		drone_list.add_child(row)

func _select_drone(drone_id: int) -> void:
	selected_drone_id = drone_id
	_recent_events.clear()
	_clear_trace()

	var d = _find_drone(drone_id)
	if d.is_empty():
		return

	detail_panel.show()
	detail_id.text = "Drone #%d %s" % [d.id, "[COMPROMISED]" if d.get("comp", false) else "[SECURE]"]
	detail_pos.text = "Pos: (%d, %d) | Corr: %d%%" % [d.x/1000000, d.y/1000000, d.get("corr", 0)]
	detail_battery.text = "Battery: %d | Trust: %d" % [d.bat/1000000, d.get("trust", 100)]
	detail_state.text = "State: %s" % _state_name(d.state)
	detail_payload.text = "Inventory: %d" % d.inv
	detail_protocol.text = "System Health: %d%%" % (100 - d.get("corr", 0))

	var state_color = _state_color(d.state, d.get("comp", false))
	$VBoxContainer/DetailPanel/MarginContainer/VBoxContainer/StateIndicator.color = state_color

	_send_inspect_command(drone_id)

func _find_drone(drone_id: int) -> Dictionary:
	for d in swarm_registry.get("drones", []):
		if d.get("id", -1) == drone_id:
			return d
	return {}

# Called by game_hub.gd every tick when the Go engine is inspecting a drone.
# trace is a DecisionFrame: {drone_id, tick, steps: [{kind, name, result, taken}]}
func load_trace(trace: Dictionary) -> void:
	if not _trace_list:
		return
	_clear_trace()

	var steps = trace.get("steps", [])
	for step in steps:
		_trace_list.add_child(_make_step_row(step))

func _clear_trace() -> void:
	if not _trace_list:
		return
	for child in _trace_list.get_children():
		child.queue_free()

func _make_step_row(step: Dictionary) -> Control:
	var lbl = Label.new()
	lbl.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	lbl.add_theme_font_size_override("font_size", 10)
	lbl.clip_text = true

	var kind = step.get("kind", "")
	var name_str = step.get("name", "")
	var result_str = step.get("result", "")
	var taken = step.get("taken", false)

	var color: Color
	if kind == "condition":
		var marker = "✓" if taken else "✗"
		lbl.text = "%s [IF] %s → %s" % [marker, name_str, result_str]
		color = Color(0, 1, 0.62, 1) if taken else Color(0.5, 0.5, 0.5, 0.7)
	else:
		lbl.text = "  [•] %s → %s" % [name_str, result_str]
		color = Color(0, 1, 1, 1)

	lbl.add_theme_color_override("font_color", color)
	return lbl

# Called by game_hub.gd every tick with the full events array from the engine.
# Accumulates events for the currently-selected drone only.
func load_events(events: Array) -> void:
	if selected_drone_id < 0 or not _events_list:
		return

	for e in events:
		if e.get("drone_id", -1) == selected_drone_id:
			_recent_events.append(e)

	if _recent_events.size() > MAX_EVENTS_SHOWN:
		_recent_events = _recent_events.slice(_recent_events.size() - MAX_EVENTS_SHOWN)

	if not detail_panel.visible:
		return

	for child in _events_list.get_children():
		child.queue_free()

	# Most recent first
	var display = _recent_events.duplicate()
	display.reverse()
	for e in display:
		_events_list.add_child(_make_event_row(e))

func _make_event_row(e: Dictionary) -> Control:
	var lbl = Label.new()
	lbl.add_theme_font_size_override("font_size", 9)
	lbl.clip_text = true

	var event_type = e.get("type", "unknown")
	var tick = e.get("tick", 0)
	lbl.text = "T%d  %s" % [tick, event_type.replace("_", " ").to_upper()]
	lbl.add_theme_color_override("font_color", _event_color(event_type))
	return lbl

func _event_color(event_type: String) -> Color:
	match event_type:
		"drone_spawned", "fabricated": return Color(0, 1, 0.62, 1)
		"drone_died": return Color(1, 0.2, 0.2, 1)
		"harvested": return Color(1, 0.75, 0, 1)
		"deposited": return Color(0.2, 0.4, 1, 1)
		"drone_infected": return Color(0.8, 0, 1, 1)
		"hazard_damage": return Color(1, 0.4, 0, 1)
		"trust_changed": return Color(0.5, 0.5, 1, 1)
		_: return Color(0.6, 0.6, 0.6, 1)

func _state_name(idx: int) -> String:
	match idx:
		0: return "Searching"
		1: return "Returning"
		2: return "Inert"
		_: return "Unknown"

func _state_color(idx: int, compromised: bool) -> Color:
	if compromised:
		return Color(0.8, 0, 1, 1)
	match idx:
		0: return Color(0, 1, 0.62, 1)
		1: return Color(0.2, 0.4, 1, 1)
		2: return Color(1, 0.2, 0.2, 1)
		_: return Color(0.5, 0.5, 0.5, 1)

func _get_net() -> Node:
	return get_tree().root.find_child("NetworkBridge", true, false)

func _send_inspect_command(drone_id: int) -> void:
	var net = _get_net()
	if not net:
		return
	if drone_id >= 0:
		net.send_command("INSPECT_DRONE", {"drone_id": drone_id})
	else:
		net.send_command("INSPECT_CLEAR", {})

func _on_view_swarm_pressed() -> void:
	view_swarm_requested.emit()

func _on_back_pressed() -> void:
	detail_panel.hide()
	_send_inspect_command(-1)
	selected_drone_id = -1
	_recent_events.clear()
	_clear_trace()

func _exit_tree() -> void:
	if selected_drone_id >= 0:
		_send_inspect_command(-1)
