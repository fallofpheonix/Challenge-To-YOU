extends Node2D

var alien_data: Dictionary = {}
var cell_size: Vector2 = Vector2(10, 10)
var overlay_visible: bool = true

func set_alien_data(data: Dictionary) -> void:
	alien_data = data
	queue_redraw()

func set_visible(v: bool) -> void:
	overlay_visible = v
	queue_redraw()

func _draw() -> void:
	if not overlay_visible:
		return

	var nodes = alien_data.get("nodes", [])
	for n in nodes:
		var state = n.get("state", "dormant")
		var center = Vector2(
			n.get("position", {}).get("x", 0) / 100.0,
			n.get("position", {}).get("y", 0) / 100.0
		) * cell_size
		var corruption = n.get("corruption_radius", 10) * cell_size.x

		match state:
			"dormant":
				draw_circle(center, 4.0, Color(0.3, 0.0, 0.3, 0.5))
			"awake":
				draw_circle(center, 6.0, Color(0.5, 0.0, 0.5, 0.8))
				draw_arc(center, corruption, 0, TAU, 32, Color(0.5, 0.0, 0.5, 0.3), 1.0)
			"broadcasting":
				draw_circle(center, 8.0, Color(1.0, 0.0, 0.5, 0.9))
				draw_circle(center, corruption, Color(1.0, 0.0, 0.5, 0.1))
				var pulse = corruption + sin(Time.get_ticks_msec() * 0.003) * 5.0
				draw_arc(center, pulse, 0, TAU, 32, Color(1.0, 0.0, 0.5, 0.5), 2.0)

				var spoof_dir = Vector2(1, 0).rotated(Time.get_ticks_msec() * 0.001)
				for i in range(4):
					var angle = i * TAU / 4.0 + Time.get_ticks_msec() * 0.001
					var tip = center + Vector2(cos(angle), sin(angle)) * corruption
					draw_line(center, tip, Color(1.0, 0.0, 0.5, 0.3), 1.0)
			"quarantined":
				draw_circle(center, 6.0, Color(0.0, 0.5, 0.0, 0.8))
				draw_arc(center, corruption * 1.2, 0, TAU, 32, Color(0.0, 1.0, 0.0, 0.5), 2.0)
