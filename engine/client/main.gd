# Chrysalis Swarm Client: Main Controller
# This script acts as the "Eyes" of the game, visualizing the Go simulation state.
extends Control

var go_stdout: FileAccess
var agent_sprite: Sprite2D
var goal_sprite: Sprite2D
var complete_label: Label
var code_editor: CodeEdit
var error_console: RichTextLabel
var level_title: Label
var description_label: RichTextLabel
var obstacles_container: Node2D
var telemetry_label: Label
var trail_container: Node2D
var trail_positions = []
const MAX_TRAIL = 5
var network_bridge: Node

var current_lang = "ps"
var current_script_path = ""
var active_pipe = null
var mission_terminal := false
var current_level_path := ""
var current_level_id := ""

# Level ordering for campaign progression
var level_order = [
	"chrysalis_1",
	"chrysalis_2",
	"chrysalis_3",
]

var templates = {
	"ps": "fn main() {\n    if (SENSE_BATTERY() < 25000000) {\n        MOVE_TOWARDS_HOME()\n    } else {\n        if (SENSE_CARGO()) {\n            DROP_RESOURCE()\n            MOVE_TOWARDS_HOME()\n        } else {\n            HARVEST()\n            if (SENSE_CARGO()) {\n                MOVE_TOWARDS_HOME()\n            } else {\n                MOVE_TOWARDS_RESOURCE()\n            }\n        }\n    }\n}\n",
	"py": "# Chrysalis Agent Loop\nwhile get_x() < get_goal_x():\n    move_forward()\n",
}

func _ready():
	var hub = get_node_or_null("/root/GameHub")
	if hub:
		hub.show()

	agent_sprite = $HBoxContainer/WorldView/SimulationSpace/Agent
	goal_sprite = $HBoxContainer/WorldView/SimulationSpace/Goal
	obstacles_container = $HBoxContainer/WorldView/SimulationSpace/Obstacles
	trail_container = $HBoxContainer/WorldView/SimulationSpace/Trail
	complete_label = $HBoxContainer/WorldView/UIMessage
	code_editor = $HBoxContainer/EditorPanel/CodeEditor
	error_console = $HBoxContainer/EditorPanel/ErrorConsole
	level_title = $HBoxContainer/WorldView/LevelTitle
	description_label = $HBoxContainer/EditorPanel/DescriptionLabel
	telemetry_label = $HBoxContainer/WorldView/TelemetryPanel/TelemetryLabel

	setup_highlighter()
	complete_label.hide()

	# Build end-of-game button panel
	_build_endgame_panel()

	set_language("ps")

	# Initialize Networking
	network_bridge = preload("res://network_bridge.gd").new()
	network_bridge.name = "NetworkBridge"
	add_child(network_bridge)
	network_bridge.packet_received.connect(_on_network_packet)

	# Read level path from env var (set by level_select.gd)
	current_level_path = OS.get_environment("PHX_LEVEL_PATH")
	if current_level_path.is_empty():
		# Default to chrysalis_1 if no level selected
		current_level_path = ProjectSettings.globalize_path("res://../core/levels/chrysalis_1.json")

	start_go_core()

func _build_endgame_panel():
	var theme = get_node_or_null("/root/ChrysalisTheme")
	var panel = PanelContainer.new()
	panel.name = "EndgamePanel"
	panel.visible = false
	panel.set_anchors_preset(Control.PRESET_CENTER)
	panel.offset_left = -100
	panel.offset_right = 100
	panel.offset_top = 120
	panel.offset_bottom = 260
	$HBoxContainer/WorldView.add_child(panel)

	if theme:
		theme.apply_panel_style(panel)

	var vbox = VBoxContainer.new()
	vbox.theme_override_constants/separation = 8
	panel.add_child(vbox)

	var retry_btn = Button.new()
	retry_btn.text = "RETRY"
	retry_btn.pressed.connect(_on_retry_pressed)
	vbox.add_child(retry_btn)
	if theme:
		theme.apply_button_style(retry_btn)

	var next_btn = Button.new()
	next_btn.text = "NEXT LEVEL"
	next_btn.name = "NextLevelBtn"
	next_btn.pressed.connect(_on_next_level_pressed)
	vbox.add_child(next_btn)
	if theme:
		theme.apply_button_style(next_btn, theme.colors.surface_alt, theme.colors.accent)

	var menu_btn = Button.new()
	menu_btn.text = "BACK TO MENU"
	menu_btn.pressed.connect(_on_back_to_menu_pressed)
	vbox.add_child(menu_btn)
	if theme:
		theme.apply_button_style(menu_btn, theme.colors.surface_alt, theme.colors.secondary_text)

func setup_highlighter():
	var highlighter = CodeHighlighter.new()

	code_editor.add_theme_color_override("font_color", Color(0.9, 0.9, 0.9, 1))

	var kw_color = Color(0, 1, 1, 1)
	var keywords = ["fn", "let", "if", "else", "while"]
	for kw in keywords:
		highlighter.add_keyword_color(kw, kw_color)

	var api_color = Color(0, 1, 0.2, 1)
	var api_funcs = ["HARVEST", "DROP_RESOURCE", "SENSE_RESOURCE", "SENSE_HOME", "SENSE_CARGO", "SENSE_BATTERY", "SENSE_TRUST", "SENSE_CORRUPTION", "SENSE_COMPROMISED", "SENSE_ALIEN_SIGNAL", "SENSE_SWARM_SIZE", "SENSE_COLONY_RESOURCES", "BROADCAST_VOTE", "MOVE_RANDOM", "MOVE_TOWARDS_RESOURCE", "MOVE_TOWARDS_HOME"]
	for fn in api_funcs:
		highlighter.add_keyword_color(fn, api_color)

	highlighter.add_color_region("//", "", Color(0.6, 0.6, 0.6, 1), true)
	highlighter.add_color_region('"', '"', Color(1, 0.75, 0, 1), false)

	highlighter.number_color = Color(0.8, 0.4, 0.8, 1)
	highlighter.symbol_color = Color(0.9, 0.9, 0.9, 1)
	highlighter.function_color = Color(0.4, 0.8, 1, 1)
	highlighter.member_variable_color = Color(0.9, 0.9, 0.9, 1)

	code_editor.syntax_highlighter = highlighter

func set_language(lang):
	current_lang = lang
	code_editor.text = templates[lang]
	var toggle_box = $HBoxContainer/EditorPanel/LangToggle
	for child in toggle_box.get_children():
		if child is Button:
			if child.name == "BtnPS" and lang == "ps": child.modulate = Color(0, 1, 1, 1)
			elif child.name == "BtnPython" and lang == "py": child.modulate = Color(0, 1, 1, 1)
			else: child.modulate = Color(0.5, 0.5, 0.5, 1)

func _on_lang_ps(): set_language("ps")
func _on_lang_python(): set_language("py")
func _on_lang_cpp(): print("C++ agent support not yet implemented")
func _on_lang_java(): print("Java agent support not yet implemented")

func start_go_core():
	var core_bin = ProjectSettings.globalize_path("res://../bin/chrysalis-core")
	current_script_path = ProjectSettings.globalize_path("res://../core/scripts/agent.ps")
	OS.set_environment("PHX_SCRIPT_PATH", current_script_path)
	OS.set_environment("PHX_LEVEL_PATH", current_level_path)

	# Extract level ID from path for campaign tracking
	var level_file = current_level_path.get_file().replace(".json", "")
	current_level_id = level_file

	# Update level title from JSON
	_load_level_info()

	if not FileAccess.file_exists(core_bin):
		push_error("[Network Error] Missing core binary. Run `make build-core` from engine/.")
		return

	var pid = OS.create_process(core_bin, [])
	if pid != -1:
		active_pipe = {"pid": pid}
		print("[Network] Launched Go core (PID: %d). Awaiting WebSocket handshake..." % pid)
	else:
		print("[Network Error] Failed to launch Go core.")

func _load_level_info():
	var file = FileAccess.open(current_level_path, FileAccess.READ)
	if file:
		var json = JSON.new()
		if json.parse(file.get_as_text()) == OK:
			var data = json.get_data()
			level_title.text = data.get("title", current_level_id)
			description_label.text = data.get("description", "")
			return
	level_title.text = current_level_id

func _on_network_packet(type: String, data: Dictionary):
	if type == "EMISSION_SNAPSHOT":
		update_ui(data["payload"])

func _process(_delta):
	pass

func update_ui(data):
	# Forward data to the Game Hub if it exists
	var hub = get_node_or_null("/root/GameHub")
	if hub:
		hub.forward_data(data)

	if data.has("mission"):
		_update_mission_status(data["mission"])

	# 1. Handle Tick
	if data.has("tick"):
		telemetry_label.text = "TICK: %03d" % data["tick"]

	# 2. Handle Drones
	if data.has("drones"):
		_update_drones_visuals(data["drones"])

		# Breakpoint condition: Drone battery critically low
		for drone in data["drones"]:
			if not mission_terminal and drone["state"] == 0 and drone["bat"] < 25000000:
				_trigger_breakpoint_halt(drone["id"])
				return

func _update_mission_status(mission: Dictionary) -> void:
	var status = mission.get("status", "running")
	if status == "running":
		mission_terminal = false
		complete_label.hide()
		_hide_endgame_panel()
		return

	mission_terminal = true
	var reason = str(mission.get("reason", "")).replace("_", " ").to_upper()
	if status == "victory":
		complete_label.text = "VICTORY"
		complete_label.modulate = Color(0, 1, 0.62, 1)
		_save_progress()
	elif status == "defeat":
		complete_label.text = "DEFEAT"
		complete_label.modulate = Color(1, 0.2, 0.2, 1)
	else:
		complete_label.text = str(status).to_upper()
		complete_label.modulate = Color(1, 1, 1, 1)

	if not reason.is_empty():
		complete_label.text += "\n%s" % reason
	complete_label.show()
	_show_endgame_panel(status == "victory")

func _show_endgame_panel(is_victory: bool):
	var panel = $HBoxContainer/WorldView/EndgamePanel
	if panel:
		panel.visible = true
		var next_btn = panel.get_node_or_null("VBoxContainer/NextLevelBtn")
		if next_btn:
			next_btn.visible = is_victory and _has_next_level()

func _hide_endgame_panel():
	var panel = $HBoxContainer/WorldView/EndgamePanel
	if panel:
		panel.visible = false

func _has_next_level() -> bool:
	var idx = level_order.find(current_level_id)
	return idx >= 0 and idx < level_order.size() - 1

func _save_progress():
	var progress_path = ProjectSettings.globalize_path("res://../core/levels/progress.json")
	var progress = {}
	if FileAccess.file_exists(progress_path):
		var file = FileAccess.open(progress_path, FileAccess.READ)
		if file:
			var json = JSON.new()
			if json.parse(file.get_as_text()) == OK:
				progress = json.get_data()
	if not progress.has("completed"):
		progress["completed"] = {}
	progress["completed"][current_level_id] = true
	var file = FileAccess.open(progress_path, FileAccess.WRITE)
	if file:
		file.store_string(JSON.stringify(progress, "\t"))
		file.close()

func _get_next_level_path() -> String:
	var idx = level_order.find(current_level_id)
	if idx < 0 or idx >= level_order.size() - 1:
		return ""
	return ProjectSettings.globalize_path("res://../core/levels/%s.json" % level_order[idx + 1])

func _on_retry_pressed():
	kill_core()
	_hide_endgame_panel()
	complete_label.hide()
	mission_terminal = false
	start_go_core()

func _on_next_level_pressed():
	var next_path = _get_next_level_path()
	if next_path.is_empty():
		return
	kill_core()
	current_level_path = next_path
	OS.set_environment("PHX_LEVEL_PATH", current_level_path)
	current_level_id = current_level_path.get_file().replace(".json", "")
	_load_level_info()
	_hide_endgame_panel()
	complete_label.hide()
	mission_terminal = false
	start_go_core()

func _on_back_to_menu_pressed():
	kill_core()
	get_tree().change_scene_to_file("res://level_select.tscn")

func _update_drones_visuals(drones: Array):
	# Swarm visuals are rendered by the drone_inspector screen via GameHub.
	# Legacy sprite is kept as a minimal fallback for the level-select view.
	if drones.size() > 0:
		var d = drones[0]
		agent_sprite.position = Vector2(d["x"] / 1000000.0, d["y"] / 1000000.0)
		var is_comp = d.get("comp", false)
		agent_sprite.modulate = Color(0.8, 0, 1, 1) if is_comp else Color(0, 1, 0.62, 1)

func _trigger_breakpoint_halt(drone_id: int) -> void:
	var inspector = get_node_or_null("/root/GameHub/InspectorModal")

	if inspector:
		print("[Breakpoint] Entity %d critical power threshold achieved. Halting core." % drone_id)
		inspector.inspect_entity_source(drone_id)

func kill_core():
	if active_pipe:
		if active_pipe.has("pid"):
			var pid = active_pipe["pid"]
			# Send SIGTERM (signal 15) so Go core can save replay before exit
			OS.kill(pid, 15)
		active_pipe = null

func _exit_tree():
	kill_core()

func _on_deploy_pressed():
	print("Deploying new script...")
	var new_code = code_editor.text
	var target_script_path = ProjectSettings.globalize_path("res://../core/scripts/agent.ps")

	var file = FileAccess.open(target_script_path, FileAccess.WRITE)
	if file:
		file.store_string(new_code)
		file.close()
		print("Script saved to disk: ", target_script_path)

func _on_back_pressed():
	kill_core()
	get_tree().change_scene_to_file("res://level_select.tscn")

func _unhandled_input(event):
	if event is InputEventKey and event.pressed and event.keycode == KEY_F11:
		var current_mode = DisplayServer.window_get_mode()
		if current_mode == DisplayServer.WINDOW_MODE_FULLSCREEN:
			DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_WINDOWED)
		else:
			DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_FULLSCREEN)
