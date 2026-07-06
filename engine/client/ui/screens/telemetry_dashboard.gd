extends Control

var telemetry_data: Dictionary = {}
var grid_data: Array = []

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var heatmap_preview: ColorRect = $VBoxContainer/ViewportPanel/VBoxContainer/HeatmapPreview
@onready var tick_counter: Label = $VBoxContainer/StatsBar/HBoxContainer/TickCounter
@onready var drone_count: Label = $VBoxContainer/StatsBar/HBoxContainer/DroneCount
@onready var resource_rate: Label = $VBoxContainer/StatsBar/HBoxContainer/ResourceRate
@onready var bandwidth_usage: Label = $VBoxContainer/StatsBar/HBoxContainer/BandwidthUsage
@onready var log_console: RichTextLabel = $VBoxContainer/LogPanel/MarginContainer/LogConsole
@onready var view_toggle_btn: Button = $VBoxContainer/ViewportPanel/VBoxContainer/Controls/ViewToggleBtn
@onready var overlay_toggle_btn: Button = $VBoxContainer/ViewportPanel/VBoxContainer/Controls/OverlayToggleBtn

var view_modes = ["Orbital", "Density", "Threat", "Logic Flow"]
var current_view_mode = 0

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/StatsBar, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)
	theme_ctrl.apply_panel_style($VBoxContainer/ViewportPanel, theme_ctrl.colors.surface, theme_ctrl.colors.border)
	theme_ctrl.apply_panel_style($VBoxContainer/LogPanel, theme_ctrl.colors.surface, theme_ctrl.colors.border)
	view_toggle_btn.text = "View: %s" % view_modes[current_view_mode]
	heatmap_preview.draw.connect(_draw_heatmap)

func load_telemetry(data: Dictionary) -> void:
	telemetry_data = data
	_update_stats()
	_update_log()

func load_grid(data: Dictionary) -> void:
	grid_data = data.get("grid", [])
	heatmap_preview.queue_redraw()

func _draw_heatmap() -> void:
	if grid_data.is_empty():
		return
	
	var cell_size = heatmap_preview.size.x / 100.0
	for cell in grid_data:
		var x = cell.get("x", 0)
		var y = cell.get("y", 0)
		var home = cell.get("home", 0) / 1000000.0
		var res = cell.get("res", 0) / 1000000.0
		
		var rect = Rect2(Vector2(x, y) * cell_size, Vector2(cell_size, cell_size))
		if res > 0.01:
			heatmap_preview.draw_rect(rect, Color(1, 0.75, 0, res * 0.5), true)
		if home > 0.01:
			heatmap_preview.draw_rect(rect, Color(0, 0.4, 1, home * 0.5), true)

func _update_stats() -> void:
	tick_counter.text = "Tick: %d" % telemetry_data.get("tick", 0)
	drone_count.text = "Drones: %d" % telemetry_data.get("swarm_size", 0)
	resource_rate.text = "Colony Silicates: %d" % telemetry_data.get("colony_res", 0)
	var bw_used = telemetry_data.get("bandwidth_used", 0)
	if bw_used > 0:
		bandwidth_usage.text = "BW: %d%%" % (bw_used * 100 / max(telemetry_data.get("bandwidth_max", 1), 1))
	else:
		bandwidth_usage.text = "BW: --"

func _update_log() -> void:
	var entries = telemetry_data.get("log", [])
	var lines = []
	for e in entries:
		var color = _log_color(e.get("level", "info"))
		lines.append("[color=%s][%s] %s[/color]" % [color, e.get("level", "INFO").to_upper(), e.get("message", "")])
	log_console.text = "\n".join(lines)

func _log_color(level: String) -> String:
	match level:
		"info": return "#48aacc"
		"warn": return "#ffbb00"
		"error": return "#ff3333"
		"debug": return "#888888"
		_: return "#cccccc"

func _on_view_toggle_pressed() -> void:
	current_view_mode = (current_view_mode + 1) % view_modes.size()
	view_toggle_btn.text = "View: %s" % view_modes[current_view_mode]

func _on_overlay_toggle_pressed() -> void:
	pass

func _on_clear_log_pressed() -> void:
	log_console.text = ""

func cycle_view_mode() -> void:
	_on_view_toggle_pressed()
