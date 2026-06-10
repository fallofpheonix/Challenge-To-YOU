extends PanelContainer

@onready var code_buffer: CodeEdit = $VBoxContainer/CodeBuffer
@onready var lock_btn: Button = $VBoxContainer/ActionControls/LockToggleBtn
@onready var apply_btn: Button = $VBoxContainer/ActionControls/ApplyBtn
@onready var id_label: Label = $VBoxContainer/HeaderPanel/DroneIDLabel

var current_drone_id: int = -1
var is_immutable: bool = true

func _ready() -> void:
	hide()
	lock_btn.pressed.connect(_on_lock_toggled)
	apply_btn.pressed.connect(_on_apply_patch)

# Called when an active drone icon 'd' is selected on the macro view canvas
func inspect_entity_source(drone_id: int) -> void:
	current_drone_id = drone_id
	id_label.text = "ENTITY_ID: %03d" % drone_id
	
	# Fetch the active code buffer file from disk
	var script_path = OS.get_environment("PHX_SCRIPT_PATH")
	if script_path == "" or not FileAccess.file_exists(script_path):
		# Fallback for development if env var not set
		script_path = "res://../core/scripts/agent.ps"
		
	var file = FileAccess.open(script_path, FileAccess.READ)
	if file:
		code_buffer.text = file.get_as_text()
		file.close()
	else:
		code_buffer.text = "// Error: Could not load source from " + script_path
	
	# Enforce immutable security configuration by default
	is_immutable = true
	code_buffer.editable = false
	lock_btn.text = "🔒 Code Locked (Read-Only)"
	apply_btn.disabled = true
	
	# Halt client processing steps to preserve the exact tick breakpoint state
	get_tree().paused = true
	show()

func _on_lock_toggled() -> void:
	is_immutable = !is_immutable
	code_buffer.editable = !is_immutable
	apply_btn.disabled = is_immutable
	
	if is_immutable:
		lock_btn.text = "🔒 Code Locked (Read-Only)"
	else:
		lock_btn.text = "🔓 Code Unlocked (Editable)"

func _on_apply_patch() -> void:
	var script_path = OS.get_environment("PHX_SCRIPT_PATH")
	if script_path == "":
		script_path = "res://../core/scripts/agent.ps"
		
	# Overwrite the script file directly
	var file = FileAccess.open(script_path, FileAccess.WRITE)
	if file:
		file.store_string(code_buffer.text)
		file.close()
		print("[Telemetry] Patch deployed for entity group. Resuming core loop.")
	
	# Unpause the simulation and hide modal window
	get_tree().paused = false
	hide()
