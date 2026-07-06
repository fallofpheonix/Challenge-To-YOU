extends Control

var current_screen: String = "telemetry"
var screen_instances: Dictionary = {}
var overlays: Dictionary = {}

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var content_area: Control = $ContentArea
@onready var nav_buttons: VBoxContainer = $Sidebar/NavButtons
@onready var screen_title: Label = $Header/HBoxContainer/TitleLabel
@onready var tick_display: Label = $Header/HBoxContainer/TickDisplay

var nav_items = [
	{"name": "telemetry", "label": "Telemetry", "icon": "📡"},
	{"name": "drones", "label": "Drone Swarm", "icon": "⚙"},
	{"name": "resources", "label": "Resources", "icon": "◆"},
	{"name": "signals", "label": "Pheromones", "icon": "〰"},
	{"name": "structures", "label": "Structures", "icon": "▲"},
	{"name": "hazards", "label": "Hazards", "icon": "⚡"},
	{"name": "alien", "label": "Alien Network", "icon": "◆"},
	{"name": "research", "label": "Research", "icon": "⊞"},
	{"name": "uplink", "label": "Uplink", "icon": "☰"},
	{"name": "replay", "label": "Replay", "icon": "▶"},
]

var screen_paths = {
	"telemetry": "res://ui/screens/telemetry_dashboard.tscn",
	"drones": "res://ui/screens/drone_inspector.tscn",
	"resources": "res://ui/screens/resource_logistics.tscn",
	"signals": "res://ui/screens/pheromone_view.tscn",
	"structures": "res://ui/screens/structure_manager.tscn",
	"hazards": "res://ui/screens/hazard_monitor.tscn",
	"alien": "res://ui/screens/alien_detector.tscn",
	"research": "res://ui/screens/research_tree.tscn",
	"uplink": "res://ui/screens/uplink_terminal.tscn",
	"replay": "res://ui/screens/replay_controls.tscn",
}

func _ready() -> void:
	theme_ctrl.apply_panel_style($Sidebar, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)
	theme_ctrl.apply_panel_style($Header, theme_ctrl.colors.secondary_background, theme_ctrl.colors.border)
	_build_nav()
	_create_overlays()
	switch_to("telemetry")

func _build_nav() -> void:
	for child in nav_buttons.get_children():
		child.queue_free()

	for item in nav_items:
		var btn = Button.new()
		btn.text = "%s  %s" % [item.icon, item.label]
		btn.custom_minimum_size = Vector2(0, 32)
		btn.size_flags_horizontal = 3
		btn.add_theme_font_size_override("font_size", 12)
		btn.pressed.connect(func(): switch_to(item.name))
		nav_buttons.add_child(btn)

		var separator = HSeparator.new()
		separator.custom_minimum_size = Vector2(0, 1)
		nav_buttons.add_child(separator)

func _create_overlays() -> void:
	var heatmap_script = preload("res://ui/overlays/heatmap_overlay.gd")
	var pheromone_script = preload("res://ui/overlays/pheromone_overlay.gd")
	var hazard_script = preload("res://ui/overlays/hazard_overlay.gd")
	var alien_script = preload("res://ui/overlays/alien_overlay.gd")

	var overlay_defs = {
		"heatmap": {"script": heatmap_script, "screen": "telemetry"},
		"pheromone": {"script": pheromone_script, "screen": "signals"},
		"hazard": {"script": hazard_script, "screen": "hazards"},
		"alien": {"script": alien_script, "screen": "alien"},
	}

	for key in overlay_defs:
		var def = overlay_defs[key]
		var node = Node2D.new()
		node.name = key.capitalize() + "Overlay"
		node.set_script(def.script)
		node.visible = false
		content_area.add_child(node)
		overlays[key] = node

func switch_to(screen_name: String) -> void:
	if current_screen == screen_name and screen_instances.has(screen_name):
		return

	if screen_instances.has(current_screen):
		var old = screen_instances[current_screen]
		content_area.remove_child(old)
		old.queue_free()
		screen_instances.erase(current_screen)

	var path = screen_paths.get(screen_name)
	if path.is_empty():
		return

	var scene = load(path)
	if scene == null:
		return

	var instance = scene.instantiate()
	instance.anchor_right = 1.0
	instance.anchor_bottom = 1.0
	instance.grow_horizontal = 2
	instance.grow_vertical = 2
	content_area.add_child(instance)
	screen_instances[screen_name] = instance
	current_screen = screen_name

	for item in nav_items:
		var idx = nav_items.find(item)
		var btn = nav_buttons.get_child(idx * 2) if idx * 2 < nav_buttons.get_child_count() else null
		if btn:
			if item.name == screen_name:
				btn.add_theme_color_override("font_color", Color(0, 1, 0.62, 1))
			else:
				btn.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7, 1))

	var item_data = _find_nav_item(screen_name)
	screen_title.text = item_data.get("label", screen_name) if item_data else screen_name

	_update_overlay_visibility(screen_name)

func _find_nav_item(name: String) -> Dictionary:
	for item in nav_items:
		if item.name == name:
			return item
	return {}

func _update_overlay_visibility(screen_name: String) -> void:
	var screen_to_overlay = {
		"telemetry": "heatmap",
		"signals": "pheromone",
		"hazards": "hazard",
		"alien": "alien",
	}
	var active_overlay = screen_to_overlay.get(screen_name, "")

	for key in overlays:
		if key == active_overlay:
			overlays[key].visible = true
		else:
			overlays[key].visible = false

func update_tick(tick: int) -> void:
	tick_display.text = "Tick: %d" % tick

func forward_data(data: Dictionary) -> void:
	if data.has("tick"):
		update_tick(data["tick"])

	for screen_name in screen_instances:
		var inst = screen_instances[screen_name]
		if inst.has_method("load_telemetry"):
			inst.load_telemetry(data)
		if inst.has_method("load_swarm") and data.has("drones"):
			inst.load_swarm({"drones": data["drones"]})
		if inst.has_method("load_resources") and data.has("grid"):
			var resources = []
			for cell in data["grid"]:
				if cell.get("cnt", 0) > 0:
					resources.append({
						"type": "silicate",
						"yield": cell["cnt"],
						"position": {"x": cell["x"], "y": cell["y"]},
						"state": "embedded",
					})
			inst.load_resources({"resources": resources})
		if inst.has_method("load_pheromones") and data.has("grid"):
			var signals = []
			for cell in data["grid"]:
				var home = cell.get("home", 0)
				var res = cell.get("res", 0)
				var alien = cell.get("alien", 0)
				var pos = {"x": cell.get("x", 0), "y": cell.get("y", 0)}
				if home > 0:
					signals.append({"type": "rally_point", "position": pos, "intensity": home})
				if res > 0:
					signals.append({"type": "resource_found", "position": pos, "intensity": res})
				if alien > 0:
					signals.append({"type": "spoofed_data", "position": pos, "intensity": alien})
			inst.load_pheromones({"signals": signals})
		if inst.has_method("load_grid") and data.has("grid"):
			inst.load_grid({"grid": data["grid"]})
		if inst.has_method("load_hazards") and data.has("hazards"):
			inst.load_hazards({"hazards": data["hazards"]})
		if inst.has_method("load_structures") and data.has("structures"):
			inst.load_structures({"structures": data["structures"]})
		if inst.has_method("load_alien_nodes") and data.has("aliens"):
			var nodes = []
			for alien in data["aliens"]:
				nodes.append({
					"type": alien.get("type", 0),
					"position": {"x": alien.get("x", 0), "y": alien.get("y", 0)},
					"corruption_radius": alien.get("rad", 0),
					"state": "broadcasting",
				})
			inst.load_alien_nodes({"nodes": nodes})
		if inst.has_method("load_research") and data.has("research"):
			inst.load_research({"research": data["research"]})
		if inst.has_method("load_uplink") and data.has("uplink"):
			inst.load_uplink({"uplink": data["uplink"]})
		if inst.has_method("load_replay") and data.has("replay"):
			inst.load_replay(data["replay"])
		if inst.has_method("load_trace") and data.has("trace"):
			inst.load_trace(data["trace"])
		if inst.has_method("load_events") and data.has("events"):
			var events = data["events"]
			if events is Array:
				inst.load_events(events)

	_forward_to_overlays(data)

func _forward_to_overlays(data: Dictionary) -> void:
	if overlays.has("heatmap") and data.has("grid"):
		var density_grid = {}
		for cell in data["grid"]:
			if cell.get("cnt", 0) > 0:
				density_grid["%d,%d" % [cell["x"], cell["y"]]] = cell["cnt"]
		overlays["heatmap"].set_density_data({"density_grid": density_grid})

	if overlays.has("pheromone") and data.has("grid"):
		var overlay_signals = []
		for cell in data["grid"]:
			var home = cell.get("home", 0)
			var res = cell.get("res", 0)
			var alien = cell.get("alien", 0)
			var pos = {"x": cell.get("x", 0), "y": cell.get("y", 0)}
			if home > 0:
				overlay_signals.append({"type": "rally_point", "position": pos, "intensity": home})
			if res > 0:
				overlay_signals.append({"type": "resource_found", "position": pos, "intensity": res})
			if alien > 0:
				overlay_signals.append({"type": "spoofed_data", "position": pos, "intensity": alien})
		overlays["pheromone"].set_grid_data({"signals": overlay_signals})

	if overlays.has("hazard") and data.has("hazards"):
		overlays["hazard"].set_hazard_data({"hazards": data["hazards"]})

	if overlays.has("alien") and data.has("aliens"):
		var nodes = []
		for alien in data["aliens"]:
			nodes.append({
				"position": {"x": alien.get("x", 0), "y": alien.get("y", 0)},
				"corruption_radius": alien.get("rad", 0),
				"state": "broadcasting",
			})
		overlays["alien"].set_alien_data({"nodes": nodes})
