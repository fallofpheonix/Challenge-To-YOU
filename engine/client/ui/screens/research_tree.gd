extends Control

signal research_selected(research_id: String)

var research_data: Dictionary = {}
var research_tiers = [
	{
		"id": "tier1",
		"name": "Tier I: Basic Protocols",
		"nodes": [
			{"id": "while_loops", "name": "WHILE Loops", "cost": "100 Isotope", "desc": "Enable iterative behavior"},
			{"id": "conditional_logic", "name": "Conditional Logic", "cost": "50 Isotope", "desc": "IF/ELSE branching"},
			{"id": "memory_cells", "name": "Memory Cells", "cost": "75 Isotope", "desc": "Store per-drone state"},
		]
	},
	{
		"id": "tier2",
		"name": "Tier II: Coordination",
		"nodes": [
			{"id": "signal_signing", "name": "Crypto Signing", "cost": "300 Isotope", "desc": "Authenticate pheromone sources"},
			{"id": "quorum_consensus", "name": "Quorum Consensus", "cost": "250 Isotope", "desc": "Fault-tolerant decisions"},
			{"id": "swarm_routing", "name": "Swarm Routing", "cost": "200 Isotope", "desc": "Traffic-aware pathfinding"},
		]
	},
	{
		"id": "tier3",
		"name": "Tier III: Advanced Logic",
		"nodes": [
			{"id": "recursion", "name": "Recursion", "cost": "500 Isotope", "desc": "Self-referential protocols"},
			{"id": "self_modify", "name": "Self-Modifying Code", "cost": "800 Isotope", "desc": "Runtime protocol adaptation"},
			{"id": "byzantine_fault_tolerance", "name": "Byzantine Fault Tolerance", "cost": "600 Isotope", "desc": "Resist alien corruption"},
		]
	},
]

@onready var theme_ctrl = get_node("/root/ChrysalisTheme")
@onready var tree_container: VBoxContainer = $ScrollContainer/TreeContainer
@onready var description_label: RichTextLabel = $DescriptionPanel/MarginContainer/DescriptionLabel

func _ready() -> void:
	theme_ctrl.apply_panel_style($DescriptionPanel, theme_ctrl.colors.surface, theme_ctrl.colors.border)
	_populate_tree()

func load_research(data: Dictionary) -> void:
	research_data = data

func _populate_tree() -> void:
	for child in tree_container.get_children():
		child.queue_free()

	for tier in research_tiers:
		var section = preload("res://ui/components/section_header.tscn").instantiate()
		section.set_title(tier.name)
		tree_container.add_child(section)

		for node in tier.nodes:
			var researched = research_data.get("completed", []).has(node.id)
			var card = preload("res://ui/components/chrysalis_card.tscn").instantiate()
			card.header_text = node.name
			card.body_text = "[b]Cost:[/b] %s\n%s" % [node.cost, node.desc]

			if researched:
				card.accent_color = Color(0, 1, 0.62, 1)
				var done_label = Label.new()
				done_label.text = "✓ RESEARCHED"
				done_label.add_theme_color_override("font_color", Color(0, 1, 0.62, 1))
				card.add_child(done_label)
			else:
				card.accent_color = Color(0.3, 0.3, 0.6, 1)
				var research_btn = Button.new()
				research_btn.text = "RESEARCH"
				research_btn.pressed.connect(func(): research_selected.emit(node.id))
				card.add_child(research_btn)

			tree_container.add_child(card)

func select_node(node_id: String) -> void:
	description_label.text = "Selected: %s" % node_id

func _on_description_back_pressed() -> void:
	description_label.text = "Select a research node to view details"
