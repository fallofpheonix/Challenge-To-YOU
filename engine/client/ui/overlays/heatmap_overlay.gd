extends Node2D

var density_data: Dictionary = {}
var cell_size: Vector2 = Vector2(10, 10)
var density_grid: Dictionary = {}
var overlay_visible: bool = true

func set_density_data(data: Dictionary) -> void:
	density_data = data
	density_grid = data.get("density_grid", {})
	queue_redraw()

func set_visible(v: bool) -> void:
	overlay_visible = v
	queue_redraw()

func _draw() -> void:
	if not overlay_visible or density_grid.is_empty():
		return

	var min_density = 0
	var max_density = 0
	for pos_key in density_grid:
		var val = density_grid[pos_key]
		if val > max_density:
			max_density = val
		if val < min_density:
			min_density = val
	var range_density = max(max_density - min_density, 1)

	for pos_key in density_grid:
		var parts = pos_key.split(",")
		if parts.size() != 2:
			continue
		var gx = int(parts[0])
		var gy = int(parts[1])
		var val = density_grid[pos_key]
		var normalized = float(val - min_density) / range_density

		var color = _heat_color(normalized)
		var rect = Rect2(
			gx * cell_size.x,
			gy * cell_size.y,
			cell_size.x,
			cell_size.y
		)
		draw_rect(rect, color, true)

func _heat_color(normalized: float) -> Color:
	if normalized < 0.2:
		return Color(0.0, 0.0, 0.3, normalized * 0.3)
	elif normalized < 0.4:
		return Color(0.0, 0.3, 0.6, normalized * 0.4)
	elif normalized < 0.6:
		return Color(0.0, 0.6, 0.3, normalized * 0.5)
	elif normalized < 0.8:
		return Color(0.8, 0.8, 0.0, normalized * 0.6)
	else:
		return Color(1.0, 0.2, 0.0, normalized * 0.7)

func clear() -> void:
	density_grid = {}
	queue_redraw()
