extends Node2D

var grid_data: Dictionary = {}
var cell_size: Vector2 = Vector2(10, 10)
var visible_layer: String = "resource_found"

func set_grid_data(data: Dictionary) -> void:
	grid_data = data
	queue_redraw()

func set_visible_layer(layer: String) -> void:
	visible_layer = layer
	queue_redraw()

func _draw() -> void:
	var signals = grid_data.get("signals", [])
	var layer_colors = {
		"resource_found": Color(0, 1, 0.2, 0.4),
		"hazard_warning": Color(1, 0.2, 0, 0.4),
		"rally_point": Color(0.2, 0.4, 1, 0.4),
		"spoofed_data": Color(1, 0, 0.5, 0.4),
	}

	var base_color = layer_colors.get(visible_layer, Color(0.5, 0.5, 0.5, 0.3))

	for s in signals:
		if s.get("type") == visible_layer:
			var pos = s.get("position", {"x": 0, "y": 0})
			var intensity = s.get("intensity", 50) / 100.0
			var center = Vector2(pos.x / 100.0, pos.y / 100.0) * cell_size
			var radius = 8.0 + intensity * 16.0
			var color = Color(base_color.r, base_color.g, base_color.b, base_color.a * intensity)

			draw_circle(center, radius, color)

			if intensity > 0.5:
				draw_circle(center, radius * 0.5, Color(base_color.r, base_color.g, base_color.b, base_color.a * 0.6))
