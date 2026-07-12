extends Control

# UI Elements
@onready var output_log: RichTextLabel = %OutputLog
@onready var command_line: LineEdit = %CommandLine
@onready var code_panel: VBoxContainer = %CodePanel
@onready var code_editor: TextEdit = %CodeEditor
@onready var code_submit_btn: Button = %CodeSubmitBtn
@onready var code_label: Label = %CodeLabel
@onready var seed_edit: LineEdit = %SeedEdit
@onready var luck_spin: SpinBox = %LuckSpin
@onready var paradigm_option: OptionButton = %ParadigmOption
@onready var connect_button: Button = %ConnectButton
@onready var glitch_overlay: ColorRect = %GlitchOverlay
@onready var network_bridge: Node = get_node("/root/NetworkBridge")

# Main Menu UI Nodes
@onready var main_menu_panel: PanelContainer = %MainMenuPanel
@onready var play_btn: Button = %PlayBtn
@onready var continue_btn: Button = %ContinueBtn
@onready var credits_btn: Button = %CreditsBtn
@onready var exit_btn: Button = %ExitBtn
@onready var save_slot_label: Label = %SaveSlotLabel

# Dialogue UI Nodes
@onready var dialogue_panel: PanelContainer = %DialoguePanel
@onready var speaker_name: Label = %SpeakerName
@onready var dialogue_text: Label = %DialogueText
@onready var dialogue_choices: HBoxContainer = %DialogueChoices
@onready var dialogue_next_btn: Button = %DialogueNextBtn

# HUD UI Nodes
@onready var hud_panel: HBoxContainer = %HUDPanel
@onready var hud_reputation: Label = %HUDReputation
@onready var hud_luck: Label = %HUDLuck
@onready var hud_unlocked_paradigms: Label = %HUDUnlockedParadigms

# Sidebar UI Nodes
@onready var room_name_label: Label = %RoomNameLabel
@onready var room_desc_label: Label = %RoomDescLabel
@onready var room_exits_container: VBoxContainer = %RoomExitsContainer
@onready var room_objects_container: VBoxContainer = %RoomObjectsContainer
@onready var inventory_container: VBoxContainer = %InventoryContainer
@onready var journal_container: VBoxContainer = %JournalContainer

var _current_challenge_id: String = ""
var _current_skill_type: String = ""
var _last_vigilance: float = 0.0
var _glitch_spike: float = 0.0

# Dialogue State
var _active_dialogue_lines: Array = []
var _current_dialogue_index: int = 0
var _typewriter_timer: float = 0.0
var _target_dialogue_text: String = ""

func _ready() -> void:
	network_bridge.fabric_updated.connect(_on_fabric_updated)
	network_bridge.ontological_purge.connect(_on_purge)
	network_bridge.archon_transmission.connect(_on_archon_transmission)
	command_line.text_submitted.connect(_on_command_submitted)
	connect_button.pressed.connect(_on_connect_pressed)
	code_submit_btn.pressed.connect(_on_code_submitted)

	# Main Menu Buttons Bindings
	play_btn.pressed.connect(_on_play_pressed)
	continue_btn.pressed.connect(_on_continue_pressed)
	credits_btn.pressed.connect(_on_credits_pressed)
	exit_btn.pressed.connect(_on_exit_pressed)

	# Dialogue Next Button Bindings
	dialogue_next_btn.pressed.connect(_on_dialogue_next_pressed)

	command_line.grab_focus()
	code_panel.visible = false
	dialogue_panel.visible = false
	output_log.scroll_following = true

	# Initially show the main menu
	main_menu_panel.visible = true
	hud_panel.visible = false

	_append("[color=#00ff88]TERMINAL READY. AWAITING SESSION INITIALIZATION.[/color]")
	_update_shader_parameters()

func _process(delta: float) -> void:
	if _glitch_spike > 0.0:
		_glitch_spike = max(0.0, _glitch_spike - delta * 1.5)
		_update_shader_parameters()

	# Typewriter Animation Effect
	if dialogue_panel.visible and dialogue_text.text.length() < _target_dialogue_text.length():
		_typewriter_timer += delta
		if _typewriter_timer > 0.03:
			_typewriter_timer = 0.0
			dialogue_text.text = _target_dialogue_text.left(dialogue_text.text.length() + 1)
			play_synth_tone(800.0, 0.01)

func _update_shader_parameters() -> void:
	if not glitch_overlay or not glitch_overlay.material:
		return
	
	var mat = glitch_overlay.material as ShaderMaterial
	var aberration_val = 0.001 + _last_vigilance * 0.012 + _glitch_spike * 0.015
	var glitch_val = _last_vigilance * 0.3 + _glitch_spike * 0.7
	var scanline_val = 0.12 + _last_vigilance * 0.18
	
	mat.set_shader_parameter("aberration", aberration_val)
	mat.set_shader_parameter("glitch_intensity", clamp(glitch_val, 0.0, 1.0))
	mat.set_shader_parameter("scanline_strength", clamp(scanline_val, 0.0, 1.0))

# --- Main Menu Options ---

func _on_play_pressed() -> void:
	main_menu_panel.visible = false
	hud_panel.visible = true
	_glitch_spike = 0.6
	_append("[color=#00ff88]Rift initialization triggered...[/color]")
	# Connect with default seed and luck values
	_on_connect_pressed()

func _on_continue_pressed() -> void:
	main_menu_panel.visible = false
	hud_panel.visible = true
	_glitch_spike = 0.4
	_append("[color=#00ff88]Resuming existing campaign profile...[/color]")
	network_bridge.connect_to_rift(-1, 1.0, "")

func _on_credits_pressed() -> void:
	_append("")
	_append("[color=#ffff00]=== CREDITS ===[/color]")
	_append("  [color=#00ffff]Challenge To YOU Team[/color]")
	_append("  Lead Worldbuilding Director, Systems Lore & Technical Lead")
	_append("  Powered by Godot 4 & Go Axiomatic Engine.")
	_append("")

func _on_exit_pressed() -> void:
	get_tree().quit()

# --- Connection Handler ---

func _on_connect_pressed() -> void:
	var seed_text = seed_edit.text.strip_edges()
	var seed_val = -1
	if seed_text != "":
		if seed_text.is_valid_int():
			seed_val = int(seed_text)
		else:
			seed_val = hash(seed_text) & 0x7FFFFFFF
	else:
		seed_val = randi() & 0x7FFFFFFF
		seed_edit.text = str(seed_val)

	var luck_val = luck_spin.value
	var p_idx = paradigm_option.selected
	var paradigm_val = paradigm_option.get_item_text(p_idx)

	_append("")
	_append("[color=#ffff00]ESTABLISHING RIFT WITH TEMPORAL COORDINATES:[/color]")
	_append("  [color=#00ffff]Seed:[/color] %d" % [seed_val])
	_append("  [color=#00ffff]Luck:[/color] %.2f" % [luck_val])
	_append("  [color=#00ffff]Era:[/color] %s" % [paradigm_val])
	
	_glitch_spike = 0.6
	network_bridge.connect_to_rift(seed_val, luck_val, paradigm_val)

# --- Dialogue System ---

func start_dialogue(lines: Array) -> void:
	_active_dialogue_lines = lines
	_current_dialogue_index = 0
	dialogue_panel.visible = true
	_display_current_dialogue()

func _display_current_dialogue() -> void:
	if _current_dialogue_index >= _active_dialogue_lines.size():
		dialogue_panel.visible = false
		_append("[color=#00ff88]Dialogue sequence complete. Initiating challenge terminal...[/color]")
		return

	var line = _active_dialogue_lines[_current_dialogue_index]
	speaker_name.text = line.get("speaker", "Speaker")
	_target_dialogue_text = line.get("text", "")
	dialogue_text.text = ""
	_typewriter_timer = 0.0

	# Clear previous choice buttons
	for child in dialogue_choices.get_children():
		child.queue_free()

	# If choices are present, display option buttons
	var choices = line.get("choices", [])
	if choices.size() > 0:
		dialogue_next_btn.visible = false
		for i in range(choices.size()):
			var choice = choices[i]
			var btn = Button.new()
			btn.text = choice.get("text", "")
			btn.pressed.connect(func(): _on_dialogue_choice_pressed(choice.get("event", "")))
			dialogue_choices.add_child(btn)
	else:
		dialogue_next_btn.visible = true

func _on_dialogue_next_pressed() -> void:
	# If typewriter is still animating, complete it immediately
	if dialogue_text.text.length() < _target_dialogue_text.length():
		dialogue_text.text = _target_dialogue_text
		return
	
	_current_dialogue_index += 1
	_display_current_dialogue()

func _on_dialogue_choice_pressed(event_id: String) -> void:
	dialogue_panel.visible = false
	if event_id != "":
		network_bridge.transmit_event(event_id)

# --- WebSocket Inbound Handlers ---

func _on_fabric_updated(payload: Dictionary) -> void:
	# Check for dialogue / cinematic sequences inside payload
	if payload.has("dialogue"):
		var raw_dialogue = payload.get("dialogue")
		if raw_dialogue is Array:
			start_dialogue(raw_dialogue)
	elif payload.has("intro") and payload.get("intro").has("scenes"):
		var scenes = payload.get("intro").get("scenes", [])
		if scenes.size() > 0:
			var dialogue_lines = []
			for scene in scenes:
				for line in scene.get("dialogue", []):
					dialogue_lines.append(line)
			if dialogue_lines.size() > 0:
				start_dialogue(dialogue_lines)

	# Update HUD stats
	if payload.has("profile"):
		var prof = payload.get("profile")
		hud_reputation.text = "Reputation: %d" % [prof.get("reputation", 0)]
		hud_luck.text = "Luck: %.2f" % [prof.get("luck", 1.0)]
		hud_unlocked_paradigms.text = "Eras Unlocked: %s" % [prof.get("unlocked_paradigms", "MAGITECH")]

	var challenge_id = payload.get("challenge_id", "")
	var skill_type: String = payload.get("skill_type", "")
	
	if challenge_id != _current_challenge_id:
		_current_challenge_id = challenge_id
		_current_skill_type = skill_type
		output_log.clear()
		_append("[color=#00ff88]NEW RIFT OPENED: %s[/color]" % [payload.get("title", "Challenge")])
		_append("[color=#88aa88]%s[/color]" % [payload.get("description", "")])
		_update_input_mode(skill_type)
		
	var state: Dictionary = payload.get("state", {})
	var message: String = payload.get("message", payload.get("error_message", ""))
	var triggerable: Array = payload.get("triggerable", [])
	_last_vigilance = float(payload.get("vigilance", 0.0))
	var cipher: String = payload.get("last_cipher", "")
	var complete: bool = payload.get("level_complete", false)
	
	if payload.get("error_message", "") != "":
		_glitch_spike = 0.8
		play_synth_tone(150.0, 0.25)
	else:
		_glitch_spike = 0.3
		play_synth_tone(1200.0, 0.02)
		
	_update_shader_parameters()
	
	_append("")
	_append("[color=#ffff00]=== STATE UPDATE ===[/color]")
	for key in state.keys():
		var val = state[key]
		_append("  [color=#00ffff]%s[/color] = %s" % [key, str(val)])
	
	if triggerable.size() > 0:
		_append("[color=#ffaa00]TRIGGERABLE: %s[/color]" % [str(triggerable)])
	
	if message != "":
		_append("[color=#ffffff]> %s[/color]" % [message])
	
	if _last_vigilance > 0:
		_append("[color=#ff4444]VIGILANCE: %.1f%%[/color]" % [_last_vigilance * 100])
	
	if cipher != "":
		_append("[color=#00ff00]CIPHER FRAGMENT: %s[/color]" % [cipher])
	
	if complete:
		_append("[color=#00ff00][b]LEVEL COMPLETE — LOGOS TOKEN: %s[/b][/color]" % [cipher])
		_glitch_spike = 0.5
		code_panel.visible = false
		play_synth_tone(523.25, 0.1)
		await get_tree().create_timer(0.12).timeout
		play_synth_tone(659.25, 0.15)

	_update_exploration_sidebar(payload)

func _update_exploration_sidebar(payload: Dictionary) -> void:
	if not payload.has("world") or not payload.has("state"):
		return
	
	var world = payload.get("world")
	var state = payload.get("state")
	var current_room = state.get("current_room", "Rune Chamber")
	
	room_name_label.text = current_room
	
	# Clear previous controls
	for child in room_exits_container.get_children():
		child.queue_free()
	for child in room_objects_container.get_children():
		child.queue_free()
	for child in inventory_container.get_children():
		child.queue_free()
	for child in journal_container.get_children():
		child.queue_free()
		
	# Populate current room info
	if world.has("rooms") and world.get("rooms").has(current_room):
		var room_info = world.get("rooms").get(current_room)
		room_desc_label.text = room_info.get("description", "")
		
		# Exits
		for exit in room_info.get("exits", []):
			var btn = Button.new()
			btn.text = "Exit to: " + exit
			btn.pressed.connect(func():
				network_bridge.transmit_event("move_room:" + exit)
			)
			room_exits_container.add_child(btn)
			
		# Inspectables
		for object in room_info.get("objects", []):
			if object == "fractured_ward" and not state.get("has_rune_shard", false):
				continue
			var btn = Button.new()
			btn.text = "Inspect: " + object
			btn.pressed.connect(func():
				network_bridge.transmit_event("inspect_object:" + object)
				if object == "fractured_ward":
					code_panel.visible = true
			)
			room_objects_container.add_child(btn)

	# Inventory Display
	if state.get("has_rune_shard", false):
		var lbl = Label.new()
		lbl.text = "- Rune Shard (Quest Item)"
		inventory_container.add_child(lbl)
	else:
		var lbl = Label.new()
		lbl.text = "Inventory is empty."
		inventory_container.add_child(lbl)

	# Journal / Objectives
	var objective_label = Label.new()
	if not state.get("has_rune_shard", false):
		objective_label.text = "[ ] Find the Rune Shard in the Forbidden Library."
	elif state.get("ward_sealed", true):
		objective_label.text = "[ ] Inspect Fractured Ward in Axiomatic Ward and trigger the release sequence."
	else:
		objective_label.text = "[x] Fractured Ward repaired!"
	journal_container.add_child(objective_label)

func _update_input_mode(skill_type: String) -> void:
	match skill_type:
		"recognize":
			code_panel.visible = false
			command_line.placeholder_text = "Submit answer: submit_answer:<line_number>"
		"optimize":
			code_panel.visible = true
			code_label.text = "OPTIMIZE CHALLENGE — Paste your optimized function below:"
			code_editor.placeholder_text = "function fib(n) {\n    // Your O(N) or better solution here\n}"
			command_line.placeholder_text = "Or type: execute_script to submit the code above"
		"write_from_spec":
			code_panel.visible = true
			code_label.text = "WRITE FROM SPEC — Implement the function below:"
			code_editor.placeholder_text = "function solution(...) {\n    // Your implementation here\n}"
			command_line.placeholder_text = "Or type: execute_script to submit the code above"
		_:
			code_panel.visible = false
			command_line.placeholder_text = "Enter command (help, inject_adrenaline, cut_power, etc.)"

func _on_purge(reason: String) -> void:
	_append("")
	_append("[color=#ff0000][b]ONTOLOGICAL PURGE: %s[/b][/color]" % [reason])
	_last_vigilance = 1.0
	_glitch_spike = 1.0
	_update_shader_parameters()

# --- Code Submission ---

func _on_code_submitted() -> void:
	var code = code_editor.text.strip_edges()
	if code == "":
		_append("[color=#ff4444]CODE PANEL: Nothing to submit — editor is empty.[/color]")
		return
	_append("[color=#00ff88]root@fabric:~#[/color] [EXECUTE_SCRIPT]")
	_append("[color=#888888]> Submitting %d chars to sandbox...[/color]" % [code.length()])
	_glitch_spike = 0.4
	network_bridge.transmit_event_with_payload("execute_script", code)

func _on_command_submitted(text: String) -> void:
	var cmd = text.strip_edges()
	if cmd == "":
		return
	
	_append("[color=#00ff88]root@fabric:~#[/color] %s" % [text])
	command_line.clear()
	
	if ":" in cmd:
		var parts = cmd.split(":", false, 1)
		var event_id = parts[0].strip_edges()
		var payload = parts[1].strip_edges() if parts.size() > 1 else ""
		_append("[color=#00ffff]TRANSMITTING: event=%s payload=%s[/color]" % [event_id, payload.left(40)])
		_glitch_spike = 0.3
		network_bridge.transmit_event_with_payload(event_id, payload)
		return
	
	if cmd.to_lower() == "execute_script" and code_panel.visible:
		_on_code_submitted()
		return
	
	if cmd.to_lower() == "help" or cmd.to_lower() == "h":
		_show_help()
	elif cmd.to_lower() == "clear" or cmd.to_lower() == "cls":
		output_log.clear()
		_append("[color=#00ff88]TERMINAL CLEARED.[/color]")
	elif cmd.to_lower() == "mending":
		_append("[color=#ffaa00]INITIATING MENDING PROTOCOL OVERRIDE... (Entropy Cost: +30)[/color]")
		_glitch_spike = 0.7
		network_bridge.transmit_event("mending_protocol")
	elif cmd.to_lower() == "profile":
		_append("[color=#00ffff]QUERYING PLAYER PROFILE METRICS...[/color]")
		network_bridge.transmit_event("profile")
	elif cmd.to_lower().begins_with("unlock "):
		var target = cmd.trim_prefix("unlock ").strip_edges()
		_append("[color=#00ffff]TRANSMITTING PARADIGM DE-SEGREGATION COMMAND FOR: %s...[/color]" % [target])
		network_bridge.transmit_event(cmd)
	else:
		network_bridge.transmit_event(cmd)

func _show_help() -> void:
	_append("[color=#ffff00]AVAILABLE COMMANDS:[/color]")
	_append("  [color=#00ffff]help[/color]                    - Show this help")
	_append("  [color=#00ffff]clear[/color]                   - Clear terminal")
	_append("  [color=#00ffff]profile[/color]                 - Show player reputation, luck, and unlocks")
	_append("  [color=#00ffff]unlock <era>[/color]            - Spend reputation to unlock era (cyberpunk/cosmic)")
	_append("  [color=#00ffff]mending[/color]                 - Trigger active AI repair protocol (Entropy Cost: +30)")
	_append("  [color=#00ffff]submit_answer:<value>[/color]   - Submit an answer for Recognize challenges")
	_append("  [color=#00ffff]execute_script[/color]          - Submit code from the code panel (Optimize / Write-from-spec)")
	_append("  [color=#00ffff]<event_id>[/color]              - Transmit event to fabric (e.g. inject_adrenaline, cut_power)")
	_append("  [color=#888888]TIP: Use event_id:payload syntax for raw payload injection[/color]")

func _append(line: String) -> void:
	output_log.append_text(line + "\n")

func _on_archon_transmission(message: String) -> void:
	_append("")
	_append("[color=#ff3333][b][ARCHON DEEP VIGILANCE][/b][/color]")
	_append("[color=#ff5555][i]%s[/i][/color]" % [message])
	_glitch_spike = 0.5
	_update_shader_parameters()

func play_synth_tone(frequency: float, duration: float) -> void:
	var player = AudioStreamPlayer.new()
	add_child(player)
	
	var generator = AudioStreamGenerator.new()
	generator.mix_rate = 22050
	generator.buffer_length = duration
	
	player.stream = generator
	player.play()
	
	var playback = player.get_stream_playback()
	if playback:
		var sample_count = int(generator.mix_rate * duration)
		var phase = 0.0
		var phase_step = frequency * 2.0 * PI / generator.mix_rate
		for i in range(sample_count):
			var sample = sin(phase) * 0.15
			playback.push_frame(Vector2(sample, sample))
			phase += phase_step
			
	await get_tree().create_timer(duration + 0.1).timeout
	player.queue_free()