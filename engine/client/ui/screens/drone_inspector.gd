extends Control

signal view_swarm_requested
signal view_entity_requested(entity_id: int)

var swarm_registry: Dictionary = {}
var selected_drone_id: int = -1

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
		total_bat += d.get("battery", 0)

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

func _state_name(idx: int) -> String:
	match idx:
		0: return "Searching"
		1: return "Returning"
		2: return "Inert"
		_: return "Unknown"

func _state_color(idx: int, compromised: bool) -> Color:
	if compromised:
		return Color(0.8, 0, 1, 1) # Viral Purple
	
	match idx:
		0: return Color(0, 1, 0.62, 1) # Cyan
		1: return Color(0.2, 0.4, 1, 1) # Blue
		2: return Color(1, 0.2, 0.2, 1) # Red (Inert)
		_: return Color(0.5, 0.5, 0.5, 1)

func _on_view_swarm_pressed() -> void:
	view_swarm_requested.emit()

func _on_back_pressed() -> void:
	detail_panel.hide()
	selected_drone_id = -1
