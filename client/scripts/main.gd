extends Control

@onready var output_log: RichTextLabel = %OutputLog
@onready var command_line: LineEdit = %CommandLine
@onready var seed_edit: LineEdit = %SeedEdit
@onready var luck_spin: SpinBox = %LuckSpin
@onready var paradigm_option: OptionButton = %ParadigmOption
@onready var connect_button: Button = %ConnectButton
@onready var glitch_overlay: ColorRect = %GlitchOverlay
@onready var network_bridge: Node = get_node("/root/NetworkBridge")

var _current_challenge_id: String = ""
var _last_vigilance: float = 0.0
var _glitch_spike: float = 0.0

func _ready() -> void:
	network_bridge.fabric_updated.connect(_on_fabric_updated)
	network_bridge.ontological_purge.connect(_on_purge)
	network_bridge.archon_transmission.connect(_on_archon_transmission)
	command_line.text_submitted.connect(_on_command_submitted)
	connect_button.pressed.connect(_on_connect_pressed)
	command_line.grab_focus()
	_append("[color=#00ff88]TERMINAL READY. TYPE 'help' FOR AVAILABLE OVERRIDES.[/color]")
	_update_shader_parameters()

func _process(delta: float) -> void:
	if _glitch_spike > 0.0:
		_glitch_spike = max(0.0, _glitch_spike - delta * 1.5)
		_update_shader_parameters()

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

func _on_fabric_updated(payload: Dictionary) -> void:
	var challenge_id = payload.get("challenge_id", "")
	if challenge_id != _current_challenge_id:
		_current_challenge_id = challenge_id
		output_log.clear()
		_append("[color=#00ff88]NEW RIFT OPENED: %s[/color]" % [payload.get("title", "Challenge")])
		_append("[color=#88aa88]%s[/color]" % [payload.get("description", "")])
		
	var state: Dictionary = payload.get("state", {})
	var message: String = payload.get("message", payload.get("error_message", ""))
	var triggerable: Array = payload.get("triggerable", [])
	_last_vigilance = float(payload.get("vigilance", 0.0))
	var cipher: String = payload.get("last_cipher", "")
	var complete: bool = payload.get("level_complete", false)
	
	if payload.get("error_message", "") != "":
		_glitch_spike = 0.8
	else:
		_glitch_spike = 0.3
		
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

func _on_purge(reason: String) -> void:
	_append("")
	_append("[color=#ff0000][b]ONTOLOGICAL PURGE: %s[/b][/color]" % [reason])
	_last_vigilance = 1.0
	_glitch_spike = 1.0
	_update_shader_parameters()

func _on_command_submitted(text: String) -> void:
	var cmd = text.strip().to_lower()
	if cmd == "":
		return
	
	_append("[color=#00ff88]root@fabric:~#[/color] %s" % [text])
	command_line.clear()
	
	if cmd == "help" or cmd == "h":
		_show_help()
	elif cmd == "clear" or cmd == "cls":
		output_log.clear()
		_append("[color=#00ff88]TERMINAL CLEARED.[/color]")
	elif cmd == "mending":
		_append("[color=#ffaa00]INITIATING MENDING PROTOCOL OVERRIDE... (Entropy Cost: +30)[/color]")
		_glitch_spike = 0.7
		network_bridge.transmit_event("mending_protocol")
	elif cmd == "profile":
		_append("[color=#00ffff]QUERYING PLAYER PROFILE METRICS...[/color]")
		network_bridge.transmit_event("profile")
	elif cmd.begins_with("unlock "):
		var target = cmd.trim_prefix("unlock ").strip_edges()
		_append("[color=#00ffff]TRANSMITTING PARADIGM DE-SEGREGATION COMMAND FOR: %s...[/color]" % [target])
		network_bridge.transmit_event(cmd)
	else:
		network_bridge.transmit_event(cmd)

func _show_help() -> void:
	_append("[color=#ffff00]AVAILABLE COMMANDS:[/color]")
	_append("  [color=#00ffff]help[/color]           - Show this help")
	_append("  [color=#00ffff]clear[/color]          - Clear terminal")
	_append("  [color=#00ffff]profile[/color]        - Show player reputation, luck, and unlocks")
	_append("  [color=#00ffff]unlock <era>[/color]   - Spend reputation to unlock era (cyberpunk/cosmic)")
	_append("  [color=#00ffff]mending[/color]        - Trigger active AI repair protocol (Entropy Cost: +30)")
	_append("  [color=#00ffff]<event_id>[/color]     - Transmit event to fabric (e.g. inject_adrenaline, cut_power)")

func _append(line: String) -> void:
	output_log.append_text(line + "\n")
	output_log.scroll_to_end()

func _on_archon_transmission(message: String) -> void:
	_append("")
	_append("[color=#ff3333][b][ARCHON DEEP VIGILANCE][/b][/color]")
	_append("[color=#ff5555][i]%s[/i][/color]" % [message])
	_glitch_spike = 0.5
	_update_shader_parameters()