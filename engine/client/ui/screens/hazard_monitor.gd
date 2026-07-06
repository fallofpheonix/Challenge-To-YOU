extends Control

var hazard_data: Dictionary = {}

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var hazard_list: VBoxContainer = $VBoxContainer/ScrollContainer/HazardList
@onready var global_threat_label: Label = $VBoxContainer/ThreatBar/HBoxContainer/GlobalThreat
@onready var active_hazard_label: Label = $VBoxContainer/ThreatBar/HBoxContainer/ActiveHazards

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/ThreatBar, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)

func load_hazards(data: Dictionary) -> void:
	hazard_data = data
	_update_threat_bar()
	_populate_list()

func _update_threat_bar() -> void:
	var hazards = hazard_data.get("hazards", [])
	var active = hazards.size()
	global_threat_label.text = "Global Threat Level: %s" % _threat_level(active)
	active_hazard_label.text = "Active Hazards: %d" % active

func _populate_list() -> void:
	for child in hazard_list.get_children():
		child.queue_free()

	var hazard_types = [
		{"name": "Magnetic Anomaly", "color": theme_ctrl.colors.hazard_magnetic, "icon": "⚡"},
		{"name": "Thermal Geyser", "color": theme_ctrl.colors.hazard_thermal, "icon": "🔥"},
	]

	for h in hazard_data.get("hazards", []):
		var htype_idx = h.get("type", 0)
		var config = hazard_types[htype_idx] if htype_idx < hazard_types.size() else {"name": "Unknown", "color": Color(0.5, 0.5, 0.5, 1), "icon": "?"}
		
		var x = h.get("x", 0)
		var y = h.get("y", 0)
		var radius = h.get("rad", 0)

		var row = preload("res://ui/components/entity_row.tscn").instantiate()
		row.set_data(
			"%s %s" % [config.icon, config.name],
			"radius: %d" % radius,
			config.color,
			"(%d, %d)" % [x, y]
		)
		hazard_list.add_child(row)

func _state_color(state: String) -> Color:
	match state:
		"dormant": return Color(0.3, 0.3, 0.3, 1)
		"expanding": return Color(1, 0.75, 0, 1)
		"active": return Color(1, 0.2, 0.2, 1)
		"dissipating": return Color(0.6, 0.3, 0.0, 1)
		_: return Color(0.5, 0.5, 0.5, 1)

func _on_toggle_overlay_pressed() -> void:
	pass
