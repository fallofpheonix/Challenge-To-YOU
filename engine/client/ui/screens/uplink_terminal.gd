extends Control

signal deploy_code_requested(code_text: String)
signal queue_reorder_requested(queue_id: int)

var uplink_data: Dictionary = {}

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var window_status: Label = $VBoxContainer/StatusBar/HBoxContainer/WindowStatus
@onready var queue_size: Label = $VBoxContainer/StatusBar/HBoxContainer/QueueSize
@onready var queue_list: VBoxContainer = $VBoxContainer/QueueContainer/ScrollContainer/QueueList
@onready var bandwidth_bar: ColorRect = $VBoxContainer/StatusBar/HBoxContainer/BandwidthBar
@onready var deploy_button: Button = $VBoxContainer/DeployBar/DeployButton

func _ready() -> void:
	theme_ctrl.apply_panel_style($VBoxContainer/StatusBar, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)
	theme_ctrl.apply_panel_style($VBoxContainer/QueueContainer, theme_ctrl.colors.surface, theme_ctrl.colors.border)
	_update_view()

func load_uplink(data: Dictionary) -> void:
	uplink_data = data
	_update_view()

func _update_view() -> void:
	var window_open = uplink_data.get("window_open", false)
	var time_remaining = uplink_data.get("time_remaining", 0)
	var queue = uplink_data.get("queue", [])
	var bandwidth = uplink_data.get("bandwidth", 0)

	if window_open:
		window_status.text = "UPLINK WINDOW OPEN"
		window_status.add_theme_color_override("font_color", Color(0, 1, 0.62, 1))
		deploy_button.disabled = false
	else:
		window_status.text = "UPLINK CLOSED  (next window: %d ticks)" % time_remaining
		window_status.add_theme_color_override("font_color", Color(1, 0.2, 0.2, 1))
		deploy_button.disabled = true

	queue_size.text = "Queue: %d / %d" % [queue.size(), bandwidth]
	bandwidth_bar.color = Color(0, 1, 0.62, 0.3 + 0.7 * float(queue.size()) / max(bandwidth, 1))

	for child in queue_list.get_children():
		child.queue_free()

	for i in range(queue.size()):
		var entry = queue[i]
		var row = preload("res://ui/components/entity_row.tscn").instantiate()
		var status_color = Color(0, 1, 0.62, 1) if entry.get("status") == "pending" else Color(1, 0.75, 0, 1) if entry.get("status") == "deploying" else Color(0.3, 0.6, 1, 1)
		row.set_data(
			"Deploy #%d" % entry.get("id", i),
			entry.get("status", "pending").capitalize(),
			status_color,
			"%d lines" % entry.get("lines", 0)
		)
		queue_list.add_child(row)

func _on_deploy_pressed() -> void:
	var code = $VBoxContainer/DeployBar/CodePreview.text
	deploy_code_requested.emit(code)
