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
	set_language("ps")
	
	# Initialize Networking
	network_bridge = preload("res://network_bridge.gd").new()
	network_bridge.name = "NetworkBridge"
	add_child(network_bridge)
	network_bridge.packet_received.connect(_on_network_packet)
	
	start_go_core()

func setup_highlighter():
	var highlighter = CodeHighlighter.new()
	
	code_editor.add_theme_color_override("font_color", Color(0.9, 0.9, 0.9, 1))
	
	var kw_color = Color(0, 1, 1, 1)
	var keywords = ["fn", "let", "if", "else", "while"]
	for kw in keywords:
		highlighter.add_keyword_color(kw, kw_color)
		
	var api_color = Color(0, 1, 0.2, 1)
	var api_funcs = ["HARVEST", "DROP_RESOURCE", "SENSE_RESOURCE", "SENSE_HOME", "SENSE_CARGO", "MOVE_RANDOM", "MOVE_TOWARDS_RESOURCE", "MOVE_TOWARDS_HOME"]
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

	if not FileAccess.file_exists(core_bin):
		push_error("[Network Error] Missing core binary. Run `make build-core` from engine/.")
		return

	var pid = OS.create_process(core_bin, [])
	if pid != -1:
		active_pipe = {"pid": pid}
		print("[Network] Launched Go core (PID: %d). Awaiting WebSocket handshake..." % pid)
	else:
		print("[Network Error] Failed to launch Go core.")

func _on_network_packet(type: String, data: Dictionary):
	if type == "EMISSION_SNAPSHOT":
		update_ui(data["payload"])

func _process(_delta):
	# Handled by network_bridge.gd asynchronously
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
		return

	mission_terminal = true
	var reason = str(mission.get("reason", "")).replace("_", " ").to_upper()
	if status == "victory":
		complete_label.text = "VICTORY"
		complete_label.modulate = Color(0, 1, 0.62, 1)
	elif status == "defeat":
		complete_label.text = "DEFEAT"
		complete_label.modulate = Color(1, 0.2, 0.2, 1)
	else:
		complete_label.text = str(status).to_upper()
		complete_label.modulate = Color(1, 1, 1, 1)

	if not reason.is_empty():
		complete_label.text += "\n%s" % reason
	complete_label.show()

func _update_drones_visuals(drones: Array):
	# Clear old legacy trail logic or sprites if needed
	# For the MVP, we'll just update the first sprite and then handle a swarm layer
	if drones.size() > 0:
		var d = drones[0]
		agent_sprite.position = Vector2(d["x"] / 1000000.0, d["y"] / 1000000.0)
		
		# Tint based on corruption
		var corr = d.get("corr", 0) / 100.0
		var is_comp = d.get("comp", false)
		
		if is_comp:
			agent_sprite.modulate = Color(0.8, 0, 1, 1) # Viral Purple
		else:
			# Lerp from Clean Cyan (0,1,0.62) to Warning Pink/Purple
			agent_sprite.modulate = Color(0, 1, 0.62).lerp(Color(0.8, 0, 1), corr)

func _trigger_breakpoint_halt(drone_id: int) -> void:
	var inspector = get_node_or_null("/root/GameHub/InspectorModal")
	
	if inspector:
		print("[Breakpoint] Entity %d critical power threshold achieved. Halting core." % drone_id)
		inspector.inspect_entity_source(drone_id)

func kill_core():
	if active_pipe:
		if active_pipe.has("pid"):
			var pid = active_pipe["pid"]
			OS.kill(pid)
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
	
	# The Go core reloads automatically on file change
	# but we can reset if needed
	# _on_reset_pressed()

func _on_back_pressed():
	kill_core()
	get_tree().change_scene_to_file("res://level_select.tscn")

func _on_reset_pressed():
	kill_core()
	start_go_core()

func _unhandled_input(event):
	if event is InputEventKey and event.pressed and event.keycode == KEY_F11:
		var current_mode = DisplayServer.window_get_mode()
		if current_mode == DisplayServer.WINDOW_MODE_FULLSCREEN:
			DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_WINDOWED)
		else:
			DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_FULLSCREEN)
