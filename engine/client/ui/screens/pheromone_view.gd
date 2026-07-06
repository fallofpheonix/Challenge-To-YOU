extends Control

var pheromone_data: Dictionary = {}
var grid_data: Array = []

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var signal_list: VBoxContainer = $VBoxContainer/ScrollContainer/SignalList
@onready var overlay_toggle: Button = $VBoxContainer/Controls/OverlayToggle
@onready var gradient_preview: PanelContainer = $VBoxContainer/GradientPreview

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/GradientPreview, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)
	# Add a dedicated drawing node for the heatmap
	var canvas = Control.new()
	canvas.name = "HeatmapCanvas"
	canvas.custom_minimum_size = Vector2(400, 400)
	canvas.size_flags_horizontal = Control.SIZE_SHRINK_CENTER
	canvas.draw.connect(_draw_heatmap.bind(canvas))
	$VBoxContainer.add_child(canvas)
	$VBoxContainer.move_child(canvas, 2) # Place it after controls but before the list

func load_pheromones(data: Dictionary) -> void:
	pheromone_data = data
	_populate_list()

func load_grid(data: Dictionary) -> void:
	grid_data = data.get("grid", [])
	var canvas = $VBoxContainer/HeatmapCanvas
	if canvas:
		canvas.queue_redraw()

func _draw_heatmap(canvas: Control) -> void:
	if grid_data.is_empty():
		return
	
	var cell_size = canvas.size.x / 100.0 # Assuming 100x100 grid
	
	for cell in grid_data:
		var x = cell.get("x", 0)
		var y = cell.get("y", 0)
		var home = cell.get("home", 0) / 1000000.0
		var res = cell.get("res", 0) / 1000000.0
		var count = cell.get("cnt", 0)
		
		var rect = Rect2(Vector2(x, y) * cell_size, Vector2(cell_size, cell_size))
		
		if count > 0:
			canvas.draw_rect(rect, Color(0.8, 0.8, 0.8, 0.5), true)
		
		if res > 0.01:
			# Amber color for resource trails
			canvas.draw_rect(rect, Color(1, 0.75, 0, res * 0.8), true)
		if home > 0.01:
			# Blue color for home trails
			canvas.draw_rect(rect, Color(0, 0.4, 1, home * 0.8), true)

func _populate_list() -> void:
	for child in signal_list.get_children():
		child.queue_free()

	var signals = pheromone_data.get("signals", [])
	var by_type = {}
	for s in signals:
		var sig_type = s.get("type", "unknown")
		if not by_type.has(sig_type):
			by_type[sig_type] = []
		by_type[sig_type].append(s)

	var type_order = ["resource_found", "hazard_warning", "rally_point", "spoofed_data"]
	var type_colors = {
		"resource_found": theme_ctrl.colors.pheromone_resource,
		"hazard_warning": theme_ctrl.colors.pheromone_hazard,
		"rally_point": theme_ctrl.colors.pheromone_rally,
		"spoofed_data": theme_ctrl.colors.pheromone_spoofed,
	}

	for sig_type in type_order:
		var entries = by_type.get(sig_type, [])
		var section = preload("res://ui/components/section_header.tscn").instantiate()
		var display_name = sig_type.replace("_", " ").capitalize()
		section.set_title("%s (%d)" % [display_name, entries.size()])
		signal_list.add_child(section)

		for s in entries:
			var row = preload("res://ui/components/entity_row.tscn").instantiate()
			var pos = s.get("position", {"x": 0, "y": 0})
			var intensity = s.get("intensity", 0)
			var intensity_color = Color(0, 1, 0.2, float(intensity) / 100.0)
			row.set_data(
				"@ (%d, %d)" % [pos.x, pos.y],
				"intensity: %d" % intensity,
				intensity_color,
				""
			)
			signal_list.add_child(row)

func _on_overlay_toggle_pressed() -> void:
	overlay_toggle.text = "Hide Overlay" if overlay_toggle.text == "Show Overlay" else "Show Overlay"
