extends Node2D

var hazard_data: Dictionary = {}
var cell_size: Vector2 = Vector2(10, 10)
var overlay_visible: bool = true

func set_hazard_data(data: Dictionary) -> void:
	hazard_data = data
	queue_redraw()

func set_visible(v: bool) -> void:
	overlay_visible = v
	queue_redraw()

func _draw() -> void:
	if not overlay_visible:
		return

	var hazards = hazard_data.get("hazards", [])
	var hazard_colors = {
		"magnetic_anomaly": Color(1, 0, 0.5, 0.25),
		"thermal_geyser": Color(1, 0.4, 0, 0.25),
		"crust_collapse": Color(0.6, 0.3, 0, 0.25),
	}

	for h in hazards:
		var htype = h.get("type", "unknown")
		var color = hazard_colors.get(htype, Color(0.5, 0.5, 0.5, 0.2))
		var center = Vector2(
			h.get("epicenter", {}).get("x", 0) / 100.0,
			h.get("epicenter", {}).get("y", 0) / 100.0
		) * cell_size
		var radius = h.get("radius", 20) * cell_size.x
		var state = h.get("state", "dormant")

		match state:
			"dormant":
				draw_circle(center, radius, Color(color.r, color.g, color.b, color.a * 0.3))
				draw_arc(center, radius, 0, TAU, 32, Color(color.r, color.g, color.b, 0.5), 1.0)
			"expanding":
				draw_circle(center, radius, color)
				var pulse = radius + sin(Time.get_ticks_msec() * 0.005) * 5.0
				draw_arc(center, pulse, 0, TAU, 32, Color(1, 0.2, 0.2, 0.5), 2.0)
			"active":
				draw_circle(center, radius, Color(color.r, color.g, color.b, 0.4))
				draw_circle(center, radius * 0.5, Color(color.r, color.g, color.b, 0.7))
			"dissipating":
				draw_circle(center, radius, Color(color.r, color.g, color.b, color.a * 0.15))
				draw_arc(center, radius, 0, TAU * 0.75, 32, Color(color.r, color.g, color.b, 0.3), 1.0)
