extends Control

var levels_dir = "res://../core/levels/"

func _ready():
	var hub = get_node_or_null("/root/GameHub")
	if hub:
		hub.hide()

	# Scan the levels directory and create buttons
	var dir = DirAccess.open(ProjectSettings.globalize_path(levels_dir))
	if dir:
		dir.list_dir_begin()
		var file_name = dir.get_next()
		while file_name != "":
			if not dir.current_is_dir() and file_name.ends_with(".json"):
				create_level_button(file_name)
			file_name = dir.get_next()
	else:
		print("An error occurred when trying to access the levels path.")

func create_level_button(file_name: String):
	var path = ProjectSettings.globalize_path(levels_dir + file_name)
	var title = file_name
	
	var file = FileAccess.open(path, FileAccess.READ)
	if file:
		var content = file.get_as_text()
		var json = JSON.new()
		if json.parse(content) == OK:
			var data = json.get_data()
			if data.has("title"): title = data["title"]
	
	var card = PanelContainer.new()
	card.custom_minimum_size = Vector2(300, 100)
	
	var style = StyleBoxFlat.new()
	style.bg_color = Color(0.1, 0.1, 0.1, 1.0)
	style.border_color = Color(0, 1, 1, 0.5)
	style.border_width_bottom = 2
	style.border_width_top = 2
	style.border_width_left = 2
	style.border_width_right = 2
	card.add_theme_stylebox_override("panel", style)
	
	var vbox = VBoxContainer.new()
	card.add_child(vbox)
	
	var header = Label.new()
	header.text = "TIER 1 · " + file_name.replace(".json", "").to_upper()
	header.add_theme_color_override("font_color", Color(0.5, 0.5, 0.5, 1))
	vbox.add_child(header)
	
	var title_lbl = Label.new()
	title_lbl.text = title
	title_lbl.add_theme_color_override("font_color", Color(0, 1, 1, 1))
	vbox.add_child(title_lbl)
	
	var play_btn = Button.new()
	play_btn.text = "PLAY →"
	play_btn.add_theme_color_override("font_color", Color(0, 1, 0, 1))
	play_btn.pressed.connect(self._on_level_button_pressed.bind(path))
	vbox.add_child(play_btn)
	
	$VBoxContainer/ScrollContainer/LevelList.add_child(card)

func _on_level_button_pressed(level_path: String):
	# Store the selected level path globally or pass it to the main scene
	# For simplicity, we'll use ProjectSettings as a temporary global store
	# before switching scenes.
	ProjectSettings.set_setting("game/current_level_path", level_path)
	
	# Transition to the main simulation scene
	get_tree().change_scene_to_file("res://main.tscn")

func toggle_fullscreen():
	var current_mode = DisplayServer.window_get_mode()
	if current_mode == DisplayServer.WINDOW_MODE_FULLSCREEN:
		DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_WINDOWED)
	else:
		DisplayServer.window_set_mode(DisplayServer.WINDOW_MODE_FULLSCREEN)

func _unhandled_input(event):
	if event is InputEventKey and event.pressed and event.keycode == KEY_F11:
		toggle_fullscreen()
