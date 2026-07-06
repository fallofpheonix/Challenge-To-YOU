extends Control

var structure_data: Dictionary = {}
var blueprint_types = [
	{"type": "hub", "label": "Hub", "cost": "500 Silicate", "desc": "Drone fabrication & storage"},
	{"type": "relay_node", "label": "Relay Node", "cost": "200 Silicate", "desc": "Extends telemetry range"},
	{"type": "storage_cache", "label": "Storage Cache", "cost": "300 Silicate", "desc": "Local resource depot"},
]

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var structure_list: VBoxContainer = $VBoxContainer/ScrollContainer/StructureList
@onready var blueprint_panel: PanelContainer = $VBoxContainer/BlueprintPanel
@onready var blueprint_container: VBoxContainer = $VBoxContainer/BlueprintPanel/MarginContainer/VBoxContainer
@onready var detail_panel: PanelContainer = $VBoxContainer/DetailPanel
@onready var detail_info: RichTextLabel = $VBoxContainer/DetailPanel/MarginContainer/RichTextLabel

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/BlueprintPanel, theme_ctrl.colors.surface, theme_ctrl.colors.border)
	theme_ctrl.apply_panel_style($VBoxContainer/DetailPanel, theme_ctrl.colors.surface, theme_ctrl.colors.border)
	detail_panel.hide()
	_populate_blueprints()

func load_structures(data: Dictionary) -> void:
	structure_data = data
	_populate_structures()

func _populate_blueprints() -> void:
	for child in blueprint_container.get_children():
		child.queue_free()

	var header = preload("res://ui/components/section_header.tscn").instantiate()
	header.set_title("Available Blueprints")
	blueprint_container.add_child(header)

	for bp in blueprint_types:
		var card = preload("res://ui/components/chrysalis_card.tscn").instantiate()
		card.header_text = bp.label
		card.body_text = "[b]Cost:[/b] %s\n%s" % [bp.cost, bp.desc]
		card.accent_color = theme_ctrl.colors.structure_hub if bp.type == "hub" else theme_ctrl.colors.structure_relay if bp.type == "relay_node" else theme_ctrl.colors.structure_storage
		var deploy_btn = Button.new()
		deploy_btn.text = "DEPLOY"
		deploy_btn.pressed.connect(func(): pass)
		card.add_child(deploy_btn)
		blueprint_container.add_child(card)

func _populate_structures() -> void:
	for child in structure_list.get_children():
		child.queue_free()

	for s in structure_data.get("structures", []):
		var row = preload("res://ui/components/entity_row.tscn").instantiate()
		var stype = s.get("type", "unknown")
		var state = s.get("state", "operational")
		var color = _structure_type_color(stype)
		var state_color = _state_color(state)
		row.set_data(
			"%s %s" % [stype.capitalize(), s.get("id", 0)],
			state.capitalize(),
			state_color,
			"integrity: %d" % s.get("integrity", 100)
		)
		var sid = s.get("id", -1)
		row.clicked.connect(func(): _select_structure(sid))
		structure_list.add_child(row)

func _select_structure(structure_id: int) -> void:
	for s in structure_data.get("structures", []):
		if s.get("id", -1) == structure_id:
			detail_panel.show()
			detail_info.text = "[b]Structure #%d[/b]\nType: %s\nState: %s\nIntegrity: %d\nInventory: %s" % [
				s.id, s.type.capitalize(), s.state.capitalize(), s.get("integrity", 100), s.get("inventory", "Empty")
			]

func _structure_type_color(stype: String) -> Color:
	match stype:
		"hub": return theme_ctrl.colors.structure_hub
		"relay_node": return theme_ctrl.colors.structure_relay
		"storage_cache": return theme_ctrl.colors.structure_storage
		_: return Color(0.5, 0.5, 0.5, 1)

func _state_color(state: String) -> Color:
	match state:
		"blueprint": return Color(0.3, 0.3, 0.6, 1)
		"constructing": return Color(1, 0.75, 0, 1)
		"operational": return Color(0, 1, 0.62, 1)
		"offline": return Color(1, 0.2, 0.2, 1)
		_: return Color(0.5, 0.5, 0.5, 1)

func _on_detail_back_pressed() -> void:
	detail_panel.hide()
