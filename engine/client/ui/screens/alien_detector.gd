extends Control

var alien_data: Dictionary = {}

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var node_list: VBoxContainer = $VBoxContainer/ScrollContainer/NodeList
@onready var corruption_level: Label = $VBoxContainer/StatusBar/HBoxContainer/CorruptionLevel
@onready var spoof_detected: Label = $VBoxContainer/StatusBar/HBoxContainer/SpoofDetected
@onready var alert_indicator: ColorRect = $VBoxContainer/StatusBar/HBoxContainer/AlertIndicator

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/StatusBar, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)

func load_alien_nodes(data: Dictionary) -> void:
	alien_data = data
	_update_status()
	_populate_list()

func _update_status() -> void:
	var nodes = alien_data.get("nodes", [])
	var total_corruption = 0
	var spoofed = 0
	for n in nodes:
		total_corruption += n.get("corruption_radius", 0)
		if n.get("state") == "broadcasting":
			spoofed += 1

	corruption_level.text = "Corruption Level: %d" % total_corruption
	spoof_detected.text = "Spoofing Nodes: %d" % spoofed

	if spoofed > 0:
		alert_indicator.color = Color(1, 0.2, 0.2, 1)
	elif total_corruption > 0:
		alert_indicator.color = Color(1, 0.75, 0, 1)
	else:
		alert_indicator.color = Color(0, 1, 0.62, 1)

func _populate_list() -> void:
	for child in node_list.get_children():
		child.queue_free()

	for n in alien_data.get("nodes", []):
		var row = preload("res://ui/components/entity_row.tscn").instantiate()
		var state = n.get("state", "dormant")
		var state_color = _state_color(state)
		var pos = n.get("position", {"x": 0, "y": 0})

		var detail = "corruption: %d" % n.get("corruption_radius", 0)
		if state == "broadcasting":
			detail += " | spoofing: %s" % n.get("spoofing_signature", "unknown")

		row.set_data(
			"Alien Node",
			state.capitalize(),
			state_color,
			detail
		)
		node_list.add_child(row)

func _state_color(state: String) -> Color:
	match state:
		"dormant": return Color(0.3, 0.0, 0.3, 1)
		"awake": return Color(0.5, 0.0, 0.5, 1)
		"broadcasting": return Color(1, 0.0, 0.5, 1)
		"quarantined": return Color(0.0, 0.5, 0.0, 1)
		_: return Color(0.5, 0.5, 0.5, 1)

func _on_toggle_overlay_pressed() -> void:
	pass

func _on_deploy_countermeasure_pressed() -> void:
	pass
