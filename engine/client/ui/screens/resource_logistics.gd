extends Control

var resource_data: Dictionary = {}

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var total_silicate: Label = $VBoxContainer/SummaryBar/HBoxContainer/SilicateCount
@onready var total_isotope: Label = $VBoxContainer/SummaryBar/HBoxContainer/IsotopeCount
@onready var resource_list: VBoxContainer = $VBoxContainer/ScrollContainer/ResourceList
@onready var flow_panel: PanelContainer = $VBoxContainer/FlowPanel
@onready var flow_label: RichTextLabel = $VBoxContainer/FlowPanel/MarginContainer/FlowLabel

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/SummaryBar, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)
	theme_ctrl.apply_panel_style(flow_panel, theme_ctrl.colors.surface, theme_ctrl.colors.border)
	flow_panel.hide()

func load_resources(data: Dictionary) -> void:
	resource_data = data
	_update_summary()
	_populate_list()
	_update_flow()

func _update_summary() -> void:
	var sil = 0
	var iso = 0
	for r in resource_data.get("resources", []):
		if r.get("type") == "silicate":
			sil += r.get("yield", 0)
		elif r.get("type") == "isotope":
			iso += r.get("yield", 0)
	total_silicate.text = "Silicate: %d" % sil
	total_isotope.text = "Isotope: %d" % iso

func _populate_list() -> void:
	for child in resource_list.get_children():
		child.queue_free()

	for r in resource_data.get("resources", []):
		var row = preload("res://ui/components/entity_row.tscn").instantiate()
		var rtype = r.get("type", "unknown")
		var color = theme_ctrl.colors.resource_silicate if rtype == "silicate" else theme_ctrl.colors.resource_isotope
		var pos = r.get("position", {"x": 0, "y": 0})
		var state = r.get("state", "embedded")
		var state_color = Color(0.6, 0.6, 0.3, 1) if state == "embedded" else Color(0.3, 0.6, 0.3, 1) if state == "carried" else Color(0.3, 0.6, 1, 1)
		row.set_data(
			"%s deposit" % rtype.capitalize(),
			"yield: %d" % r.get("yield", 0),
			state_color,
			"(%d, %d)" % [pos.x, pos.y]
		)
		resource_list.add_child(row)

func _update_flow() -> void:
	var lines = []
	lines.append("[b]Supply Chain Flow[/b]")
	lines.append("")
	var embedded = 0
	var carried = 0
	var stored = 0
	for r in resource_data.get("resources", []):
		match r.get("state"):
			"embedded": embedded += 1
			"carried": carried += 1
			"stored": stored += 1
	lines.append("Embedded: %d  |  Carried: %d  |  Stored: %d" % [embedded, carried, stored])
	flow_label.text = "\n".join(lines)
	flow_panel.show()

func _on_toggle_flow_pressed() -> void:
	flow_panel.visible = not flow_panel.visible
