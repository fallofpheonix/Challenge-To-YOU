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

var current_lang = "ps"
var current_script_path = ""
var active_pipe = null

var templates = {
	"ps": "fn main() {\n    HARVEST()\n    DROP_RESOURCE()\n    \n    if (SENSE_RESOURCE()) {\n        MOVE_TOWARDS_RESOURCE()\n    } else {\n        if (SENSE_HOME()) {\n            MOVE_TOWARDS_HOME()\n        } else {\n            MOVE_RANDOM()\n        }\n    }\n}\n",
	"py": "# Chrysalis Agent Loop\nwhile get_x() < get_goal_x():\n    move_forward()\n",
}

func _ready():
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
	start_go_core()

func setup_highlighter():
	var highlighter = CodeHighlighter.new()
	
	code_editor.add_theme_color_override("font_color", Color(0.9, 0.9, 0.9, 1))
	
	var kw_color = Color(0, 1, 1, 1)
	var keywords = ["fn", "let", "if", "else", "while"]
	for kw in keywords:
		highlighter.add_keyword_color(kw, kw_color)
		
	var api_color = Color(0, 1, 0.2, 1)
	var api_funcs = ["HARVEST", "DROP_RESOURCE", "SENSE_RESOURCE", "SENSE_HOME", "MOVE_RANDOM", "MOVE_TOWARDS_RESOURCE", "MOVE_TOWARDS_HOME"]
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
	var go_bin = "go"
	var core_path = ProjectSettings.globalize_path("res://../core/main.go")
	
	current_script_path = ProjectSettings.globalize_path("res://../core/scripts/agent.ps")

	OS.set_environment("PHX_SCRIPT_PATH", current_script_path)

	# In a production build, we'd use a compiled binary. For dev, go run is fine.
	active_pipe = OS.execute_with_pipe(go_bin, ["run", core_path], true)
	
	if active_pipe.has("stdio"):
		go_stdout = active_pipe["stdio"]
		print("Pipe opened successfully to Go core.")
	else:
		print("Failed to open pipe to Go core.")

func _process(_delta):
	if go_stdout and go_stdout.get_length() > 0:
		var line = go_stdout.get_line().strip_edges()
		if line.begins_with("{"):
			var json = JSON.new()
			var error = json.parse(line)
			if error == OK:
				var data = json.get_data()
				update_ui(data)
			else:
				print("JSON Parse Error: ", json.get_error_message(), " in line: ", line)

func update_ui(data):
	# Forward data to the Game Hub if it exists
	var hub = get_node_or_null("/root/GameHub")
	if hub:
		hub.forward_data(data)

	# 1. Handle Tick
	if data.has("tick"):
		telemetry_label.text = "TICK: %03d" % data["tick"]

	# 2. Handle Drones
	if data.has("drones"):
		# Update legacy sprites for first drone
		if data["drones"].size() > 0:
			var d = data["drones"][0]
			agent_sprite.position = Vector2(d["x"] / 1000000.0, d["y"] / 1000000.0)
		
		# Breakpoint condition: Drone battery critically low inside anomaly zone
		for drone in data["drones"]:
			if drone["state"] == 0 and drone["bat"] < 25000000: 
				_trigger_breakpoint_halt(drone["id"])
				return

func _trigger_breakpoint_halt(drone_id: int) -> void:
	var inspector = get_node_or_null("/root/GameHub/InspectorModal")
	if not inspector:
		# Fallback to local if not in Hub
		inspector = get_node_or_null("UI/InspectorModal")
	
	if inspector:
		print("[Breakpoint] Entity %d critical power threshold achieved. Halting core." % drone_id)
		inspector.inspect_entity_source(drone_id)

func kill_core():
	if active_pipe:
		if active_pipe.has("stdio") and is_instance_valid(active_pipe["stdio"]):
			active_pipe["stdio"].close()
		if active_pipe.has("pid"):
			var pid = active_pipe["pid"]
			OS.kill(pid)
		active_pipe = null

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
