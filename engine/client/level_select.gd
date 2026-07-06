extends Control

var levels_dir = "res://../core/levels/"
var completed_levels: Dictionary = {}

func _ready():
	var hub = get_node_or_null("/root/GameHub")
	if hub:
		hub.hide()

	_load_progress()
	_populate_levels()

func _load_progress():
	var progress_path = ProjectSettings.globalize_path(levels_dir + "progress.json")
	if not FileAccess.file_exists(progress_path):
		return
	var file = FileAccess.open(progress_path, FileAccess.READ)
	if file:
		var json = JSON.new()
		if json.parse(file.get_as_text()) == OK:
			var data = json.get_data()
			completed_levels = data.get("completed", {})

func _populate_levels():
	var level_list = $VBoxContainer/ScrollContainer/LevelList
	for child in level_list.get_children():
		child.queue_free()

	var levels = [
		{"file": "chrysalis_1.json", "id": "chrysalis_1", "tier": "I", "requires": []},
		{"file": "chrysalis_2.json", "id": "chrysalis_2", "tier": "II", "requires": ["chrysalis_1"]},
		{"file": "chrysalis_3.json", "id": "chrysalis_3", "tier": "III", "requires": ["chrysalis_2"]},
	]

	for level in levels:
		var unlocked = true
		for req in level["requires"]:
			if not completed_levels.has(req):
				unlocked = false
				break

		var path = ProjectSettings.globalize_path(levels_dir + level["file"])
		var title = level["file"].replace(".json", "").replace("_", " ").to_upper()
		var desc = ""

		var file = FileAccess.open(path, FileAccess.READ)
		if file:
			var json = JSON.new()
			if json.parse(file.get_as_text()) == OK:
				var data = json.get_data()
				if data.has("title"): title = data["title"]
				if data.has("description"): desc = data["description"]

		var card = PanelContainer.new()
		card.custom_minimum_size = Vector2(500, 0)

		var style = StyleBoxFlat.new()
		if unlocked:
			style.bg_color = Color(0.08, 0.12, 0.15, 1.0)
			style.border_color = Color(0, 1, 1, 0.5)
		else:
			style.bg_color = Color(0.06, 0.06, 0.06, 1.0)
			style.border_color = Color(0.3, 0.3, 0.3, 0.3)
		style.border_width_bottom = 2
		style.border_width_top = 2
		style.border_width_left = 2
		style.border_width_right = 2
		style.content_margin_left = 12
		style.content_margin_top = 8
		style.content_margin_right = 12
		style.content_margin_bottom = 8
		card.add_theme_stylebox_override("panel", style)

		var hbox = HBoxContainer.new()
		hbox.theme_override_constants/separation = 16
		card.add_child(hbox)

		var info = VBoxContainer.new()
		info.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		hbox.add_child(info)

		var tier_label = Label.new()
		tier_label.text = "TIER %s" % level["tier"]
		tier_label.add_theme_font_size_override("font_size", 10)
		tier_label.add_theme_color_override("font_color", Color(0.5, 0.5, 0.5, 1))
		info.add_child(tier_label)

		var title_lbl = Label.new()
		title_lbl.text = title
		title_lbl.add_theme_font_size_override("font_size", 16)
		title_lbl.add_theme_color_override("font_color", Color(0, 1, 1, 1) if unlocked else Color(0.4, 0.4, 0.4, 1))
		info.add_child(title_lbl)

		if desc != "":
			var desc_lbl = Label.new()
			desc_lbl.text = desc
			desc_lbl.add_theme_font_size_override("font_size", 11)
			desc_lbl.add_theme_color_override("font_color", Color(0.6, 0.6, 0.6, 1))
			desc_lbl.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
			info.add_child(desc_lbl)

		var btn_container = VBoxContainer.new()
		btn_container.alignment = BoxContainer.ALIGNMENT_CENTER
		hbox.add_child(btn_container)

		if completed_levels.has(level["id"]):
			var done_lbl = Label.new()
			done_lbl.text = "COMPLETE"
			done_lbl.add_theme_font_size_override("font_size", 12)
			done_lbl.add_theme_color_override("font_color", Color(0, 1, 0.62, 1))
			btn_container.add_child(done_lbl)

		var play_btn = Button.new()
		play_btn.custom_minimum_size = Vector2(100, 0)
		if unlocked:
			play_btn.text = "PLAY"
			play_btn.add_theme_color_override("font_color", Color(0, 1, 0, 1))
			play_btn.pressed.connect(_on_level_pressed.bind(path))
		else:
			play_btn.text = "LOCKED"
			play_btn.disabled = true
			play_btn.add_theme_color_override("font_color", Color(0.4, 0.4, 0.4, 1))
		btn_container.add_child(play_btn)

		level_list.add_child(card)

func _on_level_pressed(level_path: String):
	OS.set_environment("PHX_LEVEL_PATH", level_path)
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
